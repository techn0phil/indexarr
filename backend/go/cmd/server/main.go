package main

import (
	"fmt"
	"log"
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

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Seed database with mock data (only if empty)
	if err := repository.SeedMockData(db); err != nil {
		log.Fatalf("Failed to seed mock data: %v", err)
	}

	// Initialize WebSocket broadcaster
	broadcaster := services.NewBroadcaster()
	go broadcaster.Run()
	// log.Println("WebSocket broadcaster started")

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

	log.Printf("🎬 Movies import mode: %s %s", moviesMode, moviesModeDetails)

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

	log.Printf("📺 Series import mode: %s %s", seriesMode, seriesModeDetails)

	if cfg.RadarrURL != "" {
		log.Printf("📡 Radarr URL: %s", cfg.RadarrURL)
	}

	if cfg.SonarrURL != "" {
		log.Printf("🔊 Sonarr URL: %s", cfg.SonarrURL)
	}

	if len(cfg.MediaLibraryPaths) > 0 {
		log.Printf("📂 Library paths: %v", cfg.MediaLibraryPaths)
	}

	// Initialize user repository for database-backed users
	userRepo := repository.NewUserRepository(db)

	// Initialize authentication service
	authService := services.NewAuthService(cfg, userRepo)
	if cfg.HasAuthEnabled() {
		log.Printf("🔐 Authentication mode: %s", cfg.AuthMode)
		if cfg.IsLocalAuth() && cfg.AuthAdminUsername != "" {
			log.Printf("👤 Admin user: %s (env)", cfg.AuthAdminUsername)
		}
	} else {
		log.Println("🔓 Authentication disabled")
	}

	// Initialize OIDC service if configured
	var oidcService *services.OIDCService
	if cfg.IsOIDCAuth() && cfg.HasOIDCConfig() {
		var err error
		oidcService, err = services.NewOIDCService(cfg, userRepo)
		if err != nil {
			log.Printf("⚠️  Failed to initialize OIDC service: %v", err)
		} else {
			log.Printf("🔑 OIDC provider: %s", cfg.OIDCIssuerURL)
		}
	}

	// Initialize scheduler with both importers
	scheduler := services.NewScheduler(db, movieImporter, seriesImporter, broadcaster, cfg.ScanInterval)

	if movieImporter != nil || seriesImporter != nil {
		if cfg.ScanInterval > 0 {
			scheduler.Start()
			log.Printf("⏱️  Scheduler started with %d hour interval", cfg.ScanInterval)
		}
	} else {
		log.Println("⚠️  No importers configured, scanning disabled")
	}

	// Setup API router
	router := api.SetupRoutes(db, cfg, scheduler, broadcaster, authService, oidcService)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")
		if scheduler != nil {
			scheduler.Stop()
		}
		os.Exit(0)
	}()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("Indexarr server running on http://localhost:%s", cfg.ServerPort)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
