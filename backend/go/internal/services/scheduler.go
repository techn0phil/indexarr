package services

import (
	"log"
	"sync"
	"time"

	"indexarr/internal/models"
)

// Scheduler handles periodic background scanning
type Scheduler struct {
	scanner  *Scanner
	interval time.Duration
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(scanner *Scanner, intervalHours int) *Scheduler {
	return &Scheduler{
		scanner:  scanner,
		interval: time.Duration(intervalHours) * time.Hour,
		stopChan: make(chan struct{}),
	}
}

// Start begins the scheduled scanning
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
	log.Printf("Scheduler started with interval: %v", s.interval)
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
	// Run initial scan after a short delay
	initialDelay := 30 * time.Second
	log.Printf("Scheduler: First scan in %v", initialDelay)

	select {
	case <-time.After(initialDelay):
		s.runScan()
	case <-s.stopChan:
		return
	}

	// Then run at regular intervals
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runScan()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) runScan() {
	log.Println("Scheduler: Starting scheduled scan")
	start := time.Now()

	result, err := s.scanner.Scan()
	if err != nil {
		log.Printf("Scheduler: Scan failed: %v", err)
		return
	}

	duration := time.Since(start)
	log.Printf("Scheduler: Scan completed in %v - %d files, %d movies, %d episodes",
		duration.Round(time.Second), result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded)
}

// TriggerScan manually triggers a scan (used by API)
func (s *Scheduler) TriggerScan() (*models.ScanResult, error) {
	return s.scanner.Scan()
}

// GetScanStatus returns current scan status
func (s *Scheduler) GetScanStatus() (*models.ScanStatus, error) {
	return s.scanner.GetStatus()
}

// StopCurrentScan stops any running scan
func (s *Scheduler) StopCurrentScan() {
	s.scanner.Stop()
}
