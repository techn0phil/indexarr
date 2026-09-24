package services

import "testing"

func TestParseFilename_MovieAndSeriesCases(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		wantTitle      string
		wantYear       int
		wantIsSeries   bool
		wantSeason     int
		wantEpisode    int
		wantResolution string
		wantSource     string
	}{
		{
			name:           "movie with year resolution and source",
			filename:       "The.Matrix.1999.1080p.BluRay.x264-GROUP.mkv",
			wantTitle:      "The Matrix",
			wantYear:       1999,
			wantIsSeries:   false,
			wantSeason:     0,
			wantEpisode:    0,
			wantResolution: "1080p",
			wantSource:     "BluRay",
		},
		{
			name:           "series with SxxExx",
			filename:       "Breaking.Bad.S01E02.2160p.WEB-DL.HEVC.mkv",
			wantTitle:      "Breaking Bad",
			wantYear:       0,
			wantIsSeries:   true,
			wantSeason:     1,
			wantEpisode:    2,
			wantResolution: "4K",
			wantSource:     "WEB-DL",
		},
		{
			name:           "series with 1x05",
			filename:       "Some.Show.1x05.720p.HDTV.x264.mp4",
			wantTitle:      "Some Show",
			wantYear:       0,
			wantIsSeries:   true,
			wantSeason:     1,
			wantEpisode:    5,
			wantResolution: "720p",
			wantSource:     "HDTV",
		},
		{
			name:           "movie with parenthesis year",
			filename:       "Dune (2021) 4K Remux.mkv",
			wantTitle:      "Dune",
			wantYear:       2021,
			wantIsSeries:   false,
			wantSeason:     0,
			wantEpisode:    0,
			wantResolution: "4K",
			wantSource:     "Remux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFilename(tt.filename)

			if got.Title != tt.wantTitle {
				t.Fatalf("Title: got %q want %q", got.Title, tt.wantTitle)
			}
			if got.Year != tt.wantYear {
				t.Fatalf("Year: got %d want %d", got.Year, tt.wantYear)
			}
			if got.IsSeries != tt.wantIsSeries {
				t.Fatalf("IsSeries: got %v want %v", got.IsSeries, tt.wantIsSeries)
			}
			if got.Season != tt.wantSeason {
				t.Fatalf("Season: got %d want %d", got.Season, tt.wantSeason)
			}
			if got.Episode != tt.wantEpisode {
				t.Fatalf("Episode: got %d want %d", got.Episode, tt.wantEpisode)
			}
			if got.Resolution != tt.wantResolution {
				t.Fatalf("Resolution: got %q want %q", got.Resolution, tt.wantResolution)
			}
			if got.Source != tt.wantSource {
				t.Fatalf("Source: got %q want %q", got.Source, tt.wantSource)
			}
		})
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "mkv", filename: "movie.mkv", want: true},
		{name: "upper-case extension", filename: "movie.MP4", want: true},
		{name: "unsupported extension", filename: "readme.txt", want: false},
		{name: "no extension", filename: "movie", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsVideoFile(tt.filename); got != tt.want {
				t.Fatalf("IsVideoFile(%q): got %v want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsBlurayFolder(t *testing.T) {
	if !IsBlurayFolder("/media/Movie/BDMV") {
		t.Fatalf("expected BDMV path to be recognized")
	}
	if !IsBlurayFolder("/media/Movie/bdmv") {
		t.Fatalf("expected case-insensitive BDMV path to be recognized")
	}
	if IsBlurayFolder("/media/Movie/VIDEO_TS") {
		t.Fatalf("expected non-BDMV folder to be rejected")
	}
}

func TestGetContainer(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "mkv", filename: "movie.mkv", want: "mkv"},
		{name: "upper-case extension", filename: "movie.MP4", want: "mp4"},
		{name: "no extension", filename: "movie", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetContainer(tt.filename); got != tt.want {
				t.Fatalf("GetContainer(%q): got %q want %q", tt.filename, got, tt.want)
			}
		})
	}
}
