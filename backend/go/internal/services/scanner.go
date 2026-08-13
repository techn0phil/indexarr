package services

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
)

// scanCache holds per-scan cached data to optimize API calls
type scanCache struct {
	// Series lookup by normalized title (avoids repeated DB queries)
	seriesByTitle map[string]*models.Series
	// Full series metadata by TVDB ID (from extended endpoint)
	// seriesExtendedByTVDBId map[int]*TVDBSeriesExtended
	// All episodes by series TVDB ID (from bulk episodes endpoint)
	// episodesByTVDBId map[int][]TVDBBulkEpisode
	episodesByTVDBId map[int]*TVDBAllEpisodesResponse
	// Failed enrichment tracking (prevents retry loops)
	failedSeriesByTitle map[string]error
}

// Scanner handles media library scanning
type Scanner struct {
	db          *sql.DB
	config      *config.Config
	extractor   *Extractor
	tmdb        *TMDBClient
	tv          *TVClient
	broadcaster *Broadcaster
	running     bool
	stopChan    chan struct{}
	mu          sync.Mutex
	cache       *scanCache
}

// NewScanner creates a new scanner service
func NewScanner(db *sql.DB, cfg *config.Config, broadcaster *Broadcaster) *Scanner {
	return &Scanner{
		db:          db,
		config:      cfg,
		extractor:   NewExtractor(cfg.MediainfoPath, cfg.ScanTimeout),
		tmdb:        NewTMDBClient(cfg.TMDBAPIKey, cfg.DetectionLanguage, cfg.MetadataLanguage),
		tv:          NewTVClient(cfg.TVDBAPIKey, db), // Uses TVDB API v4 for TV shows
		broadcaster: broadcaster,
		stopChan:    make(chan struct{}),
	}
}

// IsRunning returns whether a scan is currently in progress
func (s *Scanner) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Stop signals the scanner to stop
func (s *Scanner) Stop() {
	s.mu.Lock()
	if s.running {
		close(s.stopChan)
	}
	s.mu.Unlock()
}

// Scan performs a full library scan
func (s *Scanner) Scan() (*models.ScanResult, error) {
	return s.ScanPaths(s.config.MediaLibraryPaths, nil)
}

// CountMediaFiles counts video files in the given paths without processing them
// This is a fast operation (no mediainfo extraction) for progress estimation
func (s *Scanner) CountMediaFiles(paths []string) (int, error) {
	var count int
	for _, libPath := range paths {
		if libPath == "" {
			continue
		}

		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(libPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Continue walking
			}

			if d.IsDir() {
				name := d.Name()

				// Skip hidden directories
				if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "@") {
					return fs.SkipDir
				}

				// Skip extra media folders
				for _, extraFolder := range s.config.SkipFolders {
					if strings.EqualFold(name, extraFolder) {
						return fs.SkipDir
					}
				}

				return nil
			}

			// Skip files matching ignore pattern
			if s.config.IgnoreFilePattern != nil {
				if s.config.IgnoreFilePattern.MatchString(filepath.Base(path)) {
					return nil
				}
			}

			// Count video files
			if IsVideoFile(path) {
				count++
			}

			return nil
		})

		if err != nil {
			config.GlobalLogger.Error().Err(err).Str("path", libPath).Msg("Error counting files")
		}
	}

	return count, nil
}

// ScanPaths performs a scan on specified paths (used for manual scans via API)
// If ctx is nil, uses default behavior. If ctx is provided, applies progress coordination.
func (s *Scanner) ScanPaths(paths []string, ctx *models.ProgressContext) (*models.ScanResult, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil, fmt.Errorf("scan already in progress")
	}
	s.running = true
	s.stopChan = make(chan struct{})

	// Initialize per-scan cache for API call optimization
	s.cache = &scanCache{
		seriesByTitle: make(map[string]*models.Series),
		// seriesExtendedByTVDBId: make(map[int]*TVDBSeriesExtended),
		// episodesByTVDBId:    make(map[int][]TVDBBulkEpisode),
		episodesByTVDBId:    make(map[int]*TVDBAllEpisodesResponse),
		failedSeriesByTitle: make(map[string]error),
	}

	s.mu.Unlock()

	config.GlobalLogger.Info().Msg("Scan starting")
	start := time.Now()

	defer func() {
		s.mu.Lock()
		s.running = false

		select {
		case <-s.stopChan:
			// Already closed, do nothing
		default:
			close(s.stopChan)
		}

		// Clear per-scan cache
		s.cache = nil
		s.mu.Unlock()
	}()

	result := &models.ScanResult{
		Errors: []string{},
	}

	// Check if we should suppress start/complete broadcasts (coordinated by scheduler)
	suppressBroadcasts := ctx != nil && ctx.SuppressStartComplete
	progressOffset := 0
	progressTotal := 0
	if ctx != nil {
		progressOffset = ctx.Offset
		progressTotal = ctx.TotalOverride
	}

	// Update scan status to running (only if not suppressed)
	status := &models.ScanStatus{
		Status:     "running",
		StartedAt:  time.Now().Format(time.RFC3339),
		FilesFound: 0,
	}
	if !suppressBroadcasts {
		if err := repository.UpdateScanStatus(s.db, status); err != nil {
			config.GlobalLogger.Warn().Err(err).Msg("Failed to update scan status")
		}

		// Broadcast scan start to WebSocket clients
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanStart(result.FilesFound, status.StartedAt)
		}
	}

	// Collect all media files
	var files []string
	for _, libPath := range paths {
		if libPath == "" {
			continue
		}

		config.GlobalLogger.Info().Str("path", libPath).Msg("Scanning library path")

		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("Path does not exist: %s", libPath))
			continue
		}

		err := filepath.WalkDir(libPath, func(path string, d fs.DirEntry, err error) error {
			// Check for stop signal
			select {
			case <-s.stopChan:
				return fmt.Errorf("scan stopped by user")
			default:
			}

			if err != nil {
				config.GlobalLogger.Error().Err(err).Str("path", path).Msg("Error accessing path")
				return nil // Continue walking
			}

			if d.IsDir() {
				name := d.Name()

				// Skip hidden directories
				if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "@") {
					return fs.SkipDir
				}

				// Skip extra media folders
				for _, extraFolder := range s.config.SkipFolders {
					if strings.EqualFold(name, extraFolder) {
						return fs.SkipDir
					}
				}

				// Handle Bluray format (skip BDMV/STREAM content, but include the BDMV folder itself for metadata extraction)
				if strings.EqualFold(name, "STREAM") && strings.EqualFold(filepath.Base(filepath.Dir(path)), "BDMV") {
					files = append(files, filepath.Dir(path))
					return fs.SkipDir
				}

				return nil
			}

			// Skip files matching ignore pattern
			if s.config.IgnoreFilePattern != nil {
				if s.config.IgnoreFilePattern.MatchString(filepath.Base(path)) {
					return nil
				}
			}

			// Check if it's a video file
			if IsVideoFile(path) {
				files = append(files, path)
			}

			return nil
		})

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error walking %s: %v", libPath, err))
		}
	}

	result.FilesFound = len(files)
	config.GlobalLogger.Info().Int("count", result.FilesFound).Msg("Found media files")

	// Update status with files found
	status.FilesFound = result.FilesFound
	if !suppressBroadcasts {
		repository.UpdateScanStatus(s.db, status)

		// Broadcast scan start to WebSocket clients
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanStart(result.FilesFound, status.StartedAt)
		}
	}

	var seenFiles []string

	// Process each file sequentially
	for i, filePath := range files {
		// Check for stop signal
		select {
		case <-s.stopChan:
			if !suppressBroadcasts {
				status.Status = "stopped"
				status.CompletedAt = time.Now().Format(time.RFC3339)
				status.ErrorMessage = "Scan stopped by user"
				repository.UpdateScanStatus(s.db, status)
				// Broadcast stopped event to WebSocket clients
				if s.broadcaster != nil {
					s.broadcaster.BroadcastScanStopped()
				}
			}
			return result, fmt.Errorf("scan stopped by user")
		default:
		}

		if err := s.processFile(filePath, result); err != nil {
			config.GlobalLogger.Error().Err(err).Str("file", filePath).Msg("Error processing file")
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", filepath.Base(filePath), err))
		} else {
			seenFiles = append(seenFiles, filePath)
		}

		result.FilesProcessed++

		// Update progress periodically
		if i%10 == 0 || i == len(files)-1 {
			status.FilesProcessed = result.FilesProcessed
			if !suppressBroadcasts {
				repository.UpdateScanStatus(s.db, status)
			}
			// Broadcast progress to WebSocket clients
			if s.broadcaster != nil {
				// Apply progress offset and total override if coordinated
				reportProcessed := result.FilesProcessed + progressOffset
				reportTotal := result.FilesFound
				if progressTotal > 0 {
					reportTotal = progressTotal
				}
				s.broadcaster.BroadcastScanProgress(reportProcessed, reportTotal)
			}
		}
	}

	// If paths are exactly the library paths, we can be confident that any file not seen during the scan is truly missing from disk.
	// In this case, we can safely remove stale entries from the database.
	// However, if the scan was performed on a subset of folders (e.g. for a specific series), we should not remove entries for files
	// that were outside the scanned folders, as they may still exist on disk.
	// Therefore, we only perform the stale file removal if the scanned paths match the configured library paths.
	// This prevents false deletions in cases where users scan specific folders that do not encompass the entire library.

	deletedMoviesCount := 0
	deletedEpisodesCount := 0
	if s.pathsMatchAnyLibrary(paths) {
		// Handle deletions: find movies/episodes in database that were not seen during scan and check if their files are missing from disk.
		// If so, mark them as deleted. This ensures that if a file was deleted from disk, it will be reflected in the database after the scan.
		config.GlobalLogger.Info().Msg("Removing stale media files not seen during scan")
		var err error
		deletedMoviesCount, deletedEpisodesCount, err = s.removeStaleMediaFiles(paths, seenFiles)
		if err != nil {
			config.GlobalLogger.Error().Err(err).Msg("Failed to remove stale media files")
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove stale media files: %v", err))
		} else if deletedMoviesCount > 0 || deletedEpisodesCount > 0 {
			config.GlobalLogger.Info().Int("movies", deletedMoviesCount).Int("episodes", deletedEpisodesCount).Msg("Removed stale media files")
		}

		// Remove series that have no episodes left after episode deletions
		if err := repository.DeleteEmptySeries(s.db); err != nil {
			config.GlobalLogger.Error().Err(err).Msg("Failed to delete empty series")
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to delete empty series: %v", err))
		}
	}

	// Update status to completed (only if not suppressed)
	if !suppressBroadcasts {
		status.Status = "completed"
		status.CompletedAt = time.Now().Format(time.RFC3339)
		status.FilesProcessed = result.FilesProcessed
		if len(result.Errors) > 0 {
			status.ErrorMessage = fmt.Sprintf("%d errors during scan", len(result.Errors))
		}
		repository.UpdateScanStatus(s.db, status)

		// Broadcast completion to WebSocket clients
		if s.broadcaster != nil {
			s.broadcaster.BroadcastScanComplete(result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded, status.CompletedAt)
		}
	}

	duration := time.Since(start)
	config.GlobalLogger.Info().
		Dur("duration", duration).
		Int("files_processed", result.FilesProcessed).
		Int("movies_added", result.MoviesAdded).
		Int("episodes_added", result.EpisodesAdded).
		Int("movies_removed", deletedMoviesCount).
		Int("episodes_removed", deletedEpisodesCount).
		Int("errors", len(result.Errors)).
		Msg("Scan completed")

	if len(result.Errors) > 0 {
		// Log the first 100 errors for visibility
		for i, err := range result.Errors {
			if i >= 100 {
				config.GlobalLogger.Warn().Int("count", len(result.Errors)-100).Msg("... and more")
				break
			}
			config.GlobalLogger.Warn().Str("error", err).Msg("Scan error")
		}
	}

	return result, nil
}

// pathsMatchAnyLibrary checks if all the scanned paths match some of the configured library paths (ignoring trailing slashes and case)
// Library paths checked are MediaLibraryPaths, MoviesLibraryPaths, and SeriesLibraryPaths to cover all possible configurations
// (e.g. users who have separate folders for movies and series, or users who have only one of the two types configured)
func (s *Scanner) pathsMatchAnyLibrary(paths []string) bool {
	for _, scanPath := range paths {
		scanPathClean := strings.TrimRight(filepath.Clean(scanPath), string(os.PathSeparator))
		matchFound := false
		for _, libPath := range s.config.MediaLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		for _, libPath := range s.config.MoviesLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		for _, libPath := range s.config.SeriesLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return false
		}
	}
	return true
}

// pathsMatchMovieOrMediaLibrary checks if all the scanned paths match some of the configured movie or media library paths (ignoring trailing slashes and case)
func (s *Scanner) pathsMatchMovieOrMediaLibrary(paths []string) bool {
	for _, scanPath := range paths {
		scanPathClean := strings.TrimRight(filepath.Clean(scanPath), string(os.PathSeparator))
		matchFound := false
		for _, libPath := range s.config.MediaLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		for _, libPath := range s.config.MoviesLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return false
		}
	}
	return true
}

// pathsMatchSeriesOrMediaLibrary checks if all the scanned paths match some of the configured series or media library paths (ignoring trailing slashes and case)
func (s *Scanner) pathsMatchSeriesOrMediaLibrary(paths []string) bool {
	for _, scanPath := range paths {
		scanPathClean := strings.TrimRight(filepath.Clean(scanPath), string(os.PathSeparator))
		matchFound := false
		for _, libPath := range s.config.SeriesLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		for _, libPath := range s.config.MediaLibraryPaths {
			libPathClean := strings.TrimRight(filepath.Clean(libPath), string(os.PathSeparator))
			if strings.EqualFold(scanPathClean, libPathClean) {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return false
		}
	}
	return true
}

// removeStaleMediaFiles removes media files that exist in our DB but not on disk
func (s *Scanner) removeStaleMediaFiles(paths []string, seenFiles []string) (int, int, error) {
	seenFilesMap := make(map[string]bool)
	for _, file := range seenFiles {
		seenFilesMap[file] = true
	}

	// Print seen files count for debugging
	config.GlobalLogger.Debug().Int("seen_count", len(seenFiles)).Msg("Checking for stale entries in database")

	deletedMoviesCount := 0
	deletedEpisodesCount := 0

	if s.pathsMatchMovieOrMediaLibrary(paths) {
		// Remove movies not seen during scan
		moviePaths, err := repository.GetAllMovieFilePaths(s.db)
		if err != nil {
			return deletedMoviesCount, deletedEpisodesCount, fmt.Errorf("failed to fetch movies for deletion check: %w", err)
		}

		// Print movie paths count for debugging
		config.GlobalLogger.Debug().Int("count", len(moviePaths)).Msg("Fetched movie file paths from database")

		for _, moviePath := range moviePaths {
			if !seenFilesMap[moviePath] {
				if _, err := os.Stat(moviePath); os.IsNotExist(err) {
					config.GlobalLogger.Info().Str("path", moviePath).Msg("Movie file missing from disk, deleting")
					if err := repository.DeleteMovieByPath(s.db, moviePath); err != nil {
						config.GlobalLogger.Error().Err(err).Msg("Failed to delete movie")
					} else {
						deletedMoviesCount++
					}
				}
			}
		}

		config.GlobalLogger.Info().Int("count", deletedMoviesCount).Msg("Removed movies no longer on disk")
	}

	if s.pathsMatchSeriesOrMediaLibrary(paths) {
		// Remove episodes not seen during scan
		episodePaths, err := repository.GetAllEpisodeFilePaths(s.db)
		if err != nil {
			return deletedMoviesCount, deletedEpisodesCount, fmt.Errorf("failed to fetch episodes for deletion check: %w", err)
		}

		// Print episode paths count for debugging
		config.GlobalLogger.Debug().Int("count", len(episodePaths)).Msg("Fetched episode file paths from database")

		for _, episodePath := range episodePaths {
			if !seenFilesMap[episodePath] {
				if _, err := os.Stat(episodePath); os.IsNotExist(err) {
					config.GlobalLogger.Info().Str("path", episodePath).Msg("Episode file missing from disk, deleting")
					if err := repository.DeleteEpisodeByPath(s.db, episodePath); err != nil {
						config.GlobalLogger.Error().Err(err).Msg("Failed to delete episode")
					} else {
						deletedEpisodesCount++
					}
				}
			}
		}

		config.GlobalLogger.Info().Int("count", deletedEpisodesCount).Msg("Removed episodes no longer on disk")
	}

	return deletedMoviesCount, deletedEpisodesCount, nil
}

// ScanMovie scans a single movie (used for manual refresh via API)
func (s *Scanner) ScanMovie(movieID int64) (*models.ScanResult, error) {
	config.GlobalLogger.Info().Int64("id", movieID).Msg("Starting movie refresh")
	start := time.Now()

	movie, err := repository.GetMovieByID(s.db, movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie: %w", err)
	}
	if movie == nil {
		return nil, fmt.Errorf("movie not found with ID: %d", movieID)
	}

	config.GlobalLogger.Info().Str("title", movie.Title).Int("year", movie.Year).Msg("Found movie")

	result, err := s.ScanPaths([]string{movie.FilePath}, nil)
	if err != nil {
		return nil, err
	}

	// Remove movie if it was deleted from disk
	if result.FilesProcessed == 0 {
		config.GlobalLogger.Info().Str("path", movie.FilePath).Msg("Movie file not found during refresh, deleting")
		if err := repository.DeleteMovie(s.db, movieID); err != nil {
			config.GlobalLogger.Error().Err(err).Msg("Failed to delete movie")
		}
	} else {
		// Extract media info again to update any changes (e.g. new audio tracks)
		mediaInfo, fileSize, duration, err := s.extractor.Extract(movie.FilePath)
		if err != nil {
			config.GlobalLogger.Warn().Err(err).Str("title", movie.Title).Msg("Mediainfo extraction failed during refresh")
		} else {
			movie.MediaInfo = mediaInfo
			movie.FileSize = fileSize
			movie.Duration = duration / 60 // Convert seconds to minutes
		}

		// Parse filename
		parsed := ParseFilename(movie.FilePath)

		movie.Title = parsed.Title
		movie.Year = parsed.Year
		movie.Status = "available"

		// Try to enrich with TMDB metadata
		if err := s.tmdb.EnrichMovie(movie); err != nil {
			config.GlobalLogger.Warn().Err(err).Str("title", movie.Title).Msg("TMDB enrichment failed during refresh")
		} else {
			// Update movie with new metadata
			if err := repository.UpdateMovie(s.db, movie); err != nil {
				config.GlobalLogger.Warn().Err(err).Msg("Failed to update movie during refresh")
			}
			config.GlobalLogger.Info().Str("title", movie.Title).Int("year", movie.Year).Msg("Movie refreshed")
			result.MoviesAdded = 1 // Count as "added" for refresh purposes
			result.FilesProcessed = 1
			result.FilesFound = 1
			result.Errors = []string{}
			result.EpisodesAdded = 0
		}
	}

	duration := time.Since(start)
	config.GlobalLogger.Info().Dur("duration", duration).Msg("Movie refresh completed")

	return result, nil
}

// ScanSeries scans a single series (used for manual refresh via API)
func (s *Scanner) ScanSeries(seriesID int64) (*models.ScanResult, error) {
	config.GlobalLogger.Info().Int64("id", seriesID).Msg("Starting series refresh")
	start := time.Now()

	result := &models.ScanResult{
		Errors: []string{},
	}

	// Step 1: Fetch series from database
	series, err := repository.GetSeriesByID(s.db, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series: %w", err)
	}
	if series == nil {
		return nil, fmt.Errorf("series not found with ID: %d", seriesID)
	}

	config.GlobalLogger.Info().Str("title", series.Title).Int("year", series.YearStart).Msg("Found series")

	// Step 2: Fetch all episodes to determine folders to scan
	episodes, err := repository.GetAllEpisodesForSeries(s.db, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}

	// Extract unique folder paths from existing episodes
	folderPaths := s.findSeriesFolderPaths(episodes)
	config.GlobalLogger.Info().Int("count", len(folderPaths)).Strs("paths", folderPaths).Msg("Found folders for series")

	// Step 3: Scan folder paths to detect new episodes
	scanResult, err := s.ScanPaths(folderPaths, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to scan series folders: %w", err)
	}

	result.FilesFound = scanResult.FilesFound

	// Step 4 & 5: Check for missing episodes and delete them from database
	episodesToDelete := []int64{}
	for _, episode := range episodes {
		// Check if file still exists
		if _, err := os.Stat(episode.FilePath); os.IsNotExist(err) {
			config.GlobalLogger.Info().Str("path", episode.FilePath).Int("season", episode.SeasonNum).Int("episode", episode.EpisodeNum).Msg("Episode file missing, marking for removal")
			episodesToDelete = append(episodesToDelete, episode.ID)
		}
	}

	// Delete missing episodes
	for _, episodeID := range episodesToDelete {
		if err := repository.DeleteEpisode(s.db, episodeID); err != nil {
			errMsg := fmt.Sprintf("Failed to remove missing episode %d: %v", episodeID, err)
			config.GlobalLogger.Error().Err(err).Int64("id", episodeID).Msg("Failed to remove missing episode")
			result.Errors = append(result.Errors, errMsg)
		}
	}

	// Delete seasons that have no episodes left
	if err := repository.DeleteEmptySeasons(s.db, seriesID); err != nil {
		config.GlobalLogger.Error().Err(err).Msg("Failed to delete empty seasons")
	}

	// Step 6: Check if series folder is completely missing
	if scanResult.FilesFound == 0 && len(episodesToDelete) == len(episodes) {
		config.GlobalLogger.Info().Str("title", series.Title).Msg("All episodes missing from disk, deleting series")
		if err := repository.DeleteSeries(s.db, seriesID); err != nil {
			errMsg := fmt.Sprintf("Failed to delete series: %v", err)
			config.GlobalLogger.Error().Err(err).Msg("Failed to delete series")
			result.Errors = append(result.Errors, errMsg)
		}
		return result, nil
	}

	// Step 7: Extract media info for each episode again to catch any changes
	for _, episode := range episodes {
		mediaInfo, fileSize, duration, err := s.extractor.Extract(episode.FilePath)
		if err != nil {
			config.GlobalLogger.Warn().Err(err).Str("path", episode.FilePath).Msg("Mediainfo extraction failed for episode")
			// Continue with minimal info
			mediaInfo = &models.MediaInfo{
				VideoTracks:    []models.VideoTrack{},
				AudioTracks:    []models.AudioTrack{},
				SubtitleTracks: []models.SubtitleTrack{},
			}
		}

		episode.MediaInfo = mediaInfo
		episode.FileSize = fileSize
		episode.Duration = duration

		// Update episode in database
		if err := repository.UpdateEpisode(s.db, &episode); err != nil {
			config.GlobalLogger.Warn().Err(err).Int("season", episode.SeasonNum).Int("episode", episode.EpisodeNum).Msg("Failed to update episode during refresh")
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to update episode S%02dE%02d: %v", episode.SeasonNum, episode.EpisodeNum, err))
		}
	}

	// Re-enrich series metadata from TMDB
	if err := s.tmdb.EnrichSeries(series); err != nil {
		config.GlobalLogger.Warn().Err(err).Str("title", series.Title).Msg("TMDB enrichment failed during refresh")
	}

	// Update series in database
	if err := repository.UpdateSeries(s.db, series); err != nil {
		config.GlobalLogger.Warn().Err(err).Msg("Failed to update series during refresh")
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update series: %v", err))
	}

	// Recalculate series counts
	if err := repository.UpdateSeriesCounts(s.db, seriesID); err != nil {
		config.GlobalLogger.Error().Err(err).Msg("Failed to update series counts")
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update series counts: %v", err))
	}

	result.FilesProcessed = scanResult.FilesProcessed
	result.EpisodesAdded = scanResult.EpisodesAdded
	result.MoviesAdded = 0 // No movies in series refresh

	// Merge any errors from the scan
	result.Errors = append(result.Errors, scanResult.Errors...)

	duration := time.Since(start)
	config.GlobalLogger.Info().
		Dur("duration", duration).
		Int("files_processed", result.FilesProcessed).
		Int("episodes_added", result.EpisodesAdded).
		Int("episodes_deleted", len(episodesToDelete)).
		Int("errors", len(result.Errors)).
		Msg("Series refresh completed")

	return result, nil
}

// findSeriesFolderPaths extracts unique parent directory paths from a list of episodes
// This supports series that may be split across multiple folders
func (s *Scanner) findSeriesFolderPaths(episodes []models.Episode) []string {
	folderMap := make(map[string]bool)

	for _, episode := range episodes {
		dir := filepath.Dir(episode.FilePath)
		folderMap[dir] = true
	}

	folders := make([]string, 0, len(folderMap))
	for folder := range folderMap {
		folders = append(folders, folder)
	}

	// Find common parent folder which does not belong to any media library path to avoid scanning the entire library if series episodes are stored in different folders. This is a common edge case for users who have their TV shows organized in multiple folders (e.g. by genre, by quality, etc.) but still want to be able to refresh the entire series metadata with one click.
	commonParent := findCommonParentFolder(folders)
	if commonParent != "" && !s.isLibraryPath(commonParent) {
		config.GlobalLogger.Debug().Str("path", commonParent).Msg("Series episodes in multiple folders, using common parent")
		return []string{commonParent}
	}

	return folders
}

// findCommonParentFolder takes a list of folder paths and returns the common parent folder if it exists, or an empty string if there is no common parent
func findCommonParentFolder(folders []string) string {
	if len(folders) == 0 {
		return ""
	}

	commonParent := folders[0]
	for _, folder := range folders[1:] {
		for !strings.HasPrefix(folder, commonParent) {
			commonParent = filepath.Dir(commonParent)
			if commonParent == "." || commonParent == "/" {
				return ""
			}
		}
	}

	return commonParent
}

// isLibraryPath checks if a given path is one of the configured media library paths
func (s *Scanner) isLibraryPath(path string) bool {
	for _, libPath := range s.config.MediaLibraryPaths {
		if strings.EqualFold(filepath.Clean(libPath), filepath.Clean(path)) {
			return true
		}
	}
	return false
}

// preFetchSeriesData fetches all metadata for a series in bulk (2 API calls total)
// This dramatically reduces API calls: instead of 1 call per episode, we fetch everything at once
func (s *Scanner) preFetchSeriesData(tvdbID int) error {
	// Check if already fetched
	// if _, exists := s.cache.seriesExtendedByTVDBId[tvdbID]; exists {
	// 	log.Printf("[Cache] Series metadata already cached for TVDB ID %d", tvdbID)
	// 	return nil
	// }

	config.GlobalLogger.Debug().Int("tvdbID", tvdbID).Msg("Pre-fetching all metadata for series")

	// // Fetch extended series metadata (includes seasons array)
	// seriesExtended, err := s.tv.GetTVDetails(tvdbID)
	// if err != nil {
	// 	return fmt.Errorf("failed to fetch series extended metadata: %w", err)
	// }

	// s.cache.seriesExtendedByTVDBId[tvdbID] = seriesExtended
	// log.Printf("[Cache] Successfully cached series metadata: %d seasons", len(seriesExtended.Data.Seasons))

	// Fetch all episodes in bulk
	allEpisodes, err := s.tv.GetAllEpisodes(tvdbID, s.config.MetadataLanguage)
	if err != nil {
		return fmt.Errorf("failed to fetch bulk episodes: %w", err)
	}
	// s.cache.episodesByTVDBId[tvdbID] = allEpisodes.Data.Episodes
	s.cache.episodesByTVDBId[tvdbID] = allEpisodes

	config.GlobalLogger.Debug().Int("count", len(allEpisodes.Data.Episodes)).Msg("Successfully cached episodes data")

	return nil
}

// enrichEpisodeFromCache populates episode metadata from cached bulk data (no API calls)
func (s *Scanner) enrichEpisodeFromCache(episode *models.Episode, seriesTVDBID int, seasonNum, episodeNum int) {
	// Get cached episodes for this series
	cachedEpisodes, exists := s.cache.episodesByTVDBId[seriesTVDBID]
	if !exists {
		config.GlobalLogger.Debug().Int("tvdbID", seriesTVDBID).Msg("No cached episodes found for series")
		return
	}

	// Find matching episode by season and episode number
	for _, ep := range cachedEpisodes.Data.Episodes {
		if ep.SeasonNumber == seasonNum && ep.Number == episodeNum {
			episode.Title = ep.Name
			if episode.Duration == 0 && ep.Runtime > 0 {
				episode.Duration = ep.Runtime * 60 // Convert minutes to seconds
			}
			config.GlobalLogger.Debug().Int("season", seasonNum).Int("episode", episodeNum).Str("title", ep.Name).Int("runtime", ep.Runtime).Msg("Enriched episode from cache")
			return
		}
	}

	// Episode not found in cache, use default title
	config.GlobalLogger.Debug().Int("season", seasonNum).Int("episode", episodeNum).Int("tvdbID", seriesTVDBID).Msg("Episode not found in cached data")
	episode.Title = fmt.Sprintf("Episode %d", episodeNum)
}

// processFile handles a single media file
func (s *Scanner) processFile(filePath string, result *models.ScanResult) error {
	// Get time before processing for performance logging
	processStart := time.Now()

	// Parse filename
	parsed := ParseFilename(filePath)

	var processError error

	if parsed.IsSeries {
		processError = s.processEpisode(filePath, parsed, result)
	} else {
		processError = s.processMovie(filePath, parsed, result)
	}

	// Total processing time for the file
	processDuration := time.Since(processStart)
	config.GlobalLogger.Trace().Int64("duration_ms", processDuration.Milliseconds()).Str("file", filePath).Msg("Total processing time for file")

	return processError
}

// processMovie handles a movie file
func (s *Scanner) processMovie(filePath string, parsed *ParsedFilename, result *models.ScanResult) error {
	isBluray := IsBlurayFolder(filePath)
	mediaFilePath := filePath

	var mediaInfo *models.MediaInfo
	var fileSize int64
	var duration int
	var err error

	if !isBluray {
		// Check if movie already exists by file path
		exists, err := repository.MovieExistsByFilePath(s.db, filePath)
		if err != nil {
			return fmt.Errorf("failed to check for existing movie: %w", err)
		}
		if exists {
			config.GlobalLogger.Trace().Str("path", filePath).Msg("Movie already exists for file")
			return nil
		}
	} else {
		mediaFilePath, err = s.FindMoviePlaylistInBlurayFolder(filePath)
		if err != nil {
			return fmt.Errorf("failed to find media playlist in Bluray folder: %w", err)
		}
		config.GlobalLogger.Debug().Str("path", mediaFilePath).Msg("Detected Bluray folder, using media playlist")
	}

	// Extract media info
	mediaInfo, fileSize, duration, err = s.extractor.Extract(mediaFilePath)
	if err != nil {
		config.GlobalLogger.Warn().Err(err).Str("path", mediaFilePath).Msg("Mediainfo extraction failed")
		// Continue with minimal info
		mediaInfo = &models.MediaInfo{
			VideoTracks:    []models.VideoTrack{},
			AudioTracks:    []models.AudioTrack{},
			SubtitleTracks: []models.SubtitleTrack{},
		}
	}

	// If it's a Bluray folder, extracts file path and size from the source
	if isBluray {
		mediaFilePath = filepath.Join(filePath, "STREAM", mediaInfo.VideoTracks[0].Source)

		// Check if movie already exists by file path
		exists, err := repository.MovieExistsByFilePath(s.db, mediaFilePath)
		if err != nil {
			return fmt.Errorf("failed to check for existing movie: %w", err)
		}
		if exists {
			config.GlobalLogger.Trace().Str("path", mediaFilePath).Msg("Movie already exists for file")
			return nil
		}

		_, fileSize, _, err = s.extractor.Extract(mediaFilePath)
		if err != nil {
			config.GlobalLogger.Warn().Err(err).Str("path", mediaFilePath).Msg("Mediainfo extraction failed for media file in Bluray folder")
			fileSize = 0
		}
	}

	movie := &models.Movie{
		Title:     parsed.Title,
		Year:      parsed.Year,
		Duration:  duration / 60, // Convert seconds to minutes
		Status:    "available",
		FileSize:  fileSize,
		FilePath:  mediaFilePath,
		Container: GetContainer(mediaFilePath),
		DateAdded: time.Now().Format(time.RFC3339),
		MediaInfo: mediaInfo,
	}

	// Try to enrich with TMDB metadata
	if err := s.tmdb.EnrichMovie(movie); err != nil {
		config.GlobalLogger.Warn().Err(err).Str("title", parsed.Title).Msg("TMDB enrichment failed")
		// Continue without TMDB data
	}

	// Insert into database
	_, err = repository.InsertMovie(s.db, movie)
	if err != nil {
		return fmt.Errorf("failed to insert movie: %w", err)
	}

	result.MoviesAdded++
	config.GlobalLogger.Info().Str("title", movie.Title).Int("year", movie.Year).Msg("Added movie")
	return nil
}

func (s *Scanner) FindMoviePlaylistInBlurayFolder(blurayFolderPath string) (string, error) {
	playlistFolder := filepath.Join(blurayFolderPath, "PLAYLIST")
	files, err := os.ReadDir(playlistFolder)
	if err != nil {
		return "", fmt.Errorf("failed to read PLAYLIST folder: %w", err)
	}

	// Look for the .mpls having the longest duration (heuristic for main movie playlist)
	var longestPlaylist string
	var longestDuration int

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".mpls") {
			continue
		}

		playlistPath := filepath.Join(playlistFolder, file.Name())

		_, _, duration, err := s.extractor.Extract(playlistPath)

		if err != nil {
			config.GlobalLogger.Warn().Err(err).Str("path", playlistPath).Msg("Failed to get duration for playlist")
			continue
		}

		if duration > longestDuration {
			longestDuration = duration
			longestPlaylist = playlistPath
		}
	}

	if longestPlaylist == "" {
		return "", fmt.Errorf("no valid playlist found in Bluray folder")
	}

	return longestPlaylist, nil
}

func slugify(title string) string {
	slug := title

	// Normalize accented characters
	slug = strings.ToValidUTF8(slug, "")

	// Remove accents
	slug = strings.Map(func(r rune) rune {
		if r >= 0x0300 && r <= 0x036f {
			return -1
		}
		return r
	}, slug)

	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Trim whitespace
	slug = strings.TrimSpace(slug)

	// Replace non-alphanumeric characters with hyphens
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, slug)

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// processEpisode handles a TV episode file
func (s *Scanner) processEpisode(filePath string, parsed *ParsedFilename, result *models.ScanResult) error {
	// Check if episode already exists by file path
	exists, err := repository.EpisodeExistsByFilePath(s.db, filePath)
	if err != nil {
		return fmt.Errorf("failed to check for existing episode: %w", err)
	}
	if exists {
		config.GlobalLogger.Trace().Str("path", filePath).Msg("Episode already exists for file")
		return nil
	}

	// Get current time
	processStart := time.Now()

	// Normalize title: trimmed and lowercased title concatenated with year
	normalizedTitle := fmt.Sprintf("%s (%d)", strings.ToLower(strings.TrimSpace(parsed.Title)), parsed.Year)

	// Check cache for series by normalized title first
	var series *models.Series

	if s.cache.seriesByTitle != nil {
		if cached, ok := s.cache.seriesByTitle[normalizedTitle]; ok {
			series = cached
			config.GlobalLogger.Debug().Str("title", parsed.Title).Int("year", parsed.Year).Msg("Series found in cache")
		} else if ferr, failed := s.cache.failedSeriesByTitle[normalizedTitle]; failed {
			config.GlobalLogger.Debug().Str("title", parsed.Title).Int("year", parsed.Year).Err(err).Msg("Skipping enrichment due to previous failure")
			return ferr
		}
	}

	// If not in cache, lookup in database
	if series == nil {
		config.GlobalLogger.Debug().Str("title", parsed.Title).Int("year", parsed.Year).Msg("Looking up series in database")
		series, err = repository.GetSeriesByTitleAndYear(s.db, parsed.Title, parsed.Year)
		if err != nil {
			s.cache.failedSeriesByTitle[normalizedTitle] = err
			return fmt.Errorf("failed to lookup series: %w", err)
		} else if series != nil {
			config.GlobalLogger.Debug().Str("title", series.Title).Int64("tmdbID", series.TMDBId).Int64("tvdbID", series.TVDBId).Msg("Series found in database")

			// Try to enrich with TMDB metadata
			if err := s.tmdb.EnrichSeries(series); err != nil {
				config.GlobalLogger.Warn().Err(err).Str("title", parsed.Title).Msg("TMDB enrichment failed")
				s.cache.failedSeriesByTitle[normalizedTitle] = err
			}

			s.cache.seriesByTitle[normalizedTitle] = series
			config.GlobalLogger.Debug().Str("title", series.Title).Int("year", series.YearStart).Msg("Added series to cache")
		}
	}

	var seriesID int64
	var seriesTMDBID int
	var seriesTVDBID int

	if series == nil {
		config.GlobalLogger.Info().Str("title", parsed.Title).Int("year", parsed.Year).Msg("Series not found in database, creating new series")

		// Create new series
		newSeries := &models.Series{
			Title:     parsed.Title,
			YearStart: parsed.Year,
			Slug:      slugify(parsed.Title),
			Status:    "ongoing",
			DateAdded: time.Now().Format(time.RFC3339),
		}

		// Try to enrich with TMDB metadata
		if err := s.tmdb.EnrichSeries(newSeries); err != nil {
			config.GlobalLogger.Warn().Err(err).Str("title", parsed.Title).Msg("TMDB enrichment failed")
			s.cache.failedSeriesByTitle[normalizedTitle] = err
		}

		// Check if series with same TMDB ID already exists (prevents duplicates)
		if newSeries.TMDBId > 0 {
			config.GlobalLogger.Debug().Int64("tmdbID", newSeries.TMDBId).Msg("Checking for existing series with TMDB ID")
			existingSeries, err := repository.GetSeriesByTMDBId(s.db, newSeries.TMDBId)
			if err != nil {
				s.cache.failedSeriesByTitle[normalizedTitle] = err
				return fmt.Errorf("failed to lookup series by TMDB ID: %w", err)
			}
			if existingSeries != nil {
				// Series already exists, reuse it
				seriesID = existingSeries.ID
				seriesTMDBID = int(existingSeries.TMDBId)

				// Affect TVDB ID from existing series if ID is positive, otherwise use from new series (which may be 0 if enrichment failed)
				if existingSeries.TVDBId > 0 {
					seriesTVDBID = int(existingSeries.TVDBId)
				} else {
					seriesTVDBID = int(newSeries.TVDBId)
				}

				config.GlobalLogger.Info().Str("title", existingSeries.Title).Int("tmdbID", seriesTMDBID).Int("tvdbID", seriesTVDBID).Msg("Found existing series")
				s.cache.seriesByTitle[normalizedTitle] = existingSeries
				config.GlobalLogger.Debug().Str("title", existingSeries.Title).Msg("Added series to cache")
				series = existingSeries
				// Skip the InsertSeries step below
			} else {
				// New series, insert it
				seriesID, err = repository.InsertSeries(s.db, newSeries)
				if err != nil {
					s.cache.failedSeriesByTitle[normalizedTitle] = err
					return fmt.Errorf("failed to insert series: %w", err)
				}
				newSeries.ID = seriesID // Update the series object with the DB-generated ID
				seriesTMDBID = int(newSeries.TMDBId)
				seriesTVDBID = int(newSeries.TVDBId)
				config.GlobalLogger.Info().Str("title", newSeries.Title).Int("tmdbID", seriesTMDBID).Int("tvdbID", seriesTVDBID).Msg("Added series")
				s.cache.seriesByTitle[normalizedTitle] = newSeries
				config.GlobalLogger.Debug().Str("title", newSeries.Title).Msg("Added series to cache")
				series = newSeries
			}
		} else {
			config.GlobalLogger.Warn().Str("title", parsed.Title).Msg("No TMDB ID found, inserting without TMDB enrichment")
			// No TMDB ID, insert new series anyway
			seriesID, err = repository.InsertSeries(s.db, newSeries)
			if err != nil {
				s.cache.failedSeriesByTitle[normalizedTitle] = err
				return fmt.Errorf("failed to insert series: %w", err)
			}
			newSeries.ID = seriesID // Update the series object with the DB-generated ID
			seriesTMDBID = int(newSeries.TMDBId)
			seriesTVDBID = int(newSeries.TVDBId)
			config.GlobalLogger.Info().Str("title", newSeries.Title).Int("tmdbID", seriesTMDBID).Int("tvdbID", seriesTVDBID).Msg("Added series")
			s.cache.seriesByTitle[normalizedTitle] = newSeries
			config.GlobalLogger.Debug().Str("title", newSeries.Title).Msg("Added series to cache")
			series = newSeries
		}

		// Pre-fetch all series data if TVDB ID is available (bulk optimization)
		if seriesTVDBID > 0 {
			config.GlobalLogger.Debug().Int("tvdbID", seriesTVDBID).Msg("Pre-fetching series data after creating new series")
			if err := s.preFetchSeriesData(seriesTVDBID); err != nil {
				config.GlobalLogger.Warn().Err(err).Int("tvdbID", seriesTVDBID).Msg("Failed to pre-fetch series data")
				// Continue anyway - we can still process episodes without bulk data
			} else {
				// Update series with artwork from cached extended data if available
				if seriesExtended, exists := s.cache.episodesByTVDBId[seriesTVDBID]; exists {
					series.TVDBId = int64(seriesTVDBID)    // Ensure TVDB ID is set on series from cache
					series.Slug = seriesExtended.Data.Slug // Update slug from TVDB data for better URL compatibility
					if seriesExtended.Data.Image != "" {
						series.Poster = &seriesExtended.Data.Image
						if err := repository.UpdateSeries(s.db, series); err != nil {
							config.GlobalLogger.Warn().Err(err).Msg("Failed to update series artwork from cached data")
						} else {
							config.GlobalLogger.Info().Str("title", series.Title).Msg("Updated series artwork from cached data")
						}
					}
				}

				config.GlobalLogger.Info().Int("tvdbID", seriesTVDBID).Msg("Successfully pre-fetched series data")
				// Create seasons from cached extended series data
				if seriesExtended, exists := s.cache.episodesByTVDBId[seriesTVDBID]; exists {
					// for _, season := range seriesExtended.Data.Seasons {
					// 	// Only create "Aired Order" seasons (type.id == 1)
					// 	if season.ID > 0 && season.Number > 0 && season.Type.ID == 1 {
					// 		log.Printf("Creating season %d for series '%s' from cached data", season.Number, series.Title)
					// 		_, err := repository.GetOrCreateSeason(s.db, seriesID, season.Number)
					// 		if err != nil {
					// 			log.Printf("Warning: Failed to create season %d: %v", season.Number, err)
					// 		}
					// 	}
					// }
					var seasons = make(map[int]bool) // Map to track created seasons and avoid duplicates

					for _, episode := range seriesExtended.Data.Episodes {
						// Only create "Aired Order" seasons (type.id == 1)
						if episode.SeasonNumber > 0 {
							seasons[episode.SeasonNumber] = true
						}
					}

					for seasonNum := range seasons {
						config.GlobalLogger.Info().Int("season", seasonNum).Str("series", series.Title).Msg("Creating season from cached data")
						_, err := repository.GetOrCreateSeason(s.db, seriesID, seasonNum)
						if err != nil {
							config.GlobalLogger.Error().Err(err).Int("season", seasonNum).Msg("Failed to create season")
							// Continue creating other seasons even if one fails
						}
					}
				}
			}
		}
	} else {
		config.GlobalLogger.Debug().Str("title", series.Title).Int64("id", series.ID).Msg("Using existing series")
		seriesID = series.ID
		seriesTMDBID = int(series.TMDBId)
		seriesTVDBID = int(series.TVDBId)

		// Pre-fetch series data if not already cached and TVDB ID is available
		if seriesTVDBID > 0 {
			if _, exists := s.cache.episodesByTVDBId[seriesTVDBID]; !exists {
				config.GlobalLogger.Info().Int("tvdbID", seriesTVDBID).Msg("Pre-fetching series data")
				if err := s.preFetchSeriesData(seriesTVDBID); err != nil {
					config.GlobalLogger.Warn().Err(err).Int("tvdbID", seriesTVDBID).Msg("Failed to pre-fetch series data")
				} else {
					// Update series with artwork from cached extended data if available
					if seriesExtended, exists := s.cache.episodesByTVDBId[seriesTVDBID]; exists {
						series.TVDBId = int64(seriesTVDBID)    // Ensure TVDB ID is set on series from cache
						series.Slug = seriesExtended.Data.Slug // Update slug from TVDB data for better URL compatibility
						if seriesExtended.Data.Image != "" {
							series.Poster = &seriesExtended.Data.Image
							if err := repository.UpdateSeries(s.db, series); err != nil {
								config.GlobalLogger.Warn().Err(err).Msg("Failed to update series artwork from cached data")
							} else {
								config.GlobalLogger.Info().Str("title", series.Title).Msg("Updated series artwork from cached data")
							}
						}
					}
				}
			}
		}
	}

	// Extract media info
	mediaInfo, fileSize, duration, err := s.extractor.Extract(filePath)
	if err != nil {
		config.GlobalLogger.Warn().Err(err).Str("filePath", filePath).Msg("Mediainfo extraction failed")
		// Continue with minimal info
		mediaInfo = &models.MediaInfo{
			VideoTracks:    []models.VideoTrack{},
			AudioTracks:    []models.AudioTrack{},
			SubtitleTracks: []models.SubtitleTrack{},
		}
	}

	// Create episode
	episode := &models.Episode{
		SeriesID:   seriesID,
		SeasonNum:  parsed.Season,
		EpisodeNum: parsed.Episode,
		Duration:   duration, // Already in seconds
		Status:     "available",
		FileSize:   fileSize,
		FilePath:   filePath,
		DateAdded:  time.Now().Format(time.RFC3339),
		MediaInfo:  mediaInfo,
	}

	// Try to enrich episode from cache first (TVDB bulk data), fallback to TMDB if needed
	if seriesTVDBID > 0 {
		s.enrichEpisodeFromCache(episode, seriesTVDBID, parsed.Season, parsed.Episode)
	}

	// If no title from cache, try TMDB as fallback
	if episode.Title == "" && seriesTMDBID > 0 {
		if err := s.tmdb.EnrichEpisode(episode, seriesTMDBID); err != nil {
			config.GlobalLogger.Warn().Err(err).Msg("TMDB episode enrichment failed")
		}
	}

	// If still no title, create a default title
	if episode.Title == "" {
		episode.Title = fmt.Sprintf("Episode %d", parsed.Episode)
	}

	// Ensure season exists
	_, err = repository.GetOrCreateSeason(s.db, seriesID, parsed.Season)
	if err != nil {
		config.GlobalLogger.Error().Err(err).Msg("Failed to create season")
	}

	// Check if episode already exists
	existingEpisode, err := repository.GetEpisodeBySeriesSeasonEpisode(s.db, seriesID, parsed.Season, parsed.Episode)
	if err != nil {
		return fmt.Errorf("failed to lookup episode: %w", err)
	}

	if existingEpisode != nil {
		// Episode already exists - update if file path or details changed
		existingEpisode.Title = episode.Title
		existingEpisode.Duration = episode.Duration
		existingEpisode.Status = "available"
		existingEpisode.FileSize = episode.FileSize
		existingEpisode.FilePath = filePath

		if err := repository.UpdateEpisode(s.db, existingEpisode); err != nil {
			return fmt.Errorf("failed to update episode: %w", err)
		}
		config.GlobalLogger.Info().Str("series", parsed.Title).Int("season", parsed.Season).Int("episode", parsed.Episode).Str("title", existingEpisode.Title).Msg("Updated episode")
	} else {
		// New episode - insert it
		_, err = repository.InsertEpisode(s.db, episode)
		if err != nil {
			return fmt.Errorf("failed to insert episode: %w", err)
		}
		config.GlobalLogger.Info().Str("series", parsed.Title).Int("season", parsed.Season).Int("episode", parsed.Episode).Str("title", episode.Title).Msg("Added episode")
		result.EpisodesAdded++
	}

	// Update series counts
	repository.UpdateSeriesCounts(s.db, seriesID)

	processDuration := time.Since(processStart)
	config.GlobalLogger.Debug().Int64("duration_ms", processDuration.Milliseconds()).Str("series", parsed.Title).Int("season", parsed.Season).Int("episode", parsed.Episode).Str("title", episode.Title).Msg("Processed episode")

	return nil
}

// GetStatus returns the current scan status
func (s *Scanner) GetStatus() (*models.ScanStatus, error) {
	return repository.GetScanStatus(s.db)
}
