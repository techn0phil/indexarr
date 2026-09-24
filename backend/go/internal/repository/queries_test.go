package repository

import (
	"fmt"
	"testing"

	"indexarr/internal/models"
)

func TestBuildOrClause(t *testing.T) {
	tests := []struct {
		name       string
		fieldName  string
		filter     string
		wantClause string
	}{
		{
			name:       "single value",
			fieldName:  "resolution",
			filter:     "3840",
			wantClause: "(resolution LIKE '%3840%')",
		},
		{
			name:       "multiple values",
			fieldName:  "codec",
			filter:     "H.265,H.264",
			wantClause: "(codec LIKE '%H.265%' OR codec LIKE '%H.264%')",
		},
		{
			name:       "trim spaces",
			fieldName:  "hdr",
			filter:     "HDR10+, Dolby Vision ",
			wantClause: "(hdr LIKE '%HDR10+%' OR hdr LIKE '%Dolby Vision%')",
		},
		{
			name:       "empty value",
			fieldName:  "resolution",
			filter:     "",
			wantClause: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildOrClause(tt.fieldName, tt.filter)
			if got != tt.wantClause {
				t.Fatalf("buildOrClause() got %q want %q", got, tt.wantClause)
			}
		})
	}
}

func TestGetMovies_DefaultPaginationAndPaging(t *testing.T) {
	db := setupTestDB(t)

	for i := 1; i <= 3; i++ {
		insertTestMovie(t, db, fmt.Sprintf("Movie %d", i), "available", int64(1000+i), fmt.Sprintf("/movies/%d.mkv", i))
	}

	filters := &models.FilterCriteria{}
	movies, total, err := GetMovies(db, filters)
	if err != nil {
		t.Fatalf("GetMovies returned error: %v", err)
	}
	if filters.Page != 1 || filters.PageSize != 50 {
		t.Fatalf("expected default page=1 and pageSize=50, got page=%d pageSize=%d", filters.Page, filters.PageSize)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	if len(movies) != 3 {
		t.Fatalf("expected len(movies)=3, got %d", len(movies))
	}

	pagedFilters := &models.FilterCriteria{Page: 2, PageSize: 2}
	pagedMovies, pagedTotal, err := GetMovies(db, pagedFilters)
	if err != nil {
		t.Fatalf("GetMovies paged returned error: %v", err)
	}
	if pagedTotal != 3 {
		t.Fatalf("expected paged total=3, got %d", pagedTotal)
	}
	if len(pagedMovies) != 1 {
		t.Fatalf("expected second page to contain 1 movie, got %d", len(pagedMovies))
	}
}

func TestGetMovies_FilterByResolution(t *testing.T) {
	db := setupTestDB(t)

	movie4KID := insertTestMovie(t, db, "Movie 4K", "available", 2001, "/movies/4k.mkv")
	movieHDID := insertTestMovie(t, db, "Movie HD", "available", 2002, "/movies/hd.mkv")

	if _, err := db.Exec("INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space) VALUES (?, ?, ?, ?, ?, ?, ?)", movie4KID, "H.265", "3840x2160", 24.0, "20000 kb/s", "HDR10+", "BT.2020"); err != nil {
		t.Fatalf("failed to insert 4k video track: %v", err)
	}
	if _, err := db.Exec("INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space) VALUES (?, ?, ?, ?, ?, ?, ?)", movieHDID, "H.264", "1920x1080", 24.0, "9000 kb/s", "", "BT.709"); err != nil {
		t.Fatalf("failed to insert hd video track: %v", err)
	}

	filters := &models.FilterCriteria{Resolution: "3840"}
	movies, total, err := GetMovies(db, filters)
	if err != nil {
		t.Fatalf("GetMovies returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1 for resolution filter, got %d", total)
	}
	if len(movies) != 1 {
		t.Fatalf("expected len(movies)=1, got %d", len(movies))
	}
	if movies[0].Title != "Movie 4K" {
		t.Fatalf("expected Movie 4K, got %s", movies[0].Title)
	}
}

func TestGetMovieByID_LoadsRelatedData(t *testing.T) {
	db := setupTestDB(t)

	movieID := insertTestMovie(t, db, "Detailed Movie", "available", 5001, "/movies/detailed.mkv")

	if _, err := db.Exec("INSERT INTO cast (movie_id, name, role, avatar) VALUES (?, ?, ?, ?)", movieID, "Actor One", "Lead", ""); err != nil {
		t.Fatalf("failed to insert cast row: %v", err)
	}
	if _, err := db.Exec("INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space) VALUES (?, ?, ?, ?, ?, ?, ?)", movieID, "H.265", "3840x2160", 24.0, "18000 kb/s", "HDR10", "BT.2020"); err != nil {
		t.Fatalf("failed to insert video track: %v", err)
	}
	if _, err := db.Exec("INSERT INTO audio_tracks (movie_id, codec, channels, language, sample_rate, bitrate) VALUES (?, ?, ?, ?, ?, ?)", movieID, "DTS", "5.1", "en", "48.0 kHz", "1500 kb/s"); err != nil {
		t.Fatalf("failed to insert audio track: %v", err)
	}
	if _, err := db.Exec("INSERT INTO subtitle_tracks (movie_id, language, format) VALUES (?, ?, ?)", movieID, "en", "SRT"); err != nil {
		t.Fatalf("failed to insert subtitle track: %v", err)
	}

	movie, err := GetMovieByID(db, movieID)
	if err != nil {
		t.Fatalf("GetMovieByID returned error: %v", err)
	}
	if movie.Title != "Detailed Movie" {
		t.Fatalf("unexpected movie title: got %q", movie.Title)
	}
	if len(movie.Cast) != 1 {
		t.Fatalf("expected one cast member, got %d", len(movie.Cast))
	}
	if movie.MediaInfo == nil {
		t.Fatalf("expected media info to be loaded")
	}
	if len(movie.MediaInfo.VideoTracks) != 1 || len(movie.MediaInfo.AudioTracks) != 1 || len(movie.MediaInfo.SubtitleTracks) != 1 {
		t.Fatalf("unexpected media info counts: video=%d audio=%d subs=%d", len(movie.MediaInfo.VideoTracks), len(movie.MediaInfo.AudioTracks), len(movie.MediaInfo.SubtitleTracks))
	}
}
