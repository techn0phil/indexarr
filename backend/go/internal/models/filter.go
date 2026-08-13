package models

type FilterCriteria struct {
	Status     string `json:"status"`     // available, missing, problem
	Resolution string `json:"resolution"` // 4K, 1080p, 720p
	Codec      string `json:"codec"`      // H.264, H.265, VP9, AV1
	Audio      string `json:"audio"`      // AAC, DTS, TrueHD, etc
	HDR        string `json:"hdr"`        // Dolby Vision, HDR10+, HDR10, none
	Sort       string `json:"sort"`       // title, year, added, size
	Search     string `json:"search"`     // search query for title
	Page       int    `json:"page"`       // 1-based
	PageSize   int    `json:"pageSize"`   // default 50
}

type PaginatedResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Error    string      `json:"error,omitempty"`
}

type StatsResponse struct {
	Success         bool    `json:"success"`
	TotalMovies     int64   `json:"totalMovies"`
	TotalSeries     int64   `json:"totalSeries"`
	TotalEpisodes   int64   `json:"totalEpisodes"`
	DiskSpaceGB     float64 `json:"diskSpaceGB"`
	MoviesDiskSpaceGB float64 `json:"moviesDiskSpaceGB"`
	SeriesDiskSpaceGB float64 `json:"seriesDiskSpaceGB"`
	FourKCount      int64   `json:"fourKCount"`
	FourKPercent    float64 `json:"fourKPercent"`
	ProblemsCount   int64   `json:"problemsCount"`
	AvailMovies     int64   `json:"availMovies"`
	MissingMovies   int64   `json:"missingMovies"`
	AvailEpisodes   int64   `json:"availEpisodes"`
	MissingEpisodes int64   `json:"missingEpisodes"`
	Error           string  `json:"error,omitempty"`
}
