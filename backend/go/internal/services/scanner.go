package services

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
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
	seriesExtendedByTVDBId map[int]*TVDBSeriesExtended
	// All episodes by series TVDB ID (from bulk episodes endpoint)
	episodesByTVDBId map[int][]TVDBBulkEpisode
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
		tmdb:        NewTMDBClient(cfg.TMDBAPIKey),
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
	return s.ScanPaths(s.config.MediaLibraryPaths)
}

// ScanPaths performs a scan on specified paths (used for manual scans via API)
func (s *Scanner) ScanPaths(paths []string) (*models.ScanResult, error) {
	// Initialize per-scan cache for API call optimization
	s.cache = &scanCache{
		seriesByTitle:          make(map[string]*models.Series),
		seriesExtendedByTVDBId: make(map[int]*TVDBSeriesExtended),
		episodesByTVDBId:       make(map[int][]TVDBBulkEpisode),
		failedSeriesByTitle:    make(map[string]error),
	}

	log.Println("Starting scan")
	start := time.Now()

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil, fmt.Errorf("scan already in progress")
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		// Clear per-scan cache
		s.cache = nil
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
	if err := repository.UpdateScanStatus(s.db, status); err != nil {
		log.Printf("Failed to update scan status: %v", err)
	}

	// Collect all media files
	var files []string
	for _, libPath := range paths {
		if libPath == "" {
			continue
		}

		log.Printf("Scanning library path: %s", libPath)

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
				log.Printf("Error accessing path %s: %v", path, err)
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
	log.Printf("Found %d media files", result.FilesFound)

	// Update status with files found
	status.FilesFound = result.FilesFound
	repository.UpdateScanStatus(s.db, status)

	// Broadcast scan start to WebSocket clients
	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanStart(result.FilesFound, status.StartedAt)
	}

	// Process each file sequentially
	for i, filePath := range files {
		// Check for stop signal
		select {
		case <-s.stopChan:
			status.Status = "stopped"
			status.CompletedAt = time.Now().Format(time.RFC3339)
			status.ErrorMessage = "Scan stopped by user"
			repository.UpdateScanStatus(s.db, status)
			// Broadcast stopped event to WebSocket clients
			if s.broadcaster != nil {
				s.broadcaster.BroadcastScanStopped()
			}
			return result, fmt.Errorf("scan stopped by user")
		default:
		}

		if err := s.processFile(filePath, result); err != nil {
			log.Printf("Error processing %s: %v", filePath, err)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", filepath.Base(filePath), err))
		}

		result.FilesProcessed++

		// Update progress periodically
		if i%10 == 0 || i == len(files)-1 {
			status.FilesProcessed = result.FilesProcessed
			repository.UpdateScanStatus(s.db, status)
			// Broadcast progress to WebSocket clients
			if s.broadcaster != nil {
				s.broadcaster.BroadcastScanProgress(result.FilesProcessed, result.FilesFound)
			}
		}
	}

	// Update status to completed
	status.Status = "completed"
	status.CompletedAt = time.Now().Format(time.RFC3339)
	status.FilesProcessed = result.FilesProcessed
	if len(result.Errors) > 0 {
		status.ErrorMessage = fmt.Sprintf("%d errors during scan", len(result.Errors))
	}
	repository.UpdateScanStatus(s.db, status)

	// Broadcast completion to WebSocket clients
	if s.broadcaster != nil {
		s.broadcaster.BroadcastScanComplete(result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded)
	}

	duration := time.Since(start)
	log.Printf("Scan completed in %v - %d files processed, %d movies added, %d episodes added, %d errors",
		duration.Round(time.Second), result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded, len(result.Errors))

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

// ScanMovie scans a single movie (used for manual refresh via API)
func (s *Scanner) ScanMovie(movieID int64) (*models.ScanResult, error) {
	movie, err := repository.GetMovieByID(s.db, movieID)
	if err != nil {
		return nil, fmt.Errorf("movie not found: %w", err)
	}

	result, err := s.ScanPaths([]string{movie.FilePath})
	if err != nil {
		return nil, err
	}

	// Remove movie if it was deleted from disk
	if result.FilesProcessed == 0 {
		log.Printf("Movie file not found during refresh, deleting movie: %s", movie.FilePath)
		if err := repository.DeleteMovie(s.db, movieID); err != nil {
			log.Printf("Failed to delete movie: %v", err)
		}
	} else {
		// Extract media info again to update any changes (e.g. new audio tracks)
		mediaInfo, fileSize, duration, err := s.extractor.Extract(movie.FilePath)
		if err != nil {
			log.Printf("Mediainfo extraction failed during refresh for %s: %v", movie.Title, err)
		} else {
			movie.MediaInfo = mediaInfo
			movie.FileSize = fileSize
			movie.Duration = duration / 60 // Convert seconds to minutes
		}

		// Fetch metadata from TMDB again to update any changes
		if err := s.tmdb.EnrichMovie(movie); err != nil {
			log.Printf("TMDB enrichment failed during refresh for %s: %v", movie.Title, err)
		} else {
			// Update movie with new metadata
			if err := repository.UpdateMovie(s.db, movie); err != nil {
				log.Printf("Failed to update movie during refresh: %v", err)
			}
			log.Printf("Movie refreshed: %s (%d)", movie.Title, movie.Year)
			result.MoviesAdded = 1 // Count as "added" for refresh purposes
			result.FilesProcessed = 1
			result.FilesFound = 1
			result.Errors = []string{}
			result.EpisodesAdded = 0
			result.MoviesAdded = 1
		}
	}

	return result, nil
}

// ScanSeries scans a single series (used for manual refresh via API)
func (s *Scanner) ScanSeries(seriesID int64) (*models.ScanResult, error) {
	log.Printf("Starting series refresh for ID: %d", seriesID)
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

	log.Printf("Found series: %s", series.Title)

	// Step 2: Fetch all episodes to determine folders to scan
	episodes, err := repository.GetAllEpisodesForSeries(s.db, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", err)
	}

	// Extract unique folder paths from existing episodes
	folderPaths := s.findSeriesFolderPaths(episodes)
	log.Printf("Will scan %d folder(s) for series: %v", len(folderPaths), folderPaths)

	// Step 3: Scan folder paths to detect new episodes
	scanResult, err := s.ScanPaths(folderPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to scan series folders: %w", err)
	}

	result.FilesFound = scanResult.FilesFound

	// Step 4 & 5: Check for missing episodes and delete them from database
	episodesToDelete := []int64{}
	for _, episode := range episodes {
		// Check if file still exists
		if _, err := os.Stat(episode.FilePath); os.IsNotExist(err) {
			log.Printf("Episode file missing: %s (S%02dE%02d), marking for removal", episode.FilePath, episode.SeasonNum, episode.EpisodeNum)
			episodesToDelete = append(episodesToDelete, episode.ID)
		}
	}

	// Delete missing episodes
	for _, episodeID := range episodesToDelete {
		if err := repository.DeleteEpisode(s.db, episodeID); err != nil {
			errMsg := fmt.Sprintf("Failed to remove missing episode %d: %v", episodeID, err)
			log.Printf("%s", errMsg)
			result.Errors = append(result.Errors, errMsg)
		}
	}

	// Delete seasons that have no episodes left
	if err := repository.DeleteEmptySeasons(s.db, seriesID); err != nil {
		log.Printf("Failed to delete empty seasons: %v", err)
	}

	// Step 6: Check if series folder is completely missing
	if scanResult.FilesFound == 0 && len(episodesToDelete) == len(episodes) {
		log.Printf("All episodes missing from disk, deleting series: %s", series.Title)
		if err := repository.DeleteSeries(s.db, seriesID); err != nil {
			errMsg := fmt.Sprintf("Failed to delete series: %v", err)
			log.Printf("%s", errMsg)
			result.Errors = append(result.Errors, errMsg)
		}
		return result, nil
	}

	// Step 7: Extract media info for each episode again to catch any changes
	for _, episode := range episodes {
		mediaInfo, fileSize, duration, err := s.extractor.Extract(episode.FilePath)
		if err != nil {
			log.Printf("Mediainfo extraction failed for %s: %v", episode.FilePath, err)
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
			log.Printf("Failed to update episode during refresh: %v", err)
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to update episode S%02dE%02d: %v", episode.SeasonNum, episode.EpisodeNum, err))
		}
	}

	// Re-enrich series metadata from TMDB
	if err := s.tmdb.EnrichSeries(series); err != nil {
		log.Printf("TMDB enrichment failed during refresh for %s: %v", series.Title, err)
	}

	// Update series in database
	if err := repository.UpdateSeries(s.db, series); err != nil {
		log.Printf("Failed to update series during refresh: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update series: %v", err))
	}

	// Recalculate series counts
	if err := repository.UpdateSeriesCounts(s.db, seriesID); err != nil {
		log.Printf("Failed to update series counts: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update series counts: %v", err))
	}

	result.FilesProcessed = scanResult.FilesProcessed
	result.EpisodesAdded = scanResult.EpisodesAdded
	result.MoviesAdded = 0 // No movies in series refresh

	// Merge any errors from the scan
	result.Errors = append(result.Errors, scanResult.Errors...)

	duration := time.Since(start)
	log.Printf("Series refresh completed in %v - %d files processed, %d episodes added, %d episodes deleted, %d errors",
		duration.Round(time.Second), result.FilesProcessed, result.EpisodesAdded, len(episodesToDelete), len(result.Errors))

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
		log.Printf("Series episodes are in multiple folders, using common parent folder for scan: %s", commonParent)
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
	if _, exists := s.cache.seriesExtendedByTVDBId[tvdbID]; exists {
		log.Printf("[Cache] Series metadata already cached for TVDB ID %d", tvdbID)
		return nil
	}

	log.Printf("Pre-fetching all metadata for series TVDB ID %d...", tvdbID)

	// Fetch extended series metadata (includes seasons array)
	seriesExtended, err := s.tv.GetTVDetails(tvdbID)
	if err != nil {
		return fmt.Errorf("failed to fetch series extended metadata: %w", err)
	}

	s.cache.seriesExtendedByTVDBId[tvdbID] = seriesExtended
	log.Printf("[Cache] Successfully cached series metadata: %d seasons", len(seriesExtended.Data.Seasons))

	// Fetch all episodes in bulk
	allEpisodes, err := s.tv.GetAllEpisodes(tvdbID, "fra")
	if err != nil {
		return fmt.Errorf("failed to fetch bulk episodes: %w", err)
	}
	s.cache.episodesByTVDBId[tvdbID] = allEpisodes.Data.Episodes

	log.Printf("[Cache] Successfully cached episodes data: %d episodes, %d seasons",
		len(allEpisodes.Data.Episodes), len(seriesExtended.Data.Seasons))

	return nil
}

// enrichEpisodeFromCache populates episode metadata from cached bulk data (no API calls)
func (s *Scanner) enrichEpisodeFromCache(episode *models.Episode, seriesTVDBID int, seasonNum, episodeNum int) {
	// Get cached episodes for this series
	cachedEpisodes, exists := s.cache.episodesByTVDBId[seriesTVDBID]
	if !exists {
		log.Printf("Warning: No cached episodes found for series TVDB ID %d", seriesTVDBID)
		return
	}

	// Find matching episode by season and episode number
	for _, ep := range cachedEpisodes {
		if ep.SeasonNumber == seasonNum && ep.Number == episodeNum {
			episode.Title = ep.Name
			if episode.Duration == 0 && ep.Runtime > 0 {
				episode.Duration = ep.Runtime * 60 // Convert minutes to seconds
			}
			log.Printf("[Cache] Enriched episode S%02dE%02d: %s (runtime: %dm)",
				seasonNum, episodeNum, ep.Name, ep.Runtime)
			return
		}
	}

	// Episode not found in cache, use default title
	log.Printf("Warning: Episode S%02dE%02d not found in cached data for series %d", seasonNum, episodeNum, seriesTVDBID)
	episode.Title = fmt.Sprintf("Episode %d", episodeNum)
}

// processFile handles a single media file
func (s *Scanner) processFile(filePath string, result *models.ScanResult) error {
	// Parse filename
	parsed := ParseFilename(filePath)

	// Extract media info
	mediaInfo, fileSize, duration, err := s.extractor.Extract(filePath)
	if err != nil {
		log.Printf("Mediainfo extraction failed for %s: %v", filePath, err)
		// Continue with minimal info
		mediaInfo = &models.MediaInfo{
			VideoTracks:    []models.VideoTrack{},
			AudioTracks:    []models.AudioTrack{},
			SubtitleTracks: []models.SubtitleTrack{},
		}
	}

	if parsed.IsSeries {
		return s.processEpisode(filePath, parsed, mediaInfo, fileSize, duration, result)
	}
	return s.processMovie(filePath, parsed, mediaInfo, fileSize, duration, result)
}

// processMovie handles a movie file
func (s *Scanner) processMovie(filePath string, parsed *ParsedFilename, mediaInfo *models.MediaInfo, fileSize int64, duration int, result *models.ScanResult) error {
	// Check if movie already exists by file path
	exists, err := repository.MovieExistsByFilePath(s.db, filePath)
	if err != nil {
		return fmt.Errorf("failed to check for existing movie: %w", err)
	}
	if exists {
		log.Printf("Movie already exists for file: %s", filePath)
		return nil
	}

	movie := &models.Movie{
		Title:     parsed.Title,
		Year:      parsed.Year,
		Duration:  duration / 60, // Convert seconds to minutes
		Status:    "available",
		FileSize:  fileSize,
		FilePath:  filePath,
		Container: GetContainer(filePath),
		DateAdded: time.Now().Format(time.RFC3339),
		MediaInfo: mediaInfo,
	}

	// Try to enrich with TMDB metadata
	if err := s.tmdb.EnrichMovie(movie); err != nil {
		log.Printf("TMDB enrichment failed for %s: %v", parsed.Title, err)
		// Continue without TMDB data
	}

	// Insert into database
	_, err = repository.InsertMovie(s.db, movie)
	if err != nil {
		return fmt.Errorf("failed to insert movie: %w", err)
	}

	result.MoviesAdded++
	log.Printf("Added movie: %s (%d)", movie.Title, movie.Year)
	return nil
}

// processEpisode handles a TV episode file
func (s *Scanner) processEpisode(filePath string, parsed *ParsedFilename, mediaInfo *models.MediaInfo, fileSize int64, duration int, result *models.ScanResult) error {
	normalizedTitle := strings.ToLower(strings.TrimSpace(parsed.Title))

	// Check cache for series by normalized title first
	var series *models.Series
	var err error

	if s.cache.seriesByTitle != nil {
		if cached, ok := s.cache.seriesByTitle[normalizedTitle]; ok {
			series = cached
			log.Printf("[Cache] Series found in cache: %s", parsed.Title)
		} else if ferr, failed := s.cache.failedSeriesByTitle[normalizedTitle]; failed {
			log.Printf("[Cache] Skipping enrichment for series '%s' due to previous failure: %v", parsed.Title, ferr)
			return ferr
		}
	}

	// If not in cache, lookup in database
	if series == nil {
		log.Printf("Looking up series in database: %s", parsed.Title)
		series, err = repository.GetSeriesByTitle(s.db, parsed.Title)
		if err != nil {
			s.cache.failedSeriesByTitle[normalizedTitle] = err
			return fmt.Errorf("failed to lookup series: %w", err)
		} else if series != nil {
			log.Printf("Series found in database: %s (TMDB ID: %d, TVDB ID: %d)", series.Title, series.TMDBId, series.TVDBId)
			s.cache.seriesByTitle[normalizedTitle] = series
			log.Printf("[Cache] Added series to cache: %s", series.Title)
		}
	}

	var seriesID int64
	var seriesTMDBID int
	var seriesTVDBID int

	if series == nil {
		log.Printf("Series not found in database, creating new series: %s", parsed.Title)

		// Create new series
		newSeries := &models.Series{
			Title:     parsed.Title,
			Status:    "ongoing",
			DateAdded: time.Now().Format(time.RFC3339),
		}

		// Try to enrich with TMDB metadata
		if err := s.tmdb.EnrichSeries(newSeries); err != nil {
			log.Printf("TMDB enrichment failed for %s: %v", parsed.Title, err)
			s.cache.failedSeriesByTitle[normalizedTitle] = err
		}

		// Check if series with same TMDB ID already exists (prevents duplicates)
		if newSeries.TMDBId > 0 {
			log.Printf("Checking for existing series with TMDB ID %d", newSeries.TMDBId)
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

				log.Printf("Found existing series: %s (TMDB ID: %d, TVDB ID: %d)", existingSeries.Title, seriesTMDBID, seriesTVDBID)
				s.cache.seriesByTitle[normalizedTitle] = existingSeries
				log.Printf("[Cache] Added series to cache: %s", existingSeries.Title)
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
				log.Printf("Added series: %s (TMDB ID: %d, TVDB ID: %d)", newSeries.Title, seriesTMDBID, seriesTVDBID)
				s.cache.seriesByTitle[normalizedTitle] = newSeries
				log.Printf("[Cache] Added series to cache: %s", newSeries.Title)
				series = newSeries
			}
		} else {
			log.Printf("No TMDB ID found for series '%s', inserting without TMDB enrichment", parsed.Title)
			// No TMDB ID, insert new series anyway
			seriesID, err = repository.InsertSeries(s.db, newSeries)
			if err != nil {
				s.cache.failedSeriesByTitle[normalizedTitle] = err
				return fmt.Errorf("failed to insert series: %w", err)
			}
			newSeries.ID = seriesID // Update the series object with the DB-generated ID
			seriesTMDBID = int(newSeries.TMDBId)
			seriesTVDBID = int(newSeries.TVDBId)
			log.Printf("Added series: %s (TMDB ID: %d, TVDB ID: %d)", newSeries.Title, seriesTMDBID, seriesTVDBID)
			s.cache.seriesByTitle[normalizedTitle] = newSeries
			log.Printf("[Cache] Added series to cache: %s", newSeries.Title)
			series = newSeries
		}

		// Pre-fetch all series data if TVDB ID is available (bulk optimization)
		if seriesTVDBID > 0 {
			log.Printf("Pre-fetching series data for TVDB ID %d after creating new series...", seriesTVDBID)
			if err := s.preFetchSeriesData(seriesTVDBID); err != nil {
				log.Printf("Warning: Failed to pre-fetch series data for TVDB ID %d: %v", seriesTVDBID, err)
				// Continue anyway - we can still process episodes without bulk data
			} else {
				// Update series with artwork from cached extended data if available
				if seriesExtended, exists := s.cache.seriesExtendedByTVDBId[seriesTVDBID]; exists {
					series.TVDBId = int64(seriesTVDBID) // Ensure TVDB ID is set on series from cache
					if seriesExtended.Data.Image != "" {
						series.Poster = &seriesExtended.Data.Image
						if err := repository.UpdateSeries(s.db, series); err != nil {
							log.Printf("Warning: Failed to update series artwork from cached data: %v", err)
						} else {
							log.Printf("Updated series artwork from cached data for '%s'", series.Title)
						}
					}
				}

				log.Printf("Successfully pre-fetched series data for TVDB ID %d", seriesTVDBID)
				// Create seasons from cached extended series data
				if seriesExtended, exists := s.cache.seriesExtendedByTVDBId[seriesTVDBID]; exists {
					for _, season := range seriesExtended.Data.Seasons {
						// Only create "Aired Order" seasons (type.id == 1)
						if season.ID > 0 && season.Number > 0 && season.Type.ID == 1 {
							log.Printf("Creating season %d for series '%s' from cached data", season.Number, series.Title)
							_, err := repository.GetOrCreateSeason(s.db, seriesID, season.Number)
							if err != nil {
								log.Printf("Warning: Failed to create season %d: %v", season.Number, err)
							}
						}
					}
				}
			}
		}
	} else {
		log.Printf("Using existing series: %s (ID: %d)", series.Title, series.ID)
		seriesID = series.ID
		seriesTMDBID = int(series.TMDBId)
		seriesTVDBID = int(series.TVDBId)

		// Pre-fetch series data if not already cached and TVDB ID is available
		if seriesTVDBID > 0 {
			if _, exists := s.cache.episodesByTVDBId[seriesTVDBID]; !exists {
				log.Printf("Pre-fetching series data for TVDB ID %d...", seriesTVDBID)
				if err := s.preFetchSeriesData(seriesTVDBID); err != nil {
					log.Printf("Warning: Failed to pre-fetch series data for TVDB ID %d: %v", seriesTVDBID, err)
				} else {
					// Update series with artwork from cached extended data if available
					if seriesExtended, exists := s.cache.seriesExtendedByTVDBId[seriesTVDBID]; exists {
						series.TVDBId = int64(seriesTVDBID) // Ensure TVDB ID is set on series from cache
						if seriesExtended.Data.Image != "" {
							series.Poster = &seriesExtended.Data.Image
							if err := repository.UpdateSeries(s.db, series); err != nil {
								log.Printf("Warning: Failed to update series artwork from cached data: %v", err)
							} else {
								log.Printf("Updated series artwork from cached data for '%s'", series.Title)
							}
						}
					}
				}
			}
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
			log.Printf("TMDB episode enrichment failed: %v", err)
		}
	}

	// If still no title, create a default title
	if episode.Title == "" {
		episode.Title = fmt.Sprintf("Episode %d", parsed.Episode)
	}

	// Ensure season exists
	_, err = repository.GetOrCreateSeason(s.db, seriesID, parsed.Season)
	if err != nil {
		log.Printf("Failed to create season: %v", err)
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
		log.Printf("Updated episode: %s S%02dE%02d - %s", parsed.Title, parsed.Season, parsed.Episode, existingEpisode.Title)
	} else {
		// New episode - insert it
		_, err = repository.InsertEpisode(s.db, episode)
		if err != nil {
			return fmt.Errorf("failed to insert episode: %w", err)
		}
		log.Printf("Added episode: %s S%02dE%02d - %s", parsed.Title, parsed.Season, parsed.Episode, episode.Title)
		result.EpisodesAdded++
	}

	// Update series counts
	repository.UpdateSeriesCounts(s.db, seriesID)
	return nil
}

// GetStatus returns the current scan status
func (s *Scanner) GetStatus() (*models.ScanStatus, error) {
	return repository.GetScanStatus(s.db)
}
