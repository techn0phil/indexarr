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

	// Initialize scheduler based on configuration mode
	var scheduler *services.Scheduler

	if cfg.HasRadarrConfig() {
		// Radarr import mode
		log.Println("🎬 Radarr mode: Using Radarr API for movie imports")

		radarrClient := services.NewRadarrClient(cfg.RadarrURL, cfg.RadarrAPIKey)

		// Test connection to Radarr
		if err := radarrClient.TestConnection(); err != nil {
			log.Printf("⚠️  Warning: Could not connect to Radarr: %v", err)
		} else {
			log.Println("✅ Connected to Radarr successfully")
		}

		radarrImporter := services.NewRadarrImporter(db, cfg, radarrClient, broadcaster)
		scheduler = services.NewSchedulerWithImporter(radarrImporter, cfg.ScanInterval)

		if cfg.ScanInterval > 0 {
			scheduler.Start()
			log.Printf("⏱️  Scheduler started with %d hour interval (Radarr import)", cfg.ScanInterval)
		}
	} else if len(cfg.MediaLibraryPaths) > 0 {
		// Filesystem scan mode
		log.Println("📁 Filesystem mode: Scanning media library paths")

		scanner := services.NewScanner(db, cfg, broadcaster)
		scheduler = services.NewScheduler(scanner, cfg.ScanInterval)

		if cfg.ScanInterval > 0 {
			scheduler.Start()
			log.Printf("⏱️  Scheduler started with %d hour interval (filesystem scan)", cfg.ScanInterval)
		}
	} else {
		log.Println("⚠️  No RADARR_API_KEY or MEDIA_LIBRARY_PATHS configured, scanning disabled")
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
	log.Printf("🎬 Indexarr server running on http://localhost:%s", cfg.ServerPort)
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
