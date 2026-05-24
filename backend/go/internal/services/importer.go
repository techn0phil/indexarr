package services

import "indexarr/internal/models"

// MovieImporter defines the interface for importing movies
// Implemented by: RadarrImporter, FilesystemMovieScanner
type MovieImporter interface {
	// Import performs a full import/scan of all movies
	Import() (*models.ScanResult, error)

	// ImportMovie imports/refreshes a single movie by its database ID
	ImportMovie(id int64) (*models.ScanResult, error)

	// GetStatus returns the current import/scan status
	GetStatus() (*models.ScanStatus, error)

	// Stop signals the importer to stop
	Stop()

	// IsRunning returns whether an import is currently in progress
	IsRunning() bool
}

// SeriesImporter defines the interface for importing TV series
// Implemented by: SonarrImporter, FilesystemSeriesScanner
type SeriesImporter interface {
	// Import performs a full import/scan of all series
	Import() (*models.ScanResult, error)

	// ImportSeries imports/refreshes a single series by its database ID
	ImportSeries(id int64) (*models.ScanResult, error)

	// GetStatus returns the current import/scan status
	GetStatus() (*models.ScanStatus, error)

	// Stop signals the importer to stop
	Stop()

	// IsRunning returns whether an import is currently in progress
	IsRunning() bool
}
