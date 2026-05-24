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
	log.Println("WebSocket broadcaster started")

	// Initialize importers based on configuration
	var movieImporter services.MovieImporter
	var seriesImporter services.SeriesImporter

	// Movies: Radarr OR filesystem
	moviesMode := cfg.GetMoviesImportMode()
	switch moviesMode {
	case "radarr":
		log.Println("🎬 Movies: Using Radarr API")
		radarrClient := services.NewRadarrClient(cfg.RadarrURL, cfg.RadarrAPIKey)

		// Test connection to Radarr
		if err := radarrClient.TestConnection(); err != nil {
			log.Printf("⚠️  Warning: Could not connect to Radarr: %v", err)
		} else {
			log.Println("✅ Connected to Radarr successfully")
		}

		movieImporter = services.NewRadarrImporter(db, cfg, radarrClient, broadcaster)

	case "filesystem":
		log.Println("📁 Movies: Using filesystem scanning")
		log.Printf("   Paths: %v", cfg.GetMovieLibraryPaths())
		movieImporter = services.NewFilesystemMovieScanner(db, cfg, broadcaster)

	default:
		log.Println("⚠️  Movies: Import disabled (no Radarr config or library paths)")
	}

	// Series: Sonarr OR filesystem
	seriesMode := cfg.GetSeriesImportMode()
	switch seriesMode {
	case "sonarr":
		log.Println("📺 Series: Using Sonarr API")
		sonarrClient := services.NewSonarrClient(cfg.SonarrURL, cfg.SonarrAPIKey)
		if err := sonarrClient.TestConnection(); err != nil {
			log.Printf("⚠️  Warning: Could not connect to Sonarr: %v", err)
		} else {
			log.Println("✅ Connected to Sonarr successfully")
		}
		seriesImporter = services.NewSonarrImporter(db, cfg, sonarrClient, broadcaster)

	case "filesystem":
		log.Println("📁 Series: Using filesystem scanning")
		log.Printf("   Paths: %v", cfg.GetSeriesLibraryPaths())
		seriesImporter = services.NewFilesystemSeriesScanner(db, cfg, broadcaster)

	default:
		log.Println("⚠️  Series: Import disabled (no Sonarr config or library paths)")
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
	router := api.SetupRoutes(db, cfg, scheduler, broadcaster)

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
	log.Printf("🎬 Movies import mode: %s", moviesMode)
	log.Printf("📺 Series import mode: %s", seriesMode)
	log.Printf("📡 Radarr URL: %s", cfg.RadarrURL)
	log.Printf("🔊 Sonarr URL: %s", cfg.SonarrURL)
	log.Printf("📁 Database: %s", cfg.DBPath)
	if len(cfg.MediaLibraryPaths) > 0 {
		log.Printf("📂 Library paths: %v", cfg.MediaLibraryPaths)
	}

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
