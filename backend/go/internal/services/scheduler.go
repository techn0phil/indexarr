package services

import (
	"log"
	"sync"
	"time"

	"indexarr/internal/models"
)

// Scheduler handles periodic background scanning or importing for both movies and series
type Scheduler struct {
	movieImporter  MovieImporter  // Radarr OR FilesystemMovieScanner (optional)
	seriesImporter SeriesImporter // Sonarr OR FilesystemSeriesScanner (optional)
	interval       time.Duration
	stopChan       chan struct{}
	running        bool
	mu             sync.Mutex
}

// NewScheduler creates a new scheduler with the given importers
// Either or both importers can be nil if that media type is disabled
func NewScheduler(movieImporter MovieImporter, seriesImporter SeriesImporter, intervalHours int) *Scheduler {
	return &Scheduler{
		movieImporter:  movieImporter,
		seriesImporter: seriesImporter,
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

	// Import movies first, then series (sequential to avoid CPU overload)
	if s.movieImporter != nil {
		log.Println("Scheduler: Starting movie import...")
		if _, err := s.movieImporter.Import(); err != nil {
			log.Printf("Scheduler: Movie import failed: %v", err)
		} else {
			log.Println("Scheduler: Movie import completed")
		}
	}

	if s.seriesImporter != nil {
		log.Println("Scheduler: Starting series import...")
		if _, err := s.seriesImporter.Import(); err != nil {
			log.Printf("Scheduler: Series import failed: %v", err)
		} else {
			log.Println("Scheduler: Series import completed")
		}
	}

	log.Println("Scheduler: Scheduled import completed")
}

// TriggerScan manually triggers an import for both movies and series (used by API)
func (s *Scheduler) TriggerScan() (*models.ScanResult, error) {
	result := &models.ScanResult{
		Errors: []string{},
	}

	// Import movies
	if s.movieImporter != nil {
		movieResult, err := s.movieImporter.Import()
		if err != nil {
			result.Errors = append(result.Errors, "Movie import: "+err.Error())
		} else if movieResult != nil {
			result.FilesFound += movieResult.FilesFound
			result.FilesProcessed += movieResult.FilesProcessed
			result.MoviesAdded += movieResult.MoviesAdded
			result.Errors = append(result.Errors, movieResult.Errors...)
		}
	}

	// Import series
	if s.seriesImporter != nil {
		seriesResult, err := s.seriesImporter.Import()
		if err != nil {
			result.Errors = append(result.Errors, "Series import: "+err.Error())
		} else if seriesResult != nil {
			result.FilesFound += seriesResult.FilesFound
			result.FilesProcessed += seriesResult.FilesProcessed
			result.EpisodesAdded += seriesResult.EpisodesAdded
			result.Errors = append(result.Errors, seriesResult.Errors...)
		}
	}

	return result, nil
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
	if s.movieImporter != nil {
		s.movieImporter.Stop()
	}
	if s.seriesImporter != nil {
		s.seriesImporter.Stop()
	}
}

// TriggerMovieScan manually triggers a scan/refresh for a specific movie
func (s *Scheduler) TriggerMovieScan(id int64) (*models.ScanResult, error) {
	if s.movieImporter != nil {
		return s.movieImporter.ImportMovie(id)
	}
	return nil, nil
}

// TriggerSeriesScan triggers a scan for a specific series to update its metadata
func (s *Scheduler) TriggerSeriesScan(id int64) (*models.ScanResult, error) {
	if s.seriesImporter != nil {
		return s.seriesImporter.ImportSeries(id)
	}
	return nil, nil
}
