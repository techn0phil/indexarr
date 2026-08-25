package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"indexarr/internal/api"
	"indexarr/internal/config"
	"indexarr/internal/repository"
	"indexarr/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	godotenv.Load()

	// Initialize logger
	logger := config.InitLogger()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.InitDB(cfg.DBPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Seed database with mock data (only if empty)
	if err := repository.SeedMockData(db); err != nil {
		logger.Fatal().Err(err).Msg("Failed to seed mock data")
	}

	// Temporary post-migration repair for series total counts.
	if err := services.BackfillSeriesTotalsOnStartup(db, cfg); err != nil {
		logger.Warn().Err(err).Msg("Failed to backfill series total counts")
	}

	// Initialize WebSocket broadcaster
	broadcaster := services.NewBroadcaster()
	go broadcaster.Run()
	logger.Info().Msg("WebSocket broadcaster started")

	// Initialize importers based on configuration
	var movieImporter services.MovieImporter
	var seriesImporter services.SeriesImporter

	// Movies: Radarr OR filesystem
	moviesMode := cfg.GetMoviesImportMode()
	moviesModeDetails := ""
	switch moviesMode {
	case "radarr":
		radarrClient := services.NewRadarrClient(cfg.RadarrURL, cfg.RadarrAPIKey)

		// Test connection to Radarr
		if err := radarrClient.TestConnection(); err != nil {
			moviesModeDetails = "[❌ Connection failed]"
		} else {
			moviesModeDetails = "[✅ Connected]"
		}

		movieImporter = services.NewRadarrImporter(db, cfg, radarrClient, broadcaster)

	case "filesystem":
		moviesModeDetails = fmt.Sprintf("%v", cfg.GetMovieLibraryPaths())
		movieImporter = services.NewFilesystemMovieScanner(db, cfg, broadcaster)

	default:
		moviesModeDetails = "[⚠️  No Radarr config or library paths]"
	}

	logger.Info().Str("mode", moviesMode).Str("details", moviesModeDetails).Msg("🎬 Movies importer configured")

	// Series: Sonarr OR filesystem
	seriesMode := cfg.GetSeriesImportMode()
	seriesModeDetails := ""
	switch seriesMode {
	case "sonarr":
		sonarrClient := services.NewSonarrClient(cfg.SonarrURL, cfg.SonarrAPIKey)

		// Test connection to Sonarr
		if err := sonarrClient.TestConnection(); err != nil {
			seriesModeDetails = "[❌ Connection failed]"
		} else {
			seriesModeDetails = "[✅ Connected]"
		}

		seriesImporter = services.NewSonarrImporter(db, cfg, sonarrClient, broadcaster)

	case "filesystem":
		seriesModeDetails = fmt.Sprintf("%v", cfg.GetSeriesLibraryPaths())
		seriesImporter = services.NewFilesystemSeriesScanner(db, cfg, broadcaster)

	default:
		seriesModeDetails = "[⚠️  No Sonarr config or library paths]"
	}

	logger.Info().Str("mode", seriesMode).Str("details", seriesModeDetails).Msg("📺 Series importer configured")

	if cfg.RadarrURL != "" {
		logger.Info().Str("url", cfg.RadarrURL).Msg("📡 Radarr defined")
	}

	if cfg.SonarrURL != "" {
		logger.Info().Str("url", cfg.SonarrURL).Msg("🔊 Sonarr defined")
	}

	if len(cfg.MediaLibraryPaths) > 0 {
		logger.Warn().Strs("paths", cfg.MediaLibraryPaths).Msg("⚠️  MEDIA_LIBRARY_PATHS is deprecated, use MOVIES_LIBRARY_PATHS and SERIES_LIBRARY_PATHS instead")
	}

	// Initialize scheduler with both importers
	scheduler := services.NewScheduler(db, movieImporter, seriesImporter, broadcaster, cfg.ScanInterval)

	if movieImporter != nil || seriesImporter != nil {
		if cfg.ScanInterval > 0 {
			scheduler.Start()
			logger.Info().Int("interval_hours", cfg.ScanInterval).Msg("⏱️  Scheduler started")
		}
	} else {
		logger.Warn().Msg("⚠️  No importers configured, scanning disabled")
	}

	// Initialize user repository for database-backed users
	userRepo := repository.NewUserRepository(db)

	// Initialize authentication service
	authService := services.NewAuthService(cfg, userRepo)
	if cfg.HasAuthEnabled() {
		logger.Info().Str("mode", cfg.AuthMode).Msg("🔐 Authentication enabled")
		if cfg.IsSimpleAuth() && cfg.AuthAdminUsername != "" {
			logger.Info().Str("username", cfg.AuthAdminUsername).Msg("👤 Admin user configured")
		}
	} else {
		logger.Info().Msg("🔓 Authentication disabled")
	}

	// Setup API router
	router := api.SetupRoutes(db, cfg, scheduler, broadcaster, authService)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info().Msg("Shutting down...")
		if scheduler != nil {
			scheduler.Stop()
		}
		os.Exit(0)
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	logger.Info().Str("url", fmt.Sprintf("http://localhost:%s", cfg.ServerPort)).Msg("Indexarr server running")

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal().Err(err).Msg("Server error")
	}
}
