package services

import (
	"log"
	"sync"
	"time"

	"indexarr/internal/models"
)

// Scheduler handles periodic background scanning or importing
type Scheduler struct {
	scanner  *Scanner        // Filesystem scanner (optional)
	importer *RadarrImporter // Radarr importer (optional)
	mode     string          // "radarr", "filesystem", or "disabled"
	interval time.Duration
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewScheduler creates a new scheduler with filesystem scanner
func NewScheduler(scanner *Scanner, intervalHours int) *Scheduler {
	return &Scheduler{
		scanner:  scanner,
		mode:     "filesystem",
		interval: time.Duration(intervalHours) * time.Hour,
		stopChan: make(chan struct{}),
	}
}

// NewSchedulerWithImporter creates a new scheduler with Radarr importer
func NewSchedulerWithImporter(importer *RadarrImporter, intervalHours int) *Scheduler {
	return &Scheduler{
		importer: importer,
		mode:     "radarr",
		interval: time.Duration(intervalHours) * time.Hour,
		stopChan: make(chan struct{}),
	}
}

// GetMode returns the current scheduler mode
func (s *Scheduler) GetMode() string {
	return s.mode
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
	log.Printf("Scheduler: First %s in %v", s.mode, initialDelay)

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
	log.Printf("Scheduler: Run scheduled %s", s.mode)

	var err error
	if s.importer != nil {
		_, err = s.importer.Import()
	} else if s.scanner != nil {
		_, err = s.scanner.Scan()
	} else {
		log.Println("Scheduler: No scanner or importer configured")
		return
	}

	if err != nil {
		log.Printf("Scheduler: %s failed: %v", s.mode, err)
		return
	}

	log.Printf("Scheduler: Scheduled %s completed", s.mode)
}

// TriggerScan manually triggers a scan or import (used by API)
func (s *Scheduler) TriggerScan() (*models.ScanResult, error) {
	if s.importer != nil {
		return s.importer.Import()
	}
	if s.scanner != nil {
		return s.scanner.Scan()
	}
	return nil, nil
}

// GetScanStatus returns current scan status
func (s *Scheduler) GetScanStatus() (*models.ScanStatus, error) {
	if s.importer != nil {
		return s.importer.GetStatus()
	}
	if s.scanner != nil {
		return s.scanner.GetStatus()
	}
	return &models.ScanStatus{Status: "disabled"}, nil
}

// StopCurrentScan stops any running scan or import
func (s *Scheduler) StopCurrentScan() {
	if s.importer != nil {
		s.importer.Stop()
	}
	if s.scanner != nil {
		s.scanner.Stop()
	}
}

// TriggerMovieScan triggers a scan/refresh for a specific movie
func (s *Scheduler) TriggerMovieScan(id int64) (*models.ScanResult, error) {
	if s.importer != nil {
		return s.importer.ImportMovie(id)
	}
	if s.scanner != nil {
		return s.scanner.ScanMovie(id)
	}
	return nil, nil
}

// TriggerSeriesScan triggers a scan for a specific series to update its metadata
func (s *Scheduler) TriggerSeriesScan(id int64) (*models.ScanResult, error) {
	// Series scanning only supported via filesystem scanner (for now)
	if s.scanner != nil {
		return s.scanner.ScanSeries(id)
	}
	return nil, nil
}
