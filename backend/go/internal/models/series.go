package models

type Series struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	YearStart     int    `json:"yearStart"`
	YearEnd       int    `json:"yearEnd"`
	SeasonCount   int    `json:"seasonCount"`
	EpisodeCount  int    `json:"episodeCount"`
	Synopsis      string `json:"synopsis"`
	Genres        string `json:"genres"` // comma-separated
	Rating        float64 `json:"rating"` // TVDB rating
	Popularity    float64 `json:"popularity"`
	Status        string `json:"status"` // complete, ongoing, partial
	FileSize      int64  `json:"fileSize"` // bytes
	DateAdded     string `json:"dateAdded"` // ISO 8601
	TVDBId        int64  `json:"tvdbId"`
	IMDbId        string `json:"imdbId"`
	Poster        *string `json:"poster"` // TMDB poster URL
	Cast          []Cast `json:"cast"`
	Seasons       []Season `json:"seasons"`
}

type Season struct {
	ID           int64     `json:"id"`
	SeriesID     int64     `json:"seriesId"`
	Number       int       `json:"number"`
	Episodes     []Episode `json:"episodes"`
	FileSize     int64     `json:"fileSize"` // bytes
	AvailableEps int       `json:"availableEps"`
	MissingEps   int       `json:"missingEps"`
}
