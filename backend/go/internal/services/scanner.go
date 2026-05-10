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

// Scanner handles media library scanning
type Scanner struct {
	db        *sql.DB
	config    *config.Config
	extractor *Extractor
	tmdb      *TMDBClient
	tv        *TVClient
	running   bool
	stopChan  chan struct{}
	mu        sync.Mutex
}

// NewScanner creates a new scanner service
func NewScanner(db *sql.DB, cfg *config.Config) *Scanner {
	return &Scanner{
		db:        db,
		config:    cfg,
		extractor: NewExtractor(cfg.MediainfoPath, cfg.ScanTimeout),
		tmdb:      NewTMDBClient(cfg.TMDBAPIKey),
		tv:        NewTVClient(cfg.TMDBAPIKey), // Uses TMDB for TV shows
		stopChan:  make(chan struct{}),
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
	}()

	result := &models.ScanResult{
		Errors: []string{},
	}

	// Update scan status to running
	status := &models.ScanStatus{
		Status:    "running",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	if err := repository.UpdateScanStatus(s.db, status); err != nil {
		log.Printf("Failed to update scan status: %v", err)
	}

	// Collect all media files
	var files []string
	for _, libPath := range s.config.MediaLibraryPaths {
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

			// Skip hidden directories
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "@") {
					return fs.SkipDir
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

	// Process each file sequentially
	for i, filePath := range files {
		// Check for stop signal
		select {
		case <-s.stopChan:
			status.Status = "stopped"
			status.CompletedAt = time.Now().Format(time.RFC3339)
			status.ErrorMessage = "Scan stopped by user"
			repository.UpdateScanStatus(s.db, status)
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

	log.Printf("Scan completed: %d files processed, %d movies added, %d episodes added, %d errors",
		result.FilesProcessed, result.MoviesAdded, result.EpisodesAdded, len(result.Errors))

	return result, nil
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
		log.Printf("Movie already exists for file: %s, skipping", filePath)
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
	// Check if series exists, create if not
	series, err := repository.GetSeriesByTitle(s.db, parsed.Title)
	if err != nil {
		return fmt.Errorf("failed to lookup series: %w", err)
	}

	var seriesID int64
	var seriesTMDBID int

	if series == nil {
		// Create new series
		newSeries := &models.Series{
			Title:     parsed.Title,
			Status:    "ongoing",
			DateAdded: time.Now().Format(time.RFC3339),
		}

		// Try to enrich with TMDB metadata
		if err := s.tv.EnrichSeries(newSeries); err != nil {
			log.Printf("TMDB TV enrichment failed for %s: %v", parsed.Title, err)
		}

		// Check if series with same TVDB ID already exists (prevents duplicates)
		if newSeries.TVDBId > 0 {
			existingSeries, err := repository.GetSeriesByTVDBId(s.db, newSeries.TVDBId)
			if err != nil {
				return fmt.Errorf("failed to lookup series by TVDB ID: %w", err)
			}
			if existingSeries != nil {
				// Series already exists, reuse it
				seriesID = existingSeries.ID
				seriesTMDBID = int(existingSeries.TVDBId)
				log.Printf("Found existing series: %s (TVDB ID: %d)", existingSeries.Title, newSeries.TVDBId)
				// Skip the InsertSeries step below
			} else {
				// New series, insert it
				seriesID, err = repository.InsertSeries(s.db, newSeries)
				if err != nil {
					return fmt.Errorf("failed to insert series: %w", err)
				}
				seriesTMDBID = int(newSeries.TVDBId)
				log.Printf("Added series: %s (TVDB ID: %d)", newSeries.Title, newSeries.TVDBId)
			}
		} else {
			// No TVDB ID, insert new series anyway
			seriesID, err = repository.InsertSeries(s.db, newSeries)
			if err != nil {
				return fmt.Errorf("failed to insert series: %w", err)
			}
			seriesTMDBID = int(newSeries.TVDBId)
			log.Printf("Added series: %s (no TVDB ID)", newSeries.Title)
		}
	} else {
		seriesID = series.ID
		seriesTMDBID = int(series.TVDBId)
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

	// Try to enrich episode with TMDB metadata
	if err := s.tv.EnrichEpisode(episode, seriesTMDBID); err != nil {
		log.Printf("TMDB episode enrichment failed: %v", err)
	}

	// If no title from TMDB, create a default title
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
