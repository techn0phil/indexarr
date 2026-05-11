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

	// Initialize scanner and scheduler
	var scheduler *services.Scheduler
	if len(cfg.MediaLibraryPaths) > 0 {
		scanner := services.NewScanner(db, cfg)
		scheduler = services.NewScheduler(scanner, cfg.ScanInterval)

		// Start scheduler if interval is configured
		if cfg.ScanInterval > 0 {
			scheduler.Start()
			log.Printf("Scheduler started with %d hour interval", cfg.ScanInterval)
		}
	} else {
		log.Println("⚠️  No MEDIA_LIBRARY_PATHS configured, scanning disabled")
	}

	// Setup API router
	router := api.SetupRoutes(db, cfg, scheduler)

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
