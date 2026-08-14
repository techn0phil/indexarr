package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SonarrClient is an HTTP client for Sonarr API v3
type SonarrClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSonarrClient creates a new Sonarr API client
func NewSonarrClient(baseURL, apiKey string) *SonarrClient {
	return &SonarrClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SonarrSeries represents a TV series from Sonarr API
type SonarrSeries struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	TitleSlug   string         `json:"titleSlug"`
	Year        int            `json:"year"`
	Overview    string         `json:"overview"`
	Status      string         `json:"status"` // continuing, ended
	SeasonCount int            `json:"seasonCount"`
	Path        string         `json:"path"`
	Genres      []string       `json:"genres"`
	Ratings     SonarrRatings  `json:"ratings"`
	Images      []SonarrImage  `json:"images"`
	TmdbId      int            `json:"tmdbId"`
	TvdbId      int            `json:"tvdbId"`
	TvRageId    int            `json:"tvRageId"`
	TvMazeId    int            `json:"tvMazeId"`
	ImdbId      string         `json:"imdbId"`
	Added       string         `json:"added"`   // ISO 8601 timestamp
	Runtime     int            `json:"runtime"` // minutes per episode
	Statistics  SonarrStats    `json:"statistics"`
	Seasons     []SonarrSeason `json:"seasons"`
	FirstAired  string         `json:"firstAired"` // YYYY-MM-DDTHH:MM:SSZ
	LastAired   string         `json:"lastAired"`  // YYYY-MM-DDTHH:MM:SSZ
}

// SonarrSeason represents a season within a series
type SonarrSeason struct {
	SeasonNumber int         `json:"seasonNumber"`
	Monitored    bool        `json:"monitored"`
	Statistics   SonarrStats `json:"statistics"`
}

// SonarrStats contains episode statistics
type SonarrStats struct {
	EpisodeFileCount  int     `json:"episodeFileCount"`
	EpisodeCount      int     `json:"episodeCount"`
	TotalEpisodeCount int     `json:"totalEpisodeCount"`
	SizeOnDisk        int64   `json:"sizeOnDisk"`
	PercentOfEpisodes float64 `json:"percentOfEpisodes"`
}

// SonarrEpisode represents an episode from Sonarr API
type SonarrEpisode struct {
	ID            int                `json:"id"`
	SeriesID      int                `json:"seriesId"`
	TvdbId        int                `json:"tvdbId"`
	EpisodeFileID int                `json:"episodeFileId"`
	SeasonNumber  int                `json:"seasonNumber"`
	EpisodeNumber int                `json:"episodeNumber"`
	Title         string             `json:"title"`
	AirDate       string             `json:"airDate"`    // YYYY-MM-DD
	AirDateUtc    string             `json:"airDateUtc"` // ISO 8601
	Overview      string             `json:"overview"`
	HasFile       bool               `json:"hasFile"`
	Monitored     bool               `json:"monitored"`
	Runtime       int                `json:"runtime"` // minutes
	EpisodeFile   *SonarrEpisodeFile `json:"episodeFile,omitempty"`
}

// SonarrEpisodeFile represents a media file for an episode
type SonarrEpisodeFile struct {
	ID           int              `json:"id"`
	SeriesID     int              `json:"seriesId"`
	SeasonNumber int              `json:"seasonNumber"`
	RelativePath string           `json:"relativePath"`
	Path         string           `json:"path"`
	Size         int64            `json:"size"`
	DateAdded    string           `json:"dateAdded"`
	Quality      SonarrQuality    `json:"quality"`
	MediaInfo    *SonarrMediaInfo `json:"mediaInfo,omitempty"`
	Languages    []SonarrLanguage `json:"languages,omitempty"`
}

// SonarrQuality represents the quality profile
type SonarrQuality struct {
	Quality struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Resolution int    `json:"resolution"`
	} `json:"quality"`
}

// SonarrMediaInfo contains technical details about the file (from Sonarr's own extraction)
type SonarrMediaInfo struct {
	AudioBitrate          int     `json:"audioBitrate"`
	AudioChannels         float64 `json:"audioChannels"`
	AudioCodec            string  `json:"audioCodec"`
	AudioLanguages        string  `json:"audioLanguages"`
	AudioStreamCount      int     `json:"audioStreamCount"`
	VideoBitDepth         int     `json:"videoBitDepth"`
	VideoBitrate          int     `json:"videoBitrate"`
	VideoCodec            string  `json:"videoCodec"`
	VideoFps              float64 `json:"videoFps"`
	VideoDynamicRange     string  `json:"videoDynamicRange"` // SDR, HDR, HDR10, HDR10+, Dolby Vision
	VideoDynamicRangeType string  `json:"videoDynamicRangeType"`
	Resolution            string  `json:"resolution"`
	RunTime               string  `json:"runTime"`
	ScanType              string  `json:"scanType"`
	Subtitles             string  `json:"subtitles"`
}

// SonarrRatings contains rating information
type SonarrRatings struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
}

// SonarrImage represents an image (poster, fanart, etc.)
type SonarrImage struct {
	CoverType string `json:"coverType"` // poster, fanart, banner
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}

// SonarrLanguage represents a language
type SonarrLanguage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SonarrSystemStatus represents the system status response
type SonarrSystemStatus struct {
	AppName string `json:"appName"`
	Version string `json:"version"`
}

// doRequest performs an HTTP request to the Sonarr API
func (c *SonarrClient) doRequest(method, endpoint string) ([]byte, error) {
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
func (c *SonarrClient) TestConnection() error {
	body, err := c.doRequest("GET", "/system/status")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	var status SonarrSystemStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("failed to parse status response: %w", err)
	}

	if status.AppName != "Sonarr" {
		return fmt.Errorf("unexpected app name: %s (expected Sonarr)", status.AppName)
	}

	return nil
}

// GetSeries fetches all series from Sonarr
func (c *SonarrClient) GetSeries() ([]SonarrSeries, error) {
	body, err := c.doRequest("GET", "/series")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series: %w", err)
	}

	var series []SonarrSeries
	if err := json.Unmarshal(body, &series); err != nil {
		return nil, fmt.Errorf("failed to parse series response: %w", err)
	}

	return series, nil
}

// GetSeriesByID fetches a single series by its Sonarr ID
func (c *SonarrClient) GetSeriesByID(id int) (*SonarrSeries, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/series/%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series %d: %w", id, err)
	}

	var series SonarrSeries
	if err := json.Unmarshal(body, &series); err != nil {
		return nil, fmt.Errorf("failed to parse series response: %w", err)
	}

	return &series, nil
}

// GetSeriesByTVDBId fetches a series by its TVDB ID
func (c *SonarrClient) GetSeriesByTVDBId(tvdbID int) (*SonarrSeries, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/series?tvdbId=%d", tvdbID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch series by TVDB ID %d: %w", tvdbID, err)
	}

	var series []SonarrSeries
	if err := json.Unmarshal(body, &series); err != nil {
		return nil, fmt.Errorf("failed to parse series response: %w", err)
	}

	if len(series) == 0 {
		return nil, nil
	}

	return &series[0], nil
}

// GetEpisodes fetches all episodes for a series from Sonarr
func (c *SonarrClient) GetEpisodes(seriesID int) ([]SonarrEpisode, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/episode?seriesId=%d&includeEpisodeFile=true", seriesID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episodes for series %d: %w", seriesID, err)
	}

	var episodes []SonarrEpisode
	if err := json.Unmarshal(body, &episodes); err != nil {
		return nil, fmt.Errorf("failed to parse episodes response: %w", err)
	}

	return episodes, nil
}

// GetEpisodeFile fetches a single episode file by its ID
func (c *SonarrClient) GetEpisodeFile(id int) (*SonarrEpisodeFile, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/episodefile/%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch episode file %d: %w", id, err)
	}

	var file SonarrEpisodeFile
	if err := json.Unmarshal(body, &file); err != nil {
		return nil, fmt.Errorf("failed to parse episode file response: %w", err)
	}

	return &file, nil
}

// GetAllSeriesSonarrIDs returns a list of all Sonarr series IDs (for stale detection)
func (c *SonarrClient) GetAllSeriesSonarrIDs() ([]int, error) {
	series, err := c.GetSeries()
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(series))
	for i, s := range series {
		ids[i] = s.ID
	}
	return ids, nil
}
