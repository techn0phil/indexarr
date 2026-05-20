package services

import (
	"database/sql"

	"indexarr/internal/config"
	"indexarr/internal/models"
)

// FilesystemMovieScanner implements MovieImporter for filesystem-based movie scanning
// It wraps the existing Scanner with movie-specific functionality
type FilesystemMovieScanner struct {
	scanner *Scanner
}

// NewFilesystemMovieScanner creates a new filesystem movie scanner
func NewFilesystemMovieScanner(db *sql.DB, cfg *config.Config, broadcaster *Broadcaster) *FilesystemMovieScanner {
	return &FilesystemMovieScanner{
		scanner: NewScanner(db, cfg, broadcaster),
	}
}

// Import performs a full scan for movies
func (fms *FilesystemMovieScanner) Import() (*models.ScanResult, error) {
	// The scanner will scan all media files; we filter to movies
	// For now, use the same scan method - it handles both
	return fms.scanner.ScanPaths(fms.scanner.config.GetMovieLibraryPaths())
}

// ImportMovie scans/refreshes a single movie by its database ID
func (fms *FilesystemMovieScanner) ImportMovie(id int64) (*models.ScanResult, error) {
	return fms.scanner.ScanMovie(id)
}

// GetStatus returns the current scan status
func (fms *FilesystemMovieScanner) GetStatus() (*models.ScanStatus, error) {
	return fms.scanner.GetStatus()
}

// Stop signals the scanner to stop
func (fms *FilesystemMovieScanner) Stop() {
	fms.scanner.Stop()
}

// IsRunning returns whether a scan is currently in progress
func (fms *FilesystemMovieScanner) IsRunning() bool {
	return fms.scanner.IsRunning()
}

// FilesystemSeriesScanner implements SeriesImporter for filesystem-based series scanning
// It wraps the existing Scanner with series-specific functionality
type FilesystemSeriesScanner struct {
	scanner *Scanner
}

// NewFilesystemSeriesScanner creates a new filesystem series scanner
func NewFilesystemSeriesScanner(db *sql.DB, cfg *config.Config, broadcaster *Broadcaster) *FilesystemSeriesScanner {
	return &FilesystemSeriesScanner{
		scanner: NewScanner(db, cfg, broadcaster),
	}
}

// Import performs a full scan for series
func (fss *FilesystemSeriesScanner) Import() (*models.ScanResult, error) {
	// The scanner will scan all media files; we filter to series
	// For now, use the same scan method - it handles both
	return fss.scanner.ScanPaths(fss.scanner.config.GetSeriesLibraryPaths())
}

// ImportSeries scans/refreshes a single series by its database ID
func (fss *FilesystemSeriesScanner) ImportSeries(id int64) (*models.ScanResult, error) {
	return fss.scanner.ScanSeries(id)
}

// GetStatus returns the current scan status
func (fss *FilesystemSeriesScanner) GetStatus() (*models.ScanStatus, error) {
	return fss.scanner.GetStatus()
}

// Stop signals the scanner to stop
func (fss *FilesystemSeriesScanner) Stop() {
	fss.scanner.Stop()
}

// IsRunning returns whether a scan is currently in progress
func (fss *FilesystemSeriesScanner) IsRunning() bool {
	return fss.scanner.IsRunning()
}
