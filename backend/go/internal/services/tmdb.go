package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"indexarr/internal/models"
)

const (
	tmdbBaseURL      = "https://api.themoviedb.org/3"
	tmdbImageBaseURL = "https://image.tmdb.org/t/p/w500"
)

// TMDBClient handles TMDB API requests
type TMDBClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewTMDBClient creates a new TMDB client
func NewTMDBClient(apiKey string) *TMDBClient {
	return &TMDBClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TMDBSearchResult represents a search result from TMDB
type TMDBSearchResult struct {
	Page         int `json:"page"`
	TotalResults int `json:"total_results"`
	Results      []struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		Name         string  `json:"name"` // For TV shows
		Overview     string  `json:"overview"`
		ReleaseDate  string  `json:"release_date"`
		FirstAirDate string  `json:"first_air_date"` // For TV shows
		VoteAverage  float64 `json:"vote_average"`
		Popularity   float64 `json:"popularity"`
		PosterPath   string  `json:"poster_path"`
	} `json:"results"`
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

// TMDBMovieDetails represents movie details from TMDB
type TMDBMovieDetails struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	Runtime     int     `json:"runtime"`
	VoteAverage float64 `json:"vote_average"`
	Popularity  float64 `json:"popularity"`
	IMDbID      string  `json:"imdb_id"`
	Genres      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Credits struct {
		Cast []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Character   string `json:"character"`
			ProfilePath string `json:"profile_path"`
		} `json:"cast"`
	} `json:"credits"`
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

// SearchMovie searches for a movie by title and optional year
func (c *TMDBClient) SearchMovie(title string, year int) (*TMDBSearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", title)
	params.Set("language", "en-US")
	if year > 0 {
		params.Set("primary_release_year", strconv.Itoa(year))
	}

	// Get time before request for logging
	startTime := time.Now()

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/search/movie?%s", tmdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log request duration in milliseconds
	duration := time.Since(startTime)
	if year > 0 {
		log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/search/movie?api_key=******&language=en-US&query=%s&primary_release_year=%d", tmdbBaseURL, title, year), resp.StatusCode, duration.Milliseconds())
	} else {
		log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/search/movie?api_key=******&language=en-US&query=%s", tmdbBaseURL, title), resp.StatusCode, duration.Milliseconds())
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var result TMDBSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Retry without year filter
	if year > 0 && result.TotalResults == 0 {
		return c.SearchMovie(title, 0)
	}

	// Log number of results found
	if year > 0 {
		log.Printf("Found %d results for movie '%s' (%d)", result.TotalResults, title, year)
	} else {
		log.Printf("Found %d results for movie '%s'", result.TotalResults, title)
	}

	return &result, nil
}

// SearchTV searches for a TV show by title
func (c *TMDBClient) SearchTV(title string, year int) (*TMDBTVSearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "en-US")
	params.Set("query", title)
	if year > 0 {
		params.Set("first_air_date_year", strconv.Itoa(year))
	}

	// Get time before request for logging
	startTime := time.Now()

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/search/tv?%s", tmdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log request duration in milliseconds
	duration := time.Since(startTime)

	if year > 0 {
		log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/search/tv?api_key=******&language=en-US&query=%s&first_air_date_year=%d", tmdbBaseURL, title, year), resp.StatusCode, duration.Milliseconds())
	} else {
		log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/search/tv?api_key=******&language=en-US&query=%s", tmdbBaseURL, title), resp.StatusCode, duration.Milliseconds())
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var result TMDBTVSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Retry without year filter
	if year > 0 && result.TotalResults == 0 {
		return c.SearchTV(title, 0)
	}

	// Log number of results found
	if year > 0 {
		log.Printf("Found %d results for series '%s' (%d)", result.TotalResults, title, year)
	} else {
		log.Printf("Found %d results for series '%s'", result.TotalResults, title)
	}

	return &result, nil
}

// GetMovieDetails gets detailed movie info including credits
func (c *TMDBClient) GetMovieDetails(tmdbID int) (*TMDBMovieDetails, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "fr-FR")
	params.Set("append_to_response", "credits")

	// Get time before request for logging
	startTime := time.Now()

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/movie/%d?%s", tmdbBaseURL, tmdbID, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log request duration in milliseconds
	duration := time.Since(startTime)
	log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/movie/%d?api_key=******&language=fr-FR&append_to_response=credits", tmdbBaseURL, tmdbID), resp.StatusCode, duration.Milliseconds())

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var details TMDBMovieDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return &details, nil
}

// GetTVDetails gets detailed TV show info
func (c *TMDBClient) GetTVDetails(tmdbID int) (*TMDBTVDetails, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "fr-FR")
	params.Set("append_to_response", "external_ids")

	// Get time before request for logging
	startTime := time.Now()

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/tv/%d?%s", tmdbBaseURL, tmdbID, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log request duration in milliseconds
	duration := time.Since(startTime)
	log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/tv/%d?api_key=******&language=fr-FR&append_to_response=external_ids", tmdbBaseURL, tmdbID), resp.StatusCode, duration.Milliseconds())

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
func (c *TMDBClient) GetEpisodeDetails(tmdbID, season, episode int) (*TMDBEpisodeDetails, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("language", "fr-FR")

	// Get time before request for logging
	startTime := time.Now()

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/tv/%d/season/%d/episode/%d?%s", tmdbBaseURL, tmdbID, season, episode, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log request duration in milliseconds
	duration := time.Since(startTime)
	log.Printf("GET %s - %d (%d ms)", fmt.Sprintf("%s/tv/%d/season/%d/episode/%d?api_key=******&language=fr-FR", tmdbBaseURL, tmdbID, season, episode), resp.StatusCode, duration.Milliseconds())

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

// EnrichMovie populates a movie with TMDB metadata
func (c *TMDBClient) EnrichMovie(movie *models.Movie) error {
	if c.apiKey == "" {
		return nil // No API key, skip enrichment
	}

	// Search for the movie
	results, err := c.SearchMovie(movie.Title, movie.Year)
	if err != nil {
		return err
	}

	if len(results.Results) == 0 {
		return nil // No matches found
	}

	// Use the first result
	match := results.Results[0]

	// Get full details
	details, err := c.GetMovieDetails(match.ID)
	if err != nil {
		return err
	}

	// Update movie with TMDB data
	movie.Title = details.Title
	movie.Synopsis = details.Overview
	movie.Rating = details.VoteAverage
	movie.Popularity = details.Popularity
	movie.TMDBId = int64(details.ID)
	movie.IMDbId = details.IMDbID
	if match.PosterPath != "" {
		poster := tmdbImageBaseURL + match.PosterPath
		movie.Poster = &poster
	} else {
		movie.Poster = nil
	}

	// Parse year from release date
	if details.ReleaseDate != "" && len(details.ReleaseDate) >= 4 {
		movie.Year, _ = strconv.Atoi(details.ReleaseDate[:4])
	}

	// Duration (if mediainfo didn't provide it)
	if movie.Duration == 0 {
		movie.Duration = details.Runtime
	}

	// Extract genres
	var genres []string
	for _, g := range details.Genres {
		genres = append(genres, g.Name)
	}
	movie.Genres = strings.Join(genres, ", ")

	// Extract cast (top 10)
	movie.Cast = []models.Cast{}
	for i, c := range details.Credits.Cast {
		if i >= 10 {
			break
		}
		avatar := ""
		if c.ProfilePath != "" {
			avatar = tmdbImageBaseURL + c.ProfilePath
		}
		movie.Cast = append(movie.Cast, models.Cast{
			Name:   c.Name,
			Role:   c.Character,
			Avatar: avatar,
		})
	}

	return nil
}

// EnrichSeries populates a series with metadata
func (c *TMDBClient) EnrichSeries(series *models.Series) error {
	if c.apiKey == "" {
		return nil // No API key, skip enrichment
	}

	// Search for the series
	results, err := c.SearchTV(series.Title, series.YearStart)
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
	series.TMDBId = int64(details.ID)
	series.TVDBId = int64(details.ExternalIDs.TVDBID)
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
func (c *TMDBClient) EnrichEpisode(episode *models.Episode, seriesTMDBID int) error {
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
