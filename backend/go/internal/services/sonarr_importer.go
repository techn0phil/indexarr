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

// SonarrImporter handles importing series from Sonarr API
type SonarrImporter struct {
	db          *sql.DB
	config      *config.Config
	client      *SonarrClient
	extractor   *Extractor
	broadcaster *Broadcaster
	pathFrom    string // Path mapping: from (Sonarr path prefix)
	pathTo      string // Path mapping: to (local path prefix)
	running     bool
	stopChan    chan struct{}
	mu          sync.Mutex
}

// NewSonarrImporter creates a new Sonarr importer
func NewSonarrImporter(db *sql.DB, cfg *config.Config, client *SonarrClient, broadcaster *Broadcaster) *SonarrImporter {
	pathFrom, pathTo := config.ParsePathMapping(cfg.SonarrPathMapping)
	return &SonarrImporter{
		db:          db,
		config:      cfg,
		client:      client,
		extractor:   NewExtractor(cfg.MediainfoPath, cfg.ScanTimeout),
		broadcaster: broadcaster,
		pathFrom:    pathFrom,
		pathTo:      pathTo,
		stopChan:    make(chan struct{}),
	}
}

// mapPath applies path mapping to convert Sonarr paths to local paths
func (si *SonarrImporter) mapPath(sonarrPath string) string {
	if si.pathFrom == "" || si.pathTo == "" {
		return sonarrPath
	}
	if strings.HasPrefix(sonarrPath, si.pathFrom) {
		return si.pathTo + sonarrPath[len(si.pathFrom):]
	}
	return sonarrPath
}

// IsRunning returns whether an import is currently in progress
func (si *SonarrImporter) IsRunning() bool {
	si.mu.Lock()
	defer si.mu.Unlock()
	return si.running
}

// Stop signals the importer to stop
func (si *SonarrImporter) Stop() {
	si.mu.Lock()
	if si.running {
		close(si.stopChan)
	}
	si.mu.Unlock()
}

// Import performs a full sync from Sonarr
func (si *SonarrImporter) Import() (*models.ScanResult, error) {
	si.mu.Lock()
	if si.running {
		si.mu.Unlock()
		return nil, fmt.Errorf("import already in progress")
	}
	si.running = true
	si.stopChan = make(chan struct{})
	si.mu.Unlock()

	log.Println("Starting Sonarr import")
	start := time.Now()

	defer func() {
		si.mu.Lock()
		si.running = false
		select {
		case <-si.stopChan:
			// Already closed
		default:
			close(si.stopChan)
		}
		si.mu.Unlock()
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
	if err := repository.UpdateScanStatus(si.db, status); err != nil {
		log.Printf("Failed to update scan status: %v", err)
	}

	// Fetch all series from Sonarr
	log.Println("Fetching series from Sonarr...")
	sonarrSeriesList, err := si.client.GetSeries()
	if err != nil {
		status.Status = "failed"
		status.ErrorMessage = err.Error()
		repository.UpdateScanStatus(si.db, status)
		return nil, fmt.Errorf("failed to fetch series from Sonarr: %w", err)
	}

	// Count total episodes with files for progress tracking
	totalEpisodesWithFiles := 0
	for _, ss := range sonarrSeriesList {
		totalEpisodesWithFiles += ss.Statistics.EpisodeFileCount
	}

	result.FilesFound = totalEpisodesWithFiles
	log.Printf("Found %d series with %d episode files in Sonarr", len(sonarrSeriesList), totalEpisodesWithFiles)

	// Update status with count
	status.FilesFound = totalEpisodesWithFiles
	repository.UpdateScanStatus(si.db, status)

	// Broadcast scan start
	if si.broadcaster != nil {
		si.broadcaster.BroadcastScanStart(totalEpisodesWithFiles, status.StartedAt)
	}

	// Track Sonarr IDs we've seen (for deletion detection)
	seenSonarrIds := make(map[int64]bool)

	// Process each series
	for _, ss := range sonarrSeriesList {
		// Check for stop signal
		select {
		case <-si.stopChan:
			status.Status = "stopped"
			status.CompletedAt = time.Now().Format(time.RFC3339)
			status.ErrorMessage = "Import stopped by user"
			repository.UpdateScanStatus(si.db, status)
			if si.broadcaster != nil {
				si.broadcaster.BroadcastScanStopped()
			}
			return result, fmt.Errorf("import stopped by user")
		default:
		}

		if err := si.processSonarrSeries(&ss, result, status); err != nil {
			log.Printf("Error processing series '%s': %v", ss.Title, err)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", ss.Title, err))
		} else {
			seenSonarrIds[int64(ss.ID)] = true
		}

		// Broadcast progress periodically
		if si.broadcaster != nil {
			si.broadcaster.BroadcastScanProgress(result.FilesProcessed, result.FilesFound)
		}
		status.FilesProcessed = result.FilesProcessed
		repository.UpdateScanStatus(si.db, status)
	}

	// Full sync: Remove series that are no longer in Sonarr
	deletedCount, err := si.removeStaleSeries(seenSonarrIds)
	if err != nil {
		log.Printf("Warning: Failed to remove stale series: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove stale series: %v", err))
	} else if deletedCount > 0 {
		log.Printf("Removed %d series no longer in Sonarr", deletedCount)
	}

	// Update status to completed
	status.Status = "completed"
	status.CompletedAt = time.Now().Format(time.RFC3339)
	status.FilesProcessed = result.FilesProcessed
	if len(result.Errors) > 0 {
		status.ErrorMessage = fmt.Sprintf("%d errors during import", len(result.Errors))
	}
	repository.UpdateScanStatus(si.db, status)

	// Broadcast completion
	if si.broadcaster != nil {
		si.broadcaster.BroadcastScanComplete(result.FilesProcessed, 0, result.EpisodesAdded)
	}

	duration := time.Since(start)
	log.Printf("Sonarr import completed in %v - %d episodes processed, %d added, %d series removed, %d errors",
		duration.Round(time.Second), result.FilesProcessed, result.EpisodesAdded, deletedCount, len(result.Errors))

	// Print the top 100 errors if there are any
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

// processSonarrSeries processes a single series from Sonarr
func (si *SonarrImporter) processSonarrSeries(ss *SonarrSeries, result *models.ScanResult, status *models.ScanStatus) error {
	// Map Sonarr series to our model
	series := si.mapSonarrSeries(ss)

	// Check if series already exists by Sonarr ID
	existing, err := repository.GetSeriesBySonarrID(si.db, series.SonarrID)
	if err != nil {
		return fmt.Errorf("failed to check existing series: %w", err)
	}

	var seriesID int64
	if existing != nil {
		// Update existing series
		series.ID = existing.ID
		seriesID = existing.ID
		if err := repository.UpdateSeries(si.db, series); err != nil {
			return fmt.Errorf("failed to update series: %w", err)
		}
		log.Printf("Updated series: %s (%d)", series.Title, series.YearStart)
	} else {
		// Insert new series
		id, err := repository.InsertSeries(si.db, series)
		if err != nil {
			return fmt.Errorf("failed to insert series: %w", err)
		}
		seriesID = id
		log.Printf("Added series: %s (%d)", series.Title, series.YearStart)
	}

	// Fetch episodes from Sonarr
	episodes, err := si.client.GetEpisodes(ss.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch episodes: %w", err)
	}

	// Process each episode
	for _, se := range episodes {
		// Check for stop signal
		select {
		case <-si.stopChan:
			return fmt.Errorf("import stopped by user")
		default:
		}

		if err := si.processSonarrEpisode(seriesID, &se, result); err != nil {
			log.Printf("Error processing episode S%02dE%02d of '%s': %v", se.SeasonNumber, se.EpisodeNumber, ss.Title, err)
			result.Errors = append(result.Errors, fmt.Sprintf("%s S%02dE%02d: %v", ss.Title, se.SeasonNumber, se.EpisodeNumber, err))
		}

		// Count only episodes with files as processed
		if se.HasFile {
			result.FilesProcessed++
		}
	}

	// Update series counts
	if err := repository.UpdateSeriesCounts(si.db, seriesID); err != nil {
		log.Printf("Failed to update series counts: %v", err)
	}

	return nil
}

// processSonarrEpisode processes a single episode from Sonarr
func (si *SonarrImporter) processSonarrEpisode(seriesID int64, se *SonarrEpisode, result *models.ScanResult) error {
	// Ensure season exists
	_, err := repository.GetOrCreateSeason(si.db, seriesID, se.SeasonNumber)
	if err != nil {
		return fmt.Errorf("failed to create season: %w", err)
	}

	// Map episode
	episode := si.mapSonarrEpisode(seriesID, se)

	// Extract mediainfo if episode has a file
	if se.HasFile && se.EpisodeFile != nil && se.EpisodeFile.Path != "" {
		localPath := si.mapPath(se.EpisodeFile.Path)
		mediaInfo, fileSize, duration, err := si.extractor.Extract(localPath)
		if err != nil {
			log.Printf("Mediainfo extraction failed for %s S%02dE%02d: %v", se.Title, se.SeasonNumber, se.EpisodeNumber, err)
			// Continue with Sonarr's file size
			episode.FileSize = se.EpisodeFile.Size
		} else {
			episode.MediaInfo = mediaInfo
			episode.FileSize = fileSize
			if duration > 0 {
				episode.Duration = duration // Already in seconds
			}
		}
		episode.FilePath = localPath
	}

	// Check if episode already exists
	existing, err := repository.GetEpisodeBySeriesSeasonEpisode(si.db, seriesID, se.SeasonNumber, se.EpisodeNumber)
	if err != nil {
		return fmt.Errorf("failed to check existing episode: %w", err)
	}

	if existing != nil {
		// Update existing episode
		episode.ID = existing.ID
		if err := repository.UpdateEpisode(si.db, episode); err != nil {
			return fmt.Errorf("failed to update episode: %w", err)
		}
	} else {
		// Insert new episode
		_, err := repository.InsertEpisode(si.db, episode)
		if err != nil {
			return fmt.Errorf("failed to insert episode: %w", err)
		}
		result.EpisodesAdded++
	}

	return nil
}

// mapSonarrSeries converts a SonarrSeries to our internal Series model
func (si *SonarrImporter) mapSonarrSeries(ss *SonarrSeries) *models.Series {
	series := &models.Series{
		Title:        ss.Title,
		Slug:         slugify(ss.Title),
		YearStart:    ss.Year,
		SeasonCount:  ss.SeasonCount,
		EpisodeCount: ss.Statistics.EpisodeCount,
		Synopsis:     ss.Overview,
		Status:       si.mapSeriesStatus(ss.Status),
		FileSize:     ss.Statistics.SizeOnDisk,
		TMDBId:       int64(ss.TmdbId),
		TVDBId:       int64(ss.TvdbId),
		IMDbId:       ss.ImdbId,
		SonarrID:     int64(ss.ID),
		TitleSlug:    ss.TitleSlug,
		DateAdded:    ss.Added,
	}

	// Genres (join array)
	if len(ss.Genres) > 0 {
		series.Genres = strings.Join(ss.Genres, ", ")
	}

	// Rating
	if ss.Ratings.Value > 0 {
		series.Rating = ss.Ratings.Value
	}

	// Poster URL
	posterURL := si.extractPosterURL(ss)
	if posterURL != "" {
		series.Poster = &posterURL
	}

	// Use Sonarr's added date, fallback to now
	if series.DateAdded == "" {
		series.DateAdded = time.Now().Format(time.RFC3339)
	}

	return series
}

// mapSonarrEpisode converts a SonarrEpisode to our internal Episode model
func (si *SonarrImporter) mapSonarrEpisode(seriesID int64, se *SonarrEpisode) *models.Episode {
	episode := &models.Episode{
		SeriesID:   seriesID,
		SeasonNum:  se.SeasonNumber,
		EpisodeNum: se.EpisodeNumber,
		Title:      se.Title,
		Duration:   se.Runtime * 60, // Convert minutes to seconds
		DateAdded:  time.Now().Format(time.RFC3339),
	}

	// Set status based on file presence
	if se.HasFile {
		episode.Status = "available"
		if se.EpisodeFile != nil {
			episode.FileSize = se.EpisodeFile.Size
			episode.FilePath = se.EpisodeFile.Path
		}
	} else {
		episode.Status = "missing"
	}

	// Default title if empty
	if episode.Title == "" {
		episode.Title = fmt.Sprintf("Episode %d", se.EpisodeNumber)
	}

	return episode
}

// mapSeriesStatus converts Sonarr status to our status
func (si *SonarrImporter) mapSeriesStatus(sonarrStatus string) string {
	switch sonarrStatus {
	case "continuing":
		return "ongoing"
	case "ended":
		return "complete"
	default:
		return "ongoing"
	}
}

// extractPosterURL extracts the poster URL from Sonarr images
func (si *SonarrImporter) extractPosterURL(ss *SonarrSeries) string {
	for _, img := range ss.Images {
		if img.CoverType == "poster" {
			// Prefer remote URL, fallback to local URL
			if img.RemoteURL != "" {
				return img.RemoteURL
			}
			if img.URL != "" {
				// Local URL needs Sonarr base URL prepended
				return fmt.Sprintf("%s%s", si.client.baseURL, img.URL)
			}
		}
	}
	return ""
}

// removeStaleSeries removes series that exist in our DB but not in Sonarr
func (si *SonarrImporter) removeStaleSeries(seenSonarrIds map[int64]bool) (int, error) {
	// Get all Sonarr IDs from our database
	existingIds, err := repository.GetAllSeriesSonarrIDs(si.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get existing series IDs: %w", err)
	}

	var deletedCount int
	for _, sonarrId := range existingIds {
		if sonarrId == 0 {
			// Skip series without Sonarr ID (shouldn't happen with Sonarr imports)
			continue
		}
		if !seenSonarrIds[sonarrId] {
			// Series is in our DB but not in Sonarr - delete it
			if err := repository.DeleteSeriesBySonarrID(si.db, sonarrId); err != nil {
				log.Printf("Failed to delete series with Sonarr ID %d: %v", sonarrId, err)
			} else {
				deletedCount++
			}
		}
	}

	return deletedCount, nil
}

// ImportSeries imports/refreshes a single series by its database ID
func (si *SonarrImporter) ImportSeries(seriesID int64) (*models.ScanResult, error) {
	log.Printf("Starting single series refresh for ID: %d", seriesID)
	start := time.Now()

	result := &models.ScanResult{
		Errors: []string{},
	}

	// Get the series from our database to find its Sonarr ID
	series, err := repository.GetSeriesByID(si.db, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series: %w", err)
	}
	if series == nil {
		return nil, fmt.Errorf("series not found with ID: %d", seriesID)
	}

	// If series has no Sonarr ID, try to find by TVDB ID
	var sonarrSeries *SonarrSeries
	if series.SonarrID > 0 {
		sonarrSeries, err = si.client.GetSeriesByID(int(series.SonarrID))
	} else if series.TVDBId > 0 {
		sonarrSeries, err = si.client.GetSeriesByTVDBId(int(series.TVDBId))
	} else {
		return nil, fmt.Errorf("series has no Sonarr ID or TVDB ID")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch series from Sonarr: %w", err)
	}
	if sonarrSeries == nil {
		// Series no longer in Sonarr - delete it
		log.Printf("Series no longer in Sonarr, deleting: %s", series.Title)
		if err := repository.DeleteSeries(si.db, seriesID); err != nil {
			return nil, fmt.Errorf("failed to delete series: %w", err)
		}
		result.FilesProcessed = 1
		return result, nil
	}

	// Create a status struct for processSonarrSeries
	status := &models.ScanStatus{}

	// Process the series
	result.FilesFound = sonarrSeries.Statistics.EpisodeFileCount
	if err := si.processSonarrSeries(sonarrSeries, result, status); err != nil {
		return nil, fmt.Errorf("failed to process series: %w", err)
	}

	duration := time.Since(start)
	log.Printf("Series refresh completed in %d ms", duration.Milliseconds())

	return result, nil
}

// GetStatus returns the current import/scan status
func (si *SonarrImporter) GetStatus() (*models.ScanStatus, error) {
	return repository.GetScanStatus(si.db)
}

// slugify is reused from scanner.go for creating URL-safe slugs
// Note: This function is also defined in scanner.go - consider moving to a utils package
func slugifyTitle(title string) string {
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
