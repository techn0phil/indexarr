package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"indexarr/internal/models"
	"indexarr/internal/repository"
)

const (
	tvdbBaseURL = "https://api4.thetvdb.com/v4"
)

// TVClient handles TV show metadata lookups using TVDB API v4
type TVClient struct {
	apiKey      string
	db          *sql.DB
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

// NewTVClient creates a new TV metadata client with database for token persistence
func NewTVClient(apiKey string, db *sql.DB) *TVClient {
	client := &TVClient{
		apiKey: apiKey,
		db:     db,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	// Try to load token from database
	if db != nil {
		token, expiry, err := repository.GetTVDBToken(db)
		if err == nil && token != "" && time.Now().Before(expiry) {
			client.token = token
			client.tokenExpiry = expiry
			log.Printf("Loaded TVDB token from database (expires: %s)", expiry.Format(time.RFC3339))
		}
	}

	return client
}

// TVDBLoginRequest represents the login request body
type TVDBLoginRequest struct {
	APIKey string `json:"apikey"`
}

// TVDBLoginResponse represents the login response
type TVDBLoginResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token string `json:"token"`
	} `json:"data"`
}

// TVDBSearchResponse represents search results from TVDB
type TVDBSearchResponse struct {
	Status string `json:"status"`
	Data   []struct {
		TVDBId       string `json:"tvdb_id"`
		Name         string `json:"name"`
		Overview     string `json:"overview"`
		Year         string `json:"year"`
		PrimaryType  string `json:"primary_type"` // "series", "movie", etc.
		Status       string `json:"status"`
		ImageURL     string `json:"image_url"`
		FirstAirTime string `json:"first_air_time"`
	} `json:"data"`
}

// TVDBSeriesExtended represents detailed series info
type TVDBSeriesExtended struct {
	Status string `json:"status"`
	Data   struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Slug       string `json:"slug"`
		Image      string `json:"image"`
		Overview   string `json:"overview"`
		FirstAired string `json:"firstAired"`
		LastAired  string `json:"lastAired"`
		Status     struct {
			Name string `json:"name"` // "Ended", "Continuing", etc.
		} `json:"status"`
		OriginalLanguage string  `json:"originalLanguage"`
		Score            float64 `json:"score"` // User rating
		Artworks         []struct {
			Image    string `json:"image"`
			Type     int    `json:"type"` // 2 = poster, 3 = banner, etc.
			Language string `json:"language"`
		} `json:"artworks"`
		Genres []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"genres"`
		RemoteIDs []struct {
			ID         string `json:"id"`
			Type       int    `json:"type"`       // 2 = IMDB, 10 = TMDB
			SourceName string `json:"sourceName"` // "IMDB", "TheMovieDB.com"
		} `json:"remoteIds"`
		Characters []struct {
			ID           int    `json:"id"`
			Name         string `json:"name"`
			PeopleName   string `json:"peopleName"`
			PersonImgURL string `json:"personImgURL"`
		} `json:"characters"`
		Seasons []struct {
			ID     int `json:"id"`
			Number int `json:"number"`
			Type   struct {
				ID   int    `json:"id"`   // 1 = Aired order, 2 = DVD order, 3 = Absolute order, 4 = Alternate order, etc.
				Name string `json:"name"` // "Aired order", "DVD order", etc.
				Type string `json:"type"` // "official", "alternate", etc.
			} `json:"type"`
		} `json:"seasons"`
	} `json:"data"`
}

// TVDBEpisodeExtended represents detailed episode info
type TVDBEpisodeExtended struct {
	Status string `json:"status"`
	Data   struct {
		ID               int     `json:"id"`
		Name             string  `json:"name"`
		Overview         string  `json:"overview"`
		Aired            string  `json:"aired"`
		Runtime          int     `json:"runtime"` // minutes
		SeasonNumber     int     `json:"seasonNumber"`
		Number           int     `json:"number"` // episode number
		Image            string  `json:"image"`
		Year             string  `json:"year"`
		IsMovie          int     `json:"isMovie"`
		AirsAfterSeason  int     `json:"airsAfterSeason"`
		AirsBeforeSeason int     `json:"airsBeforeSeason"`
		Score            float64 `json:"score"`
	} `json:"data"`
}

// TVDBSeasonExtended represents season with episodes
type TVDBSeasonExtended struct {
	Status string `json:"status"`
	Data   struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Number   int    `json:"number"`
		SeriesID int    `json:"series"`
		Episodes []struct {
			ID           int    `json:"id"`
			Name         string `json:"name"`
			SeasonNumber int    `json:"seasonNumber"`
			Number       int    `json:"number"` // episode number
			Runtime      int    `json:"runtime"`
			Aired        string `json:"aired"`
		} `json:"episodes"`
	} `json:"data"`
}

// TVDBBulkEpisode represents an episode from bulk episodes endpoint
type TVDBBulkEpisode struct {
	ID           int    `json:"id"`
	SeriesID     int    `json:"seriesId"`
	Name         string `json:"name"`
	Aired        string `json:"aired"`
	Runtime      int    `json:"runtime"` // minutes
	Overview     string `json:"overview"`
	SeasonNumber int    `json:"seasonNumber"`
	Number       int    `json:"number"` // episode number
	Image        string `json:"image"`
	Year         string `json:"year"`
}

// TVDBAllEpisodesResponse represents bulk episodes response
type TVDBAllEpisodesResponse struct {
	Status string `json:"status"`
	Data   struct {
		Episodes []TVDBBulkEpisode `json:"episodes"`
	} `json:"data"`
	Links struct {
		Prev       interface{} `json:"prev"`
		Self       string      `json:"self"`
		Next       interface{} `json:"next"`
		TotalItems int         `json:"total_items"`
		PageSize   int         `json:"page_size"`
	} `json:"links"`
}

// ensureValidToken checks token validity and refreshes if needed
func (c *TVClient) ensureValidToken() error {
	c.mu.RLock()
	// Check if token exists and is valid for at least 1 more day (proactive refresh)
	if c.token != "" && time.Now().Add(24*time.Hour).Before(c.tokenExpiry) {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	// Need to refresh token
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have refreshed)
	if c.token != "" && time.Now().Add(24*time.Hour).Before(c.tokenExpiry) {
		return nil
	}

	// Perform login
	return c.login()
}

// login authenticates with TVDB API v4 and stores the token
func (c *TVClient) login() error {
	if c.apiKey == "" {
		return fmt.Errorf("TVDB API key not configured")
	}

	reqBody := TVDBLoginRequest{
		APIKey: c.apiKey,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	resp, err := c.httpClient.Post(tvdbBaseURL+"/login", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("TVDB login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("TVDB login failed: %s - %s", resp.Status, string(body))
	}

	var loginResp TVDBLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if loginResp.Data.Token == "" {
		return fmt.Errorf("TVDB login response missing token")
	}

	// Token expires after 30 days (be conservative, use 29 days)
	expiry := time.Now().Add(29 * 24 * time.Hour)

	c.token = loginResp.Data.Token
	c.tokenExpiry = expiry

	// Persist token to database
	if c.db != nil {
		if err := repository.SaveTVDBToken(c.db, c.token, expiry); err != nil {
			log.Printf("Warning: Failed to save TVDB token to database: %v", err)
		} else {
			log.Printf("TVDB token saved to database (expires: %s)", expiry.Format(time.RFC3339))
		}
	}

	log.Println("TVDB authentication successful")
	return nil
}

// makeAuthenticatedRequest makes an authenticated request to TVDB API
func (c *TVClient) makeAuthenticatedRequest(method, path string, params url.Values) (*http.Response, error) {
	// Ensure token is valid
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	// Build URL
	reqURL := tvdbBaseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	// Create request
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Add authorization header
	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.token)
	c.mu.RUnlock()

	// Get time before request for logging
	startTime := time.Now()

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Log request duration in milliseconds
	duration := time.Since(startTime)
	log.Printf("[TVDB] API Response: %s %s - %d (%d ms)", method, reqURL, resp.StatusCode, duration.Milliseconds())

	// Handle 401 (token expired) - reactive refresh
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		log.Println("TVDB token expired (401), refreshing...")

		// Force token refresh
		c.mu.Lock()
		c.token = "" // Clear token to force refresh
		c.mu.Unlock()

		if err := c.ensureValidToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh TVDB token: %w", err)
		}

		// Retry request with new token
		req, err := http.NewRequest(method, reqURL, nil)
		if err != nil {
			return nil, err
		}

		c.mu.RLock()
		req.Header.Set("Authorization", "Bearer "+c.token)
		c.mu.RUnlock()

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// SearchTV searches for a TV show by title
func (c *TVClient) SearchTV(title string) (*TVDBSearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TVDB API key not configured")
	}

	params := url.Values{}
	params.Set("query", title)
	params.Set("type", "series")

	resp, err := c.makeAuthenticatedRequest("GET", "/search", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TVDB search failed: %s - %s", resp.Status, string(body))
	}

	var result TVDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}

// GetTVDetails gets detailed TV show info
func (c *TVClient) GetTVDetails(tvdbID int) (*TVDBSeriesExtended, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TVDB API key not configured")
	}

	params := url.Values{}
	params.Set("meta", "episodes")
	params.Set("short", "true")

	resp, err := c.makeAuthenticatedRequest("GET", fmt.Sprintf("/series/%d/extended", tvdbID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TVDB series details failed: %s - %s", resp.Status, string(body))
	}

	var result TVDBSeriesExtended
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode series details: %w", err)
	}

	return &result, nil
}

// GetSeasonDetails gets season with episode list
func (c *TVClient) GetSeasonDetails(seasonID int) (*TVDBSeasonExtended, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TVDB API key not configured")
	}

	resp, err := c.makeAuthenticatedRequest("GET", fmt.Sprintf("/seasons/%d/extended", seasonID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TVDB season details failed: %s - %s", resp.Status, string(body))
	}

	var result TVDBSeasonExtended
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode season details: %w", err)
	}

	return &result, nil
}

// GetEpisodeDetails gets episode details by episode ID
func (c *TVClient) GetEpisodeDetails(episodeID int) (*TVDBEpisodeExtended, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TVDB API key not configured")
	}

	resp, err := c.makeAuthenticatedRequest("GET", fmt.Sprintf("/episodes/%d/extended", episodeID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TVDB episode details failed: %s - %s", resp.Status, string(body))
	}

	var result TVDBEpisodeExtended
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode episode details: %w", err)
	}

	return &result, nil
}

// GetAllEpisodes gets all episodes for a series in bulk (optimized endpoint)
func (c *TVClient) GetAllEpisodes(tvdbID int, language string) (*TVDBAllEpisodesResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TVDB API key not configured")
	}

	// Default to French if no language specified
	if language == "" {
		language = "fra"
	}

	params := url.Values{}
	params.Set("page", "0")

	resp, err := c.makeAuthenticatedRequest("GET", fmt.Sprintf("/series/%d/episodes/default/%s", tvdbID, language), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TVDB bulk episodes failed: %s - %s", resp.Status, string(body))
	}

	var result TVDBAllEpisodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode bulk episodes: %w", err)
	}

	log.Printf("Fetched %d episodes for series %d (page size: %d, total: %d)",
		len(result.Data.Episodes), tvdbID, result.Links.PageSize, result.Links.TotalItems)

	// TODO: Handle pagination if total_items > page_size (rare for most series)
	if result.Links.TotalItems > result.Links.PageSize {
		log.Printf("Warning: Series %d has %d total episodes but only fetched %d (pagination not yet implemented)",
			tvdbID, result.Links.TotalItems, len(result.Data.Episodes))
	}

	return &result, nil
}

// EnrichSeries populates a series with metadata from TVDB
func (c *TVClient) EnrichSeries(series *models.Series) error {
	if c.apiKey == "" {
		return nil // No API key, skip enrichment
	}

	// Search for the series
	results, err := c.SearchTV(series.Title)
	if err != nil {
		return fmt.Errorf("TVDB search failed: %w", err)
	}

	if len(results.Data) == 0 {
		log.Printf("No TVDB results found for: %s", series.Title)
		return nil // No matches found
	}

	// Use the first result
	match := results.Data[0]

	// Convert TVDB ID from string to int
	tvdbID, err := strconv.Atoi(match.TVDBId)
	if err != nil {
		return fmt.Errorf("invalid TVDB ID: %s", match.TVDBId)
	}

	// Get full details
	details, err := c.GetTVDetails(tvdbID)
	if err != nil {
		return fmt.Errorf("TVDB details fetch failed: %w", err)
	}

	// Update series with TVDB data
	series.Title = details.Data.Name
	series.Synopsis = details.Data.Overview
	series.TVDBId = int64(details.Data.ID)

	// Get TMDB ID from remote IDs if available
	for _, remoteID := range details.Data.RemoteIDs {
		if remoteID.Type == 12 { // TMDB
			tmdbID, err := strconv.Atoi(remoteID.ID)
			if err == nil {
				series.TMDBId = int64(tmdbID)
			}
			break
		}
	}

	// Calculate rating (TVDB score is 0-10, we keep it as float64)
	series.Rating = details.Data.Score

	// Placeholder for popularity (TVDB doesn't provide this directly)
	series.Popularity = 0

	// Parse years from air dates
	if details.Data.FirstAired != "" && len(details.Data.FirstAired) >= 4 {
		series.YearStart, _ = strconv.Atoi(details.Data.FirstAired[:4])
	}
	if details.Data.LastAired != "" && len(details.Data.LastAired) >= 4 {
		series.YearEnd, _ = strconv.Atoi(details.Data.LastAired[:4])
	}

	// Map status
	switch strings.ToLower(details.Data.Status.Name) {
	case "ended", "canceled":
		series.Status = "complete"
	case "continuing":
		series.Status = "ongoing"
	default:
		series.Status = "ongoing"
	}

	// Extract genres
	var genres []string
	for _, g := range details.Data.Genres {
		genres = append(genres, g.Name)
	}
	series.Genres = strings.Join(genres, ", ")

	// Extract cast (top 10)
	series.Cast = []models.Cast{}
	for i, char := range details.Data.Characters {
		if i >= 10 {
			break
		}
		avatar := ""
		if char.PersonImgURL != "" {
			avatar = char.PersonImgURL
		}
		series.Cast = append(series.Cast, models.Cast{
			Name:   char.PeopleName,
			Role:   char.Name, // Character name
			Avatar: avatar,
		})
	}

	series.Poster = &details.Data.Image

	// // Extract poster (type 2 = poster)
	// for _, artwork := range details.Data.Artworks {
	// 	if artwork.Type == 2 && artwork.Image != "" {
	// 		poster := artwork.Image
	// 		series.Poster = &poster
	// 		break
	// 	}
	// }

	// Extract external IDs
	for _, remoteID := range details.Data.RemoteIDs {
		switch remoteID.Type {
		case 2: // IMDB
			series.IMDbId = remoteID.ID
		}
	}

	// Season and episode counts from seasons array
	series.SeasonCount = len(details.Data.Seasons)
	// Episode count will be updated later by scanner via UpdateSeriesCounts

	return nil
}

// EnrichEpisode populates an episode with metadata (requires series TVDB ID)
func (c *TVClient) EnrichEpisode(episode *models.Episode, seriesTVDBID int) error {
	if c.apiKey == "" || seriesTVDBID == 0 {
		return nil // No API key or no series ID, skip enrichment
	}

	// Get series details to find season ID
	seriesDetails, err := c.GetTVDetails(seriesTVDBID)
	if err != nil {
		log.Printf("Failed to get series details for episode enrichment: %v", err)
		return nil // Don't fail the scan
	}

	// Find the matching season
	var seasonID int
	for _, season := range seriesDetails.Data.Seasons {
		if season.Number == episode.SeasonNum {
			seasonID = season.ID
			break
		}
	}

	if seasonID == 0 {
		log.Printf("Season %d not found in TVDB for series ID %d", episode.SeasonNum, seriesTVDBID)
		return nil
	}

	// Get season details with episodes
	seasonDetails, err := c.GetSeasonDetails(seasonID)
	if err != nil {
		log.Printf("Failed to get season details: %v", err)
		return nil
	}

	// Find the matching episode by episode number
	var episodeID int
	for _, ep := range seasonDetails.Data.Episodes {
		if ep.Number == episode.EpisodeNum && ep.SeasonNumber == episode.SeasonNum {
			episodeID = ep.ID
			episode.Title = ep.Name
			if episode.Duration == 0 && ep.Runtime > 0 {
				episode.Duration = ep.Runtime * 60 // Convert minutes to seconds
			}
			break
		}
	}

	// If we found an episode ID, get extended details for more metadata
	if episodeID > 0 {
		episodeDetails, err := c.GetEpisodeDetails(episodeID)
		if err != nil {
			log.Printf("Failed to get episode details: %v", err)
			return nil
		}

		if episodeDetails.Data.Name != "" {
			episode.Title = episodeDetails.Data.Name
		}
		if episode.Duration == 0 && episodeDetails.Data.Runtime > 0 {
			episode.Duration = episodeDetails.Data.Runtime * 60 // Convert minutes to seconds
		}
	}

	// Default title if still empty
	if episode.Title == "" {
		episode.Title = fmt.Sprintf("Episode %d", episode.EpisodeNum)
	}

	return nil
}
