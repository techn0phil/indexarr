package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
)

// RadarrImporter handles importing movies from Radarr API
type RadarrImporter struct {
	db          *sql.DB
	config      *config.Config
	client      *RadarrClient
	extractor   *Extractor
	broadcaster *Broadcaster
	running     bool
	stopChan    chan struct{}
	mu          sync.Mutex
}

// NewRadarrImporter creates a new Radarr importer
func NewRadarrImporter(db *sql.DB, cfg *config.Config, client *RadarrClient, broadcaster *Broadcaster) *RadarrImporter {
	return &RadarrImporter{
		db:          db,
		config:      cfg,
		client:      client,
		extractor:   NewExtractor(cfg.MediainfoPath, cfg.ScanTimeout),
		broadcaster: broadcaster,
		stopChan:    make(chan struct{}),
	}
}

// IsRunning returns whether an import is currently in progress
func (ri *RadarrImporter) IsRunning() bool {
	ri.mu.Lock()
	defer ri.mu.Unlock()
	return ri.running
}

// Stop signals the importer to stop
func (ri *RadarrImporter) Stop() {
	ri.mu.Lock()
	if ri.running {
		close(ri.stopChan)
	}
	ri.mu.Unlock()
}

// Import performs a full sync from Radarr
func (ri *RadarrImporter) Import() (*models.ScanResult, error) {
	ri.mu.Lock()
	if ri.running {
		ri.mu.Unlock()
		return nil, fmt.Errorf("import already in progress")
	}
	ri.running = true
	ri.stopChan = make(chan struct{})
	ri.mu.Unlock()

	log.Println("Starting Radarr import")
	start := time.Now()

	defer func() {
		ri.mu.Lock()
		ri.running = false
		select {
		case <-ri.stopChan:
			// Already closed
		default:
			close(ri.stopChan)
		}
		ri.mu.Unlock()
	}()

	result := &models.ScanResult{
		Errors: []string{},
	}

	// Update scan status to running
	status := &models.ScanStatus{
		Status:     "running",
		StartedAt:  time.Now().Format(time.RFC3339),
		FilesFound: 0,
	}
	if err := repository.UpdateScanStatus(ri.db, status); err != nil {
		log.Printf("Failed to update scan status: %v", err)
	}

	// Fetch all movies from Radarr
	log.Println("Fetching movies from Radarr...")
	radarrMovies, err := ri.client.GetMovies()
	if err != nil {
		status.Status = "failed"
		status.ErrorMessage = err.Error()
		repository.UpdateScanStatus(ri.db, status)
		return nil, fmt.Errorf("failed to fetch movies from Radarr: %w", err)
	}

	// Filter to only movies with files
	var moviesWithFiles []RadarrMovie
	for _, m := range radarrMovies {
		if m.HasFile && m.MovieFile != nil {
			moviesWithFiles = append(moviesWithFiles, m)
		}
	}

	result.FilesFound = len(moviesWithFiles)
	log.Printf("Found %d movies with files in Radarr (out of %d total)", result.FilesFound, len(radarrMovies))

	// Update status with count
	status.FilesFound = result.FilesFound
	repository.UpdateScanStatus(ri.db, status)

	// Broadcast scan start
	if ri.broadcaster != nil {
		ri.broadcaster.BroadcastScanStart(result.FilesFound, status.StartedAt)
	}

	// Track TMDB IDs we've seen (for deletion detection)
	seenTMDBIds := make(map[int64]bool)

	// Process each movie
	for i, rm := range moviesWithFiles {
		// Check for stop signal
		select {
		case <-ri.stopChan:
			status.Status = "stopped"
			status.CompletedAt = time.Now().Format(time.RFC3339)
			status.ErrorMessage = "Import stopped by user"
			repository.UpdateScanStatus(ri.db, status)
			if ri.broadcaster != nil {
				ri.broadcaster.BroadcastScanStopped()
			}
			return result, fmt.Errorf("import stopped by user")
		default:
		}

		if err := ri.processRadarrMovie(&rm, result); err != nil {
			log.Printf("Error processing movie '%s': %v", rm.Title, err)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", rm.Title, err))
		} else {
			seenTMDBIds[int64(rm.TmdbID)] = true
		}

		result.FilesProcessed++

		// Update progress periodically
		if i%10 == 0 || i == len(moviesWithFiles)-1 {
			status.FilesProcessed = result.FilesProcessed
			repository.UpdateScanStatus(ri.db, status)
			if ri.broadcaster != nil {
				ri.broadcaster.BroadcastScanProgress(result.FilesProcessed, result.FilesFound)
			}
		}
	}

	// Full sync: Remove movies that are no longer in Radarr
	deletedCount, err := ri.removeStaleMovies(seenTMDBIds)
	if err != nil {
		log.Printf("Warning: Failed to remove stale movies: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove stale movies: %v", err))
	} else if deletedCount > 0 {
		log.Printf("Removed %d movies no longer in Radarr", deletedCount)
	}

	// Update status to completed
	status.Status = "completed"
	status.CompletedAt = time.Now().Format(time.RFC3339)
	status.FilesProcessed = result.FilesProcessed
	if len(result.Errors) > 0 {
		status.ErrorMessage = fmt.Sprintf("%d errors during import", len(result.Errors))
	}
	repository.UpdateScanStatus(ri.db, status)

	// Broadcast completion
	if ri.broadcaster != nil {
		ri.broadcaster.BroadcastScanComplete(result.FilesProcessed, result.MoviesAdded, 0)
	}

	duration := time.Since(start)
	log.Printf("Radarr import completed in %v - %d movies processed, %d added/updated, %d removed, %d errors",
		duration.Round(time.Second), result.FilesProcessed, result.MoviesAdded, deletedCount, len(result.Errors))

	return result, nil
}

// processRadarrMovie processes a single movie from Radarr
func (ri *RadarrImporter) processRadarrMovie(rm *RadarrMovie, result *models.ScanResult) error {
	// Map Radarr movie to our model
	movie := ri.mapRadarrMovie(rm)

	// Extract mediainfo from the actual file
	if rm.MovieFile != nil && rm.MovieFile.Path != "" {
		mediaInfo, fileSize, duration, err := ri.extractor.Extract(rm.MovieFile.Path)
		if err != nil {
			log.Printf("Mediainfo extraction failed for %s: %v", rm.Title, err)
			// Continue with Radarr's file size
			movie.FileSize = rm.MovieFile.Size
		} else {
			movie.MediaInfo = mediaInfo
			movie.FileSize = fileSize
			if duration > 0 {
				movie.Duration = duration / 60 // Convert seconds to minutes
			}
		}
	}

	// Check if movie already exists by TMDB ID
	existing, err := repository.GetMovieByTMDBId(ri.db, movie.TMDBId)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing movie: %w", err)
	}

	if existing != nil {
		// Update existing movie
		movie.ID = existing.ID
		if err := repository.UpdateMovie(ri.db, movie); err != nil {
			return fmt.Errorf("failed to update movie: %w", err)
		}
		log.Printf("Updated movie: %s (%d)", movie.Title, movie.Year)
	} else {
		// Insert new movie
		_, err := repository.InsertMovie(ri.db, movie)
		if err != nil {
			return fmt.Errorf("failed to insert movie: %w", err)
		}
		result.MoviesAdded++
		log.Printf("Added movie: %s (%d)", movie.Title, movie.Year)
	}

	return nil
}

// mapRadarrMovie converts a RadarrMovie to our internal Movie model
func (ri *RadarrImporter) mapRadarrMovie(rm *RadarrMovie) *models.Movie {
	movie := &models.Movie{
		Title:     rm.Title,
		Year:      rm.Year,
		Synopsis:  rm.Overview,
		Duration:  rm.Runtime, // minutes
		Status:    "available",
		TMDBId:    int64(rm.TmdbID),
		IMDbId:    rm.ImdbID,
		DateAdded: rm.Added,
		Container: ri.extractContainer(rm),
	}

	// Genres (join array)
	if len(rm.Genres) > 0 {
		movie.Genres = strings.Join(rm.Genres, ", ")
	}

	// Rating (prefer TMDB, fallback to IMDB)
	if rm.Ratings.Tmdb.Value > 0 {
		movie.Rating = rm.Ratings.Tmdb.Value
	} else if rm.Ratings.Imdb.Value > 0 {
		movie.Rating = rm.Ratings.Imdb.Value
	}

	// Poster URL
	posterURL := ri.extractPosterURL(rm)
	if posterURL != "" {
		movie.Poster = &posterURL
	}

	// File path and size
	if rm.MovieFile != nil {
		movie.FilePath = rm.MovieFile.Path
		movie.FileSize = rm.MovieFile.Size
	}

	// Use Radarr's added date, fallback to now
	if movie.DateAdded == "" {
		movie.DateAdded = time.Now().Format(time.RFC3339)
	}

	return movie
}

// extractPosterURL extracts the poster URL from Radarr images
func (ri *RadarrImporter) extractPosterURL(rm *RadarrMovie) string {
	for _, img := range rm.Images {
		if img.CoverType == "poster" {
			// Prefer remote URL (TMDB), fallback to local URL
			if img.RemoteURL != "" {
				return img.RemoteURL
			}
			if img.URL != "" {
				// Local URL needs Radarr base URL prepended
				return fmt.Sprintf("%s%s", ri.client.baseURL, img.URL)
			}
		}
	}
	return ""
}

// extractContainer extracts the container format from the file path
func (ri *RadarrImporter) extractContainer(rm *RadarrMovie) string {
	if rm.MovieFile == nil || rm.MovieFile.Path == "" {
		return ""
	}
	return GetContainer(rm.MovieFile.Path)
}

// removeStaleMovies removes movies that exist in our DB but not in Radarr
func (ri *RadarrImporter) removeStaleMovies(seenTMDBIds map[int64]bool) (int, error) {
	// Get all TMDB IDs from our database
	existingIds, err := repository.GetAllMovieTMDBIds(ri.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get existing movie IDs: %w", err)
	}

	var deletedCount int
	for _, tmdbId := range existingIds {
		if tmdbId == 0 {
			// Skip movies without TMDB ID (shouldn't happen with Radarr imports)
			continue
		}
		if !seenTMDBIds[tmdbId] {
			// Movie is in our DB but not in Radarr - delete it
			if err := repository.DeleteMovieByTMDBId(ri.db, tmdbId); err != nil {
				log.Printf("Failed to delete movie with TMDB ID %d: %v", tmdbId, err)
			} else {
				deletedCount++
			}
		}
	}

	return deletedCount, nil
}

// ImportMovie imports/refreshes a single movie by its database ID
func (ri *RadarrImporter) ImportMovie(movieID int64) (*models.ScanResult, error) {
	log.Printf("Starting single movie refresh for ID: %d", movieID)
	start := time.Now()

	result := &models.ScanResult{
		Errors: []string{},
	}

	// Get the movie from our database to find its TMDB ID
	movie, err := repository.GetMovieByID(ri.db, movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie: %w", err)
	}
	if movie == nil {
		return nil, fmt.Errorf("movie not found with ID: %d", movieID)
	}

	// Fetch updated data from Radarr by TMDB ID
	radarrMovie, err := ri.client.GetMovieByTMDBId(int(movie.TMDBId))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie from Radarr: %w", err)
	}
	if radarrMovie == nil {
		// Movie no longer in Radarr - delete it
		log.Printf("Movie no longer in Radarr, deleting: %s", movie.Title)
		if err := repository.DeleteMovie(ri.db, movieID); err != nil {
			return nil, fmt.Errorf("failed to delete movie: %w", err)
		}
		result.FilesProcessed = 1
		return result, nil
	}

	// Process the movie
	result.FilesFound = 1
	if err := ri.processRadarrMovie(radarrMovie, result); err != nil {
		return nil, fmt.Errorf("failed to process movie: %w", err)
	}
	result.FilesProcessed = 1

	duration := time.Since(start)
	log.Printf("Movie refresh completed in %d ms", duration.Milliseconds())

	return result, nil
}

// GetStatus returns the current import/scan status
func (ri *RadarrImporter) GetStatus() (*models.ScanStatus, error) {
	return repository.GetScanStatus(ri.db)
}
