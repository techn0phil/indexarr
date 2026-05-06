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

// SearchMovie searches for a movie by title and optional year
func (c *TMDBClient) SearchMovie(title string, year int) (*TMDBSearchResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	params := url.Values{}
	params.Set("api_key", c.apiKey)
	params.Set("query", title)
	params.Set("language", "fr-FR")
	if year > 0 {
		params.Set("year", strconv.Itoa(year))
	}

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/search/movie?%s", tmdbBaseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error: %s - %s", resp.Status, string(body))
	}

	var result TMDBSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
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

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/movie/%d?%s", tmdbBaseURL, tmdbID, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
