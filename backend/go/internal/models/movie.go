package models

type Movie struct {
	ID         int64      `json:"id"`
	Title      string     `json:"title"`
	Year       int        `json:"year"`
	Duration   int        `json:"duration"` // minutes
	Synopsis   string     `json:"synopsis"`
	Genres     string     `json:"genres"` // comma-separated
	Rating     float64    `json:"rating"` // TMDB rating
	Popularity float64    `json:"popularity"`
	Status     string     `json:"status"`   // available, missing, problem
	FileSize   int64      `json:"fileSize"` // bytes
	FilePath   string     `json:"filePath"`
	Container  string     `json:"container"` // mkv, mp4, etc
	DateAdded  string     `json:"dateAdded"` // ISO 8601
	TMDBId     int64      `json:"tmdbId"`
	IMDbId     string     `json:"imdbId"`
	Poster     *string    `json:"poster"`
	Cast       []Cast     `json:"cast"`
	MediaInfo  *MediaInfo `json:"mediaInfo"`
}

type Cast struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Avatar string `json:"avatar"` // URL
}
