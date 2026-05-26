package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RadarrClient is an HTTP client for Radarr API v3
type RadarrClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewRadarrClient creates a new Radarr API client
func NewRadarrClient(baseURL, apiKey string) *RadarrClient {
	return &RadarrClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RadarrMovie represents a movie from Radarr API
type RadarrMovie struct {
	ID          int              `json:"id"`
	Title       string           `json:"title"`
	Year        int              `json:"year"`
	Overview    string           `json:"overview"`
	Runtime     int              `json:"runtime"` // minutes
	HasFile     bool             `json:"hasFile"`
	Path        string           `json:"path"`
	FolderName  string           `json:"folderName"`
	Genres      []string         `json:"genres"`
	Ratings     RadarrRatings    `json:"ratings"`
	Images      []RadarrImage    `json:"images"`
	TmdbID      int              `json:"tmdbId"`
	ImdbID      string           `json:"imdbId"`
	Added       string           `json:"added"` // ISO 8601 timestamp
	MovieFile   *RadarrMovieFile `json:"movieFile,omitempty"`
	Status      string           `json:"status"` // released, inCinemas, announced
	IsAvailable bool             `json:"isAvailable"`
}

// RadarrMovieFile represents the file information for a movie
type RadarrMovieFile struct {
	ID           int              `json:"id"`
	MovieID      int              `json:"movieId"`
	RelativePath string           `json:"relativePath"`
	Path         string           `json:"path"`
	Size         int64            `json:"size"`
	DateAdded    string           `json:"dateAdded"`
	Quality      RadarrQuality    `json:"quality"`
	MediaInfo    *RadarrMediaInfo `json:"mediaInfo,omitempty"`
	Languages    []RadarrLanguage `json:"languages,omitempty"`
}

// RadarrQuality represents the quality profile
type RadarrQuality struct {
	Quality struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Resolution int    `json:"resolution"`
	} `json:"quality"`
}

// RadarrMediaInfo contains technical details about the file (from Radarr's own extraction)
type RadarrMediaInfo struct {
	AudioBitrate      int     `json:"audioBitrate"`
	AudioChannels     float64 `json:"audioChannels"`
	AudioCodec        string  `json:"audioCodec"`
	AudioLanguages    string  `json:"audioLanguages"`
	AudioStreamCount  int     `json:"audioStreamCount"`
	VideoBitDepth     int     `json:"videoBitDepth"`
	VideoBitrate      int     `json:"videoBitrate"`
	VideoCodec        string  `json:"videoCodec"`
	VideoFps          float64 `json:"videoFps"`
	VideoDynamicRange string  `json:"videoDynamicRange"` // SDR, HDR, HDR10, HDR10+, Dolby Vision
	Resolution        string  `json:"resolution"`
	RunTime           string  `json:"runTime"`
	ScanType          string  `json:"scanType"`
	Subtitles         string  `json:"subtitles"`
}

// RadarrRatings contains rating information
type RadarrRatings struct {
	Imdb   RadarrRating `json:"imdb"`
	Tmdb   RadarrRating `json:"tmdb"`
	Rotten RadarrRating `json:"rottenTomatoes"`
}

// RadarrRating represents a single rating value
type RadarrRating struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// RadarrImage represents an image (poster, fanart, etc.)
type RadarrImage struct {
	CoverType string `json:"coverType"` // poster, fanart, banner
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}

// RadarrLanguage represents a language
type RadarrLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RadarrSystemStatus represents the system status response
type RadarrSystemStatus struct {
	AppName string `json:"appName"`
	Version string `json:"version"`
}

// doRequest performs an HTTP request to the Radarr API
func (c *RadarrClient) doRequest(method, endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v3%s", c.baseURL, endpoint)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// TestConnection verifies the API connection is working
func (c *RadarrClient) TestConnection() error {
	body, err := c.doRequest("GET", "/system/status")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	var status RadarrSystemStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("failed to parse status response: %w", err)
	}

	if status.AppName != "Radarr" {
		return fmt.Errorf("unexpected app name: %s (expected Radarr)", status.AppName)
	}

	return nil
}

// GetMovies fetches all movies from Radarr
func (c *RadarrClient) GetMovies() ([]RadarrMovie, error) {
	body, err := c.doRequest("GET", "/movie")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movies: %w", err)
	}

	var movies []RadarrMovie
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, fmt.Errorf("failed to parse movies response: %w", err)
	}

	return movies, nil
}

// GetMovie fetches a single movie by its Radarr ID
func (c *RadarrClient) GetMovie(id int) (*RadarrMovie, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/movie/%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie %d: %w", id, err)
	}

	var movie RadarrMovie
	if err := json.Unmarshal(body, &movie); err != nil {
		return nil, fmt.Errorf("failed to parse movie response: %w", err)
	}

	return &movie, nil
}

// GetMovieByTMDBId fetches a movie by its TMDB ID
func (c *RadarrClient) GetMovieByTMDBId(tmdbID int) (*RadarrMovie, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/movie?tmdbId=%d", tmdbID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie by TMDB ID %d: %w", tmdbID, err)
	}

	var movies []RadarrMovie
	if err := json.Unmarshal(body, &movies); err != nil {
		return nil, fmt.Errorf("failed to parse movies response: %w", err)
	}

	if len(movies) == 0 {
		return nil, nil
	}

	return &movies[0], nil
}
