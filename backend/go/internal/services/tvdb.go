package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"indexarr/internal/models"
)

// TVClient handles TV show metadata lookups using TMDB's TV endpoints
// Note: TVDB v4 requires complex auth flow, so we use TMDB for TV shows too
type TVClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewTVClient creates a new TV metadata client
func NewTVClient(apiKey string) *TVClient {
	return &TVClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TMDBTVSearchResult represents TV search results
type TMDBTVSearchResult struct {
	Page         int `json:"page"`
	TotalResults int `json:"total_results"`
	Results      []struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		Overview     string  `json:"overview"`
		FirstAirDate string  `json:"first_air_date"`
		VoteAverage  float64 `json:"vote_average"`
		Popularity   float64 `json:"popularity"`
		PosterPath   string  `json:"poster_path"`
	} `json:"results"`
}

// TMDBTVDetails represents TV show details
type TMDBTVDetails struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Overview         string  `json:"overview"`
	FirstAirDate     string  `json:"first_air_date"`
	LastAirDate      string  `json:"last_air_date"`
	VoteAverage      float64 `json:"vote_average"`
	Popularity       float64 `json:"popularity"`
	Status           string  `json:"status"` // "Ended", "Returning Series", etc.
	NumberOfSeasons  int     `json:"number_of_seasons"`
	NumberOfEpisodes int     `json:"number_of_episodes"`
	Genres           []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	ExternalIDs struct {
		IMDbID string `json:"imdb_id"`
		TVDBID int    `json:"tvdb_id"`
	} `json:"external_ids"`
	Credits struct {
		Cast []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Character   string `json:"character"`
			ProfilePath string `json:"profile_path"`
		} `json:"cast"`
	} `json:"credits"`
}

// TMDBEpisodeDetails represents episode details
type TMDBEpisodeDetails struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	AirDate       string  `json:"air_date"`
	SeasonNumber  int     `json:"season_number"`
	EpisodeNumber int     `json:"episode_number"`
	Runtime       int     `json:"runtime"`
	VoteAverage   float64 `json:"vote_average"`
}

// SearchTV searches for a TV show by title
func (c *TVClient) SearchTV(title string) (*TMDBTVSearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", title)
	params.Set("language", "fr-FR")

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/search/tv?%s", tmdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var result TMDBTVSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTVDetails gets detailed TV show info
func (c *TVClient) GetTVDetails(tmdbID int) (*TMDBTVDetails, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "fr-FR")
	params.Set("append_to_response", "credits,external_ids")

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/tv/%d?%s", tmdbBaseURL, tmdbID, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var details TMDBTVDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetEpisodeDetails gets episode details
func (c *TVClient) GetEpisodeDetails(tmdbID, season, episode int) (*TMDBEpisodeDetails, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "fr-FR")

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/tv/%d/season/%d/episode/%d?%s", tmdbBaseURL, tmdbID, season, episode, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var details TMDBEpisodeDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

// EnrichSeries populates a series with metadata
func (c *TVClient) EnrichSeries(series *models.Series) error {
	if c.apiKey == "" {
		return nil // No API key, skip enrichment
	}

	// Search for the series
	results, err := c.SearchTV(series.Title)
	if err != nil {
		return err
	}

	if len(results.Results) == 0 {
		return nil // No matches found
	}

	// Use the first result
	match := results.Results[0]

	// Get full details
	details, err := c.GetTVDetails(match.ID)
	if err != nil {
		return err
	}

	// Update series with TMDB data
	series.Title = details.Name
	series.Synopsis = details.Overview
	series.Rating = details.VoteAverage
	series.Popularity = details.Popularity

	// ----------------------------------------------------
	// -------------  To review !! ------------------------
	// ----------------------------------------------------
	// series.TVDBId = int64(details.ExternalIDs.TVDBID)
	series.TVDBId = int64(details.ID)
	// ----------------------------------------------------
	// ----------------------------------------------------
	// ----------------------------------------------------

	series.IMDbId = details.ExternalIDs.IMDbID
	series.SeasonCount = details.NumberOfSeasons
	series.EpisodeCount = details.NumberOfEpisodes

	// Parse years
	if details.FirstAirDate != "" && len(details.FirstAirDate) >= 4 {
		series.YearStart, _ = strconv.Atoi(details.FirstAirDate[:4])
	}
	if details.LastAirDate != "" && len(details.LastAirDate) >= 4 {
		series.YearEnd, _ = strconv.Atoi(details.LastAirDate[:4])
	}

	// Set status based on TMDB status
	switch details.Status {
	case "Ended", "Canceled":
		series.Status = "complete"
	case "Returning Series", "In Production":
		series.Status = "ongoing"
	default:
		series.Status = "ongoing"
	}

	// Extract genres
	var genres []string
	for _, g := range details.Genres {
		genres = append(genres, g.Name)
	}
	series.Genres = strings.Join(genres, ", ")

	// Extract cast (top 10)
	series.Cast = []models.Cast{}
	for i, c := range details.Credits.Cast {
		if i >= 10 {
			break
		}
		avatar := ""
		if c.ProfilePath != "" {
			avatar = tmdbImageBaseURL + c.ProfilePath
		}
		series.Cast = append(series.Cast, models.Cast{
			Name:   c.Name,
			Role:   c.Character,
			Avatar: avatar,
		})
	}

	// Extract poster URL
	if match.PosterPath != "" {
		poster := tmdbImageBaseURL + match.PosterPath
		series.Poster = &poster
	}

	return nil
}

// EnrichEpisode populates an episode with metadata (requires series TMDB ID)
func (c *TVClient) EnrichEpisode(episode *models.Episode, seriesTMDBID int) error {
	if c.apiKey == "" || seriesTMDBID == 0 {
		return nil // No API key or no series ID, skip enrichment
	}

	details, err := c.GetEpisodeDetails(seriesTMDBID, episode.SeasonNum, episode.EpisodeNum)
	if err != nil {
		return err // Don't fail if episode lookup fails
	}

	episode.Title = details.Name
	if episode.Duration == 0 && details.Runtime > 0 {
		episode.Duration = details.Runtime * 60 // Convert minutes to seconds
	}

	return nil
}
