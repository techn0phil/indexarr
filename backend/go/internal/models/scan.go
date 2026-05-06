package models

type ScanStatus struct {
	ID             int64  `json:"id"`
	Status         string `json:"status"` // idle, running, completed, error
	StartedAt      string `json:"startedAt,omitempty"`
	CompletedAt    string `json:"completedAt,omitempty"`
	FilesFound     int    `json:"filesFound"`
	FilesProcessed int    `json:"filesProcessed"`
	ErrorMessage   string `json:"errorMessage,omitempty"`
}

type ScanResult struct {
	FilesFound     int      `json:"filesFound"`
	FilesProcessed int      `json:"filesProcessed"`
	MoviesAdded    int      `json:"moviesAdded"`
	EpisodesAdded  int      `json:"episodesAdded"`
	Errors         []string `json:"errors"`
}
