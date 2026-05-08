package models

type MediaInfo struct {
	ID             int64           `json:"id"`
	VideoTracks    []VideoTrack    `json:"videoTracks"`
	AudioTracks    []AudioTrack    `json:"audioTracks"`
	SubtitleTracks []SubtitleTrack `json:"subtitleTracks"`
}

type VideoTrack struct {
	Codec      string  `json:"codec"`      // H.264, H.265, VP9, etc
	Resolution string  `json:"resolution"` // 3840x2160, 1920x1080, etc
	FPS        float64 `json:"fps"`        // 23.976, 24, 30, 60, etc
	Bitrate    string  `json:"bitrate"`    // 35.2 Mbps
	HDR        string  `json:"hdr"`        // Dolby Vision, HDR10, HDR10+, none
	ColorSpace string  `json:"colorSpace"` // BT.2020, BT.709, etc
}

type AudioTrack struct {
	Codec      string `json:"codec"`      // AAC, AC3, TrueHD, DTS, etc
	Channels   string `json:"channels"`   // 2.0, 5.1, 7.1, etc
	Language   string `json:"language"`   // English, French, etc
	SampleRate string `json:"sampleRate"` // 48000 Hz
	Bitrate    string `json:"bitrate"`    // 128 kbps
	Default    string `json:"default"`    // Yes, No
}

type SubtitleTrack struct {
	Language string `json:"language"` // English, French, etc
	Format   string `json:"format"`   // SRT, ASS, PGS, etc
	Forced   string `json:"forced"`   // Yes, No
	Default  string `json:"default"`  // Yes, No
}
