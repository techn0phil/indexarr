package services

import "indexarr/internal/models"

// MovieImporter defines the interface for importing movies
// Implemented by: RadarrImporter, FilesystemMovieScanner
type MovieImporter interface {
	// Import performs a full import/scan of all movies
	// If ctx is nil, uses default behavior (broadcasts start/complete).
	// If ctx is provided, applies offset and may suppress broadcasts.
	Import(ctx *models.ProgressContext) (*models.ScanResult, error)

	// ImportMovie imports/refreshes a single movie by its database ID
	ImportMovie(id int64) (*models.ScanResult, error)

	// GetPendingFileCount returns the number of files to be processed without starting import.
	// For API-based importers, this fetches and caches data for subsequent Import() call.
	GetPendingFileCount() (int, error)

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
	// If ctx is nil, uses default behavior (broadcasts start/complete).
	// If ctx is provided, applies offset and may suppress broadcasts.
	Import(ctx *models.ProgressContext) (*models.ScanResult, error)

	// ImportSeries imports/refreshes a single series by its database ID
	ImportSeries(id int64) (*models.ScanResult, error)

	// GetPendingFileCount returns the number of episode files to be processed without starting import.
	// For API-based importers, this fetches and caches data for subsequent Import() call.
	GetPendingFileCount() (int, error)

	// GetStatus returns the current import/scan status
	GetStatus() (*models.ScanStatus, error)

	// Stop signals the importer to stop
	Stop()

	// IsRunning returns whether an import is currently in progress
	IsRunning() bool
}
