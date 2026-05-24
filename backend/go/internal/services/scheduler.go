package services

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"indexarr/internal/models"
	"indexarr/internal/repository"
)

// Scheduler handles periodic background scanning or importing for both movies and series
type Scheduler struct {
	db             *sql.DB
	movieImporter  MovieImporter  // Radarr OR FilesystemMovieScanner (optional)
	seriesImporter SeriesImporter // Sonarr OR FilesystemSeriesScanner (optional)
	broadcaster    *Broadcaster
	interval       time.Duration
	stopChan       chan struct{}
	running        bool
	stopRequested  bool // Flag to abort between movie and series imports
	mu             sync.Mutex
}

// NewScheduler creates a new scheduler with the given importers
// Either or both importers can be nil if that media type is disabled
func NewScheduler(db *sql.DB, movieImporter MovieImporter, seriesImporter SeriesImporter, broadcaster *Broadcaster, intervalHours int) *Scheduler {
	return &Scheduler{
		db:             db,
		movieImporter:  movieImporter,
		seriesImporter: seriesImporter,
		broadcaster:    broadcaster,
		interval:       time.Duration(intervalHours) * time.Hour,
		stopChan:       make(chan struct{}),
	}
}

// GetMode returns a description of the current scheduler mode (for backward compatibility)
func (s *Scheduler) GetMode() string {
	if s.movieImporter != nil && s.seriesImporter != nil {
		return "dual"
	}
	if s.movieImporter != nil {
		return "movies"
	}
	if s.seriesImporter != nil {
		return "series"
	}
	return "disabled"
}

// Start begins the scheduled scanning/importing
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go s.run()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
	log.Println("Scheduler stopped")
}

// IsRunning returns whether the scheduler is active
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) run() {
	// Run initial scan/import after a short delay
	initialDelay := 30 * time.Second
	log.Printf("Scheduler: First import in %v (mode: %s)", initialDelay, s.GetMode())

	select {
	case <-time.After(initialDelay):
		s.runImport()
	case <-s.stopChan:
		return
	}

	// Then run at regular intervals
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runImport()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) runImport() {
	log.Println("Scheduler: Running scheduled import")
	start := time.Now()

	// Reset stop flag
	s.mu.Lock()
	s.stopRequested = false
	s.mu.Unlock()

	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanStart(0, time.Now().Format(time.RFC3339))
	}

	// Phase 1: Pre-fetch counts from both importers
	var movieCount, seriesCount int
	var err error

	if s.movieImporter != nil {
		movieCount, err = s.movieImporter.GetPendingFileCount()
		if err != nil {
			log.Printf("Scheduler: Failed to get movie count: %v", err)
			movieCount = 0
		}
	}

	if s.seriesImporter != nil {
		seriesCount, err = s.seriesImporter.GetPendingFileCount()
		if err != nil {
			log.Printf("Scheduler: Failed to get series count: %v", err)
			seriesCount = 0
		}
	}

	totalCount := movieCount + seriesCount
	log.Printf("Scheduler: Combined total: %d files (%d movies + %d episodes)", totalCount, movieCount, seriesCount)

	// Phase 2: Broadcast combined scan_start
	startedAt := time.Now().Format(time.RFC3339)
	status := &models.ScanStatus{
		Status:     "running",
		StartedAt:  startedAt,
		FilesFound: totalCount,
	}
	if err := repository.UpdateScanStatus(s.db, status); err != nil {
		log.Printf("Scheduler: Failed to update scan status: %v", err)
	}
	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanStart(totalCount, startedAt)
	}

	var movieProcessed, moviesAdded, episodesAdded int
	var totalProcessed int
	var errors []string

	// Phase 3: Run movie import with progress coordination
	if s.movieImporter != nil && movieCount > 0 {
		log.Println("Scheduler: Starting movie import...")
		moviectx := &models.ProgressContext{
			Offset:                0,
			TotalOverride:         totalCount,
			SuppressStartComplete: true,
		}

		movieResult, err := s.movieImporter.Import(moviectx)
		if err != nil {
			log.Printf("Scheduler: Movie import failed: %v", err)
			errors = append(errors, "Movie import: "+err.Error())
		} else {
			log.Println("Scheduler: Movie import completed")
		}
		if movieResult != nil {
			movieProcessed = movieResult.FilesProcessed
			totalProcessed += movieResult.FilesProcessed
			moviesAdded += movieResult.MoviesAdded
			errors = append(errors, movieResult.Errors...)
		}
	}

	// Check if stop was requested between imports
	s.mu.Lock()
	stopped := s.stopRequested
	s.mu.Unlock()

	if stopped {
		log.Println("Scheduler: Stopped by user after movie import")
		status.Status = "stopped"
		status.CompletedAt = time.Now().Format(time.RFC3339)
		status.FilesProcessed = totalProcessed
		status.ErrorMessage = "Import stopped by user"
		repository.UpdateScanStatus(s.db, status)
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanStopped()
		}

		duration := time.Since(start)
		log.Printf("TriggerScan: Interrupted after %v - %d files, %d movies, %d episodes, %d errors",
			duration.Round(time.Second), totalProcessed, moviesAdded, episodesAdded, len(errors))

		// Print the top 100 errors if there are any
		if len(errors) > 0 {
			// Log the first 100 errors for visibility
			for i, err := range errors {
				if i >= 100 {
					log.Printf("  - ... %d more lines", len(errors)-100)
					break
				}
				log.Printf("  - %s", err)
			}
		}

		return
	}

	// Phase 4: Run series import with progress coordination (offset by movie count)
	if s.seriesImporter != nil && seriesCount > 0 {
		log.Println("Scheduler: Starting series import...")
		seriesctx := &models.ProgressContext{
			Offset:                movieProcessed,
			TotalOverride:         totalCount,
			SuppressStartComplete: true,
		}

		seriesResult, err := s.seriesImporter.Import(seriesctx)
		if err != nil {
			log.Printf("Scheduler: Series import failed: %v", err)
			errors = append(errors, "Series import: "+err.Error())
		} else {
			log.Println("Scheduler: Series import completed")
		}
		if seriesResult != nil {
			totalProcessed += seriesResult.FilesProcessed
			episodesAdded += seriesResult.EpisodesAdded
			errors = append(errors, seriesResult.Errors...)
		}
	}

	// Check if stopped during series import
	s.mu.Lock()
	stopped = s.stopRequested
	s.mu.Unlock()

	if stopped {
		log.Println("Scheduler: Stopped by user during series import")
		status.Status = "stopped"
		status.CompletedAt = time.Now().Format(time.RFC3339)
		status.FilesProcessed = totalProcessed
		status.ErrorMessage = "Import stopped by user"
		repository.UpdateScanStatus(s.db, status)
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanStopped()
		}

		duration := time.Since(start)
		log.Printf("TriggerScan: Interrupted after %v - %d files, %d movies, %d episodes, %d errors",
			duration.Round(time.Second), totalProcessed, moviesAdded, episodesAdded, len(errors))

		// Print the top 100 errors if there are any
		if len(errors) > 0 {
			// Log the first 100 errors for visibility
			for i, err := range errors {
				if i >= 100 {
					log.Printf("  - ... %d more lines", len(errors)-100)
					break
				}
				log.Printf("  - %s", err)
			}
		}

		return
	}

	// Phase 5: Broadcast combined scan_complete
	status.Status = "completed"
	status.CompletedAt = time.Now().Format(time.RFC3339)
	status.FilesProcessed = totalProcessed
	if len(errors) > 0 {
		status.ErrorMessage = fmt.Sprintf("%d errors during import", len(errors))
	}
	repository.UpdateScanStatus(s.db, status)

	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanComplete(totalProcessed, moviesAdded, episodesAdded)
	}

	duration := time.Since(start)
	log.Printf("Scheduler: Scheduled import completed in %v - %d files, %d movies, %d episodes, %d errors",
		duration.Round(time.Second), totalProcessed, moviesAdded, episodesAdded, len(errors))

	// Print the top 100 errors if there are any
	if len(errors) > 0 {
		// Log the first 100 errors for visibility
		for i, err := range errors {
			if i >= 100 {
				log.Printf("  - ... %d more lines", len(errors)-100)
				break
			}
			log.Printf("  - %s", err)
		}
	}
}

// TriggerScan manually triggers an import for both movies and series (used by API)
// Uses coordinated progress reporting for a single 0-100% progress bar across both
func (s *Scheduler) TriggerScan() (*models.ScanResult, error) {
	log.Println("TriggerScan: Starting coordinated import")
	start := time.Now()

	// Reset stop flag
	s.mu.Lock()
	s.stopRequested = false
	s.mu.Unlock()

	result := &models.ScanResult{
		Errors: []string{},
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanStart(0, time.Now().Format(time.RFC3339))
	}

	// Phase 1: Pre-fetch counts from both importers
	var movieCount, seriesCount int
	var err error

	if s.movieImporter != nil {
		movieCount, err = s.movieImporter.GetPendingFileCount()
		if err != nil {
			log.Printf("TriggerScan: Failed to get movie count: %v", err)
			result.Errors = append(result.Errors, "Movie count: "+err.Error())
			movieCount = 0
		}
	}

	if s.seriesImporter != nil {
		seriesCount, err = s.seriesImporter.GetPendingFileCount()
		if err != nil {
			log.Printf("TriggerScan: Failed to get series count: %v", err)
			result.Errors = append(result.Errors, "Series count: "+err.Error())
			seriesCount = 0
		}
	}

	totalCount := movieCount + seriesCount
	result.FilesFound = totalCount
	log.Printf("TriggerScan: Combined total: %d files (%d movies + %d episodes)", totalCount, movieCount, seriesCount)

	// Phase 2: Broadcast combined scan_start
	startedAt := time.Now().Format(time.RFC3339)
	status := &models.ScanStatus{
		Status:     "running",
		StartedAt:  startedAt,
		FilesFound: totalCount,
	}
	if err := repository.UpdateScanStatus(s.db, status); err != nil {
		log.Printf("TriggerScan: Failed to update scan status: %v", err)
	}
	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanStart(totalCount, startedAt)
	}

	// Phase 3: Run movie import with progress coordination
	var movieProcessed int
	if s.movieImporter != nil && movieCount > 0 {
		moviectx := &models.ProgressContext{
			Offset:                0,
			TotalOverride:         totalCount,
			SuppressStartComplete: true,
		}

		movieResult, err := s.movieImporter.Import(moviectx)
		if err != nil {
			log.Printf("TriggerScan: Movie import error: %v", err)
			result.Errors = append(result.Errors, "Movie import: "+err.Error())
		}
		if movieResult != nil {
			movieProcessed = movieResult.FilesProcessed
			result.FilesProcessed += movieResult.FilesProcessed
			result.MoviesAdded += movieResult.MoviesAdded
			result.Errors = append(result.Errors, movieResult.Errors...)
		}
	}

	// Check if stop was requested between imports
	s.mu.Lock()
	stopped := s.stopRequested
	s.mu.Unlock()

	if stopped {
		log.Println("TriggerScan: Stopped by user after movie import")
		status.Status = "stopped"
		status.CompletedAt = time.Now().Format(time.RFC3339)
		status.FilesProcessed = result.FilesProcessed
		status.ErrorMessage = "Import stopped by user"
		repository.UpdateScanStatus(s.db, status)
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanStopped()
		}

		duration := time.Since(start)
		log.Printf("TriggerScan: Interrupted after %v - %d files, %d movies, %d episodes, %d errors",
			duration.Round(time.Second), result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded, len(result.Errors))

		// Print the top 100 errors if there are any
		if len(result.Errors) > 0 {
			// Log the first 100 errors for visibility
			for i, err := range result.Errors {
				if i >= 100 {
					log.Printf("  - ... %d more lines", len(result.Errors)-100)
					break
				}
				log.Printf("  - %s", err)
			}
		}

		return result, fmt.Errorf("import stopped by user")
	}

	// Phase 4: Run series import with progress coordination (offset by movie count)
	if s.seriesImporter != nil && seriesCount > 0 {
		seriesctx := &models.ProgressContext{
			Offset:                movieProcessed,
			TotalOverride:         totalCount,
			SuppressStartComplete: true,
		}

		seriesResult, err := s.seriesImporter.Import(seriesctx)
		if err != nil {
			log.Printf("TriggerScan: Series import error: %v", err)
			result.Errors = append(result.Errors, "Series import: "+err.Error())
		}
		if seriesResult != nil {
			result.FilesProcessed += seriesResult.FilesProcessed
			result.EpisodesAdded += seriesResult.EpisodesAdded
			result.Errors = append(result.Errors, seriesResult.Errors...)
		}
	}

	// Check if stopped during series import
	s.mu.Lock()
	stopped = s.stopRequested
	s.mu.Unlock()

	if stopped {
		log.Println("TriggerScan: Stopped by user during series import")
		status.Status = "stopped"
		status.CompletedAt = time.Now().Format(time.RFC3339)
		status.FilesProcessed = result.FilesProcessed
		status.ErrorMessage = "Import stopped by user"
		repository.UpdateScanStatus(s.db, status)
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanStopped()
		}

		duration := time.Since(start)
		log.Printf("TriggerScan: Interrupted after %v - %d files, %d movies, %d episodes, %d errors",
			duration.Round(time.Second), result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded, len(result.Errors))

		// Print the top 100 errors if there are any
		if len(result.Errors) > 0 {
			// Log the first 100 errors for visibility
			for i, err := range result.Errors {
				if i >= 100 {
					log.Printf("  - ... %d more lines", len(result.Errors)-100)
					break
				}
				log.Printf("  - %s", err)
			}
		}

		return result, fmt.Errorf("import stopped by user")
	}

	// Phase 5: Broadcast combined scan_complete
	status.Status = "completed"
	status.CompletedAt = time.Now().Format(time.RFC3339)
	status.FilesProcessed = result.FilesProcessed
	if len(result.Errors) > 0 {
		status.ErrorMessage = fmt.Sprintf("%d errors during import", len(result.Errors))
	}
	repository.UpdateScanStatus(s.db, status)

	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanComplete(result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded)
	}

	duration := time.Since(start)
	log.Printf("TriggerScan: Completed in %v - %d files, %d movies, %d episodes, %d errors",
		duration.Round(time.Second), result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded, len(result.Errors))

	// Print the top 100 errors if there are any
	if len(result.Errors) > 0 {
		// Log the first 100 errors for visibility
		for i, err := range result.Errors {
			if i >= 100 {
				log.Printf("  - ... %d more lines", len(result.Errors)-100)
				break
			}
			log.Printf("  - %s", err)
		}
	}

	return result, nil
}

// TriggerMoviesScan triggers a scan for movies only
func (s *Scheduler) TriggerMoviesScan() (*models.ScanResult, error) {
	if s.movieImporter == nil {
		return nil, nil
	}
	return s.movieImporter.Import(nil)
}

// TriggerSeriesScan triggers a scan for series only
func (s *Scheduler) TriggerSeriesScan() (*models.ScanResult, error) {
	if s.seriesImporter == nil {
		return nil, nil
	}
	return s.seriesImporter.Import(nil)
}

// GetScanStatus returns current scan status (combines movie and series status)
func (s *Scheduler) GetScanStatus() (*models.ScanStatus, error) {
	// Check movie importer status first
	if s.movieImporter != nil {
		status, err := s.movieImporter.GetStatus()
		if err != nil {
			return nil, err
		}
		// If movie import is running, return that status
		if status.Status == "running" {
			return status, nil
		}
	}

	// Check series importer status
	if s.seriesImporter != nil {
		status, err := s.seriesImporter.GetStatus()
		if err != nil {
			return nil, err
		}
		return status, nil
	}

	// No importers configured
	if s.movieImporter != nil {
		return s.movieImporter.GetStatus()
	}

	return &models.ScanStatus{Status: "disabled"}, nil
}

// StopCurrentScan stops any running scan or import
func (s *Scheduler) StopCurrentScan() {
	// Set flag to prevent series import from starting if movies are still running
	s.mu.Lock()
	s.stopRequested = true
	s.mu.Unlock()

	// Stop both importers
	if s.movieImporter != nil {
		s.movieImporter.Stop()
	}
	if s.seriesImporter != nil {
		s.seriesImporter.Stop()
	}
}

// TriggerSingleMovieScan manually triggers a scan/refresh for a specific movie
func (s *Scheduler) TriggerSingleMovieScan(id int64) (*models.ScanResult, error) {
	if s.movieImporter != nil {
		return s.movieImporter.ImportMovie(id)
	}
	return nil, nil
}

// TriggerSingleSeriesScan triggers a scan for a specific series to update its metadata
func (s *Scheduler) TriggerSingleSeriesScan(id int64) (*models.ScanResult, error) {
	if s.seriesImporter != nil {
		return s.seriesImporter.ImportSeries(id)
	}
	return nil, nil
}
