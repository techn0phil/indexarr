package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"

	"github.com/rs/zerolog"
)

func TestRadarrImporterMapHelpers(t *testing.T) {
	importer := &RadarrImporter{
		client:   &RadarrClient{baseURL: "http://radarr.local"},
		pathFrom: "/radarr",
		pathTo:   "/local",
	}

	rm := &RadarrMovie{
		Title:    "Mapped Movie",
		Year:     2024,
		Overview: "overview",
		Runtime:  120,
		Genres:   []string{"Action", "Sci-Fi"},
		Ratings: RadarrRatings{
			Imdb: RadarrRating{Value: 7.2},
		},
		Images: []RadarrImage{{CoverType: "poster", URL: "/MediaCover/55/poster.jpg"}},
		TmdbID: 321,
		ImdbID: "tt123",
		Added:  "",
		MovieFile: &RadarrMovieFile{
			Path: "/radarr/Movies/Mapped Movie (2024)/movie.mkv",
			Size: 100,
		},
	}

	if got := importer.mapPath("/radarr/Movies/Mapped Movie (2024)/movie.mkv"); got != "/local/Movies/Mapped Movie (2024)/movie.mkv" {
		t.Fatalf("mapPath() = %q", got)
	}

	movie := importer.mapRadarrMovie(rm)
	if movie.TMDBId != 321 {
		t.Fatalf("TMDBId = %d, want 321", movie.TMDBId)
	}
	if movie.Rating != 7.2 {
		t.Fatalf("Rating = %v, want imdb fallback 7.2", movie.Rating)
	}
	if movie.Container != "mkv" {
		t.Fatalf("Container = %q, want mkv", movie.Container)
	}
	if movie.Poster == nil || *movie.Poster != "http://radarr.local/MediaCover/55/poster.jpg" {
		t.Fatalf("Poster = %v, want local URL expanded", movie.Poster)
	}
	if movie.DateAdded == "" {
		t.Fatalf("expected non-empty DateAdded fallback")
	}
}

func TestRadarrImporterImportUsesCacheAndRemovesStale(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	db := setupServicesTestDBWithMigrations(t)

	stale := &models.Movie{
		Title:     "Stale",
		Year:      2001,
		Status:    "available",
		FilePath:  "/does/not/matter.mkv",
		DateAdded: "2024-01-01T00:00:00Z",
		TMDBId:    999,
	}
	if _, err := repository.InsertMovie(db, stale); err != nil {
		t.Fatalf("InsertMovie(stale) error = %v", err)
	}

	movieCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/movie":
			movieCalls++
			fmt.Fprint(w, `[
				{"id":10,"title":"Imported Movie","year":2025,"overview":"plot","runtime":95,"hasFile":true,"genres":["Drama"],"ratings":{"tmdb":{"value":8.4}},"images":[{"coverType":"poster","remoteUrl":"https://img.example/poster.jpg"}],"tmdbId":12345,"imdbId":"tt456","added":"2025-05-01T10:00:00Z","movieFile":{"path":"","size":555}},
				{"id":11,"title":"No File","hasFile":false}
			]`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer srv.Close()

	client := NewRadarrClient(srv.URL, "radarr-key")
	cfg := &config.Config{MediainfoPath: "mediainfo", ScanTimeout: 1}
	importer := NewRadarrImporter(db, cfg, client, nil)

	count, err := importer.GetPendingFileCount()
	if err != nil {
		t.Fatalf("GetPendingFileCount() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("GetPendingFileCount() = %d, want 1", count)
	}

	result, err := importer.Import(nil)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.FilesFound != 1 || result.FilesProcessed != 1 {
		t.Fatalf("unexpected result counters: %+v", result)
	}
	if result.MoviesAdded != 1 {
		t.Fatalf("MoviesAdded = %d, want 1", result.MoviesAdded)
	}

	if movieCalls != 1 {
		t.Fatalf("expected one /movie API call due to cache use, got %d", movieCalls)
	}

	imported, err := repository.GetMovieByTMDBId(db, 12345)
	if err != nil {
		t.Fatalf("GetMovieByTMDBId(imported) error = %v", err)
	}
	if imported == nil {
		t.Fatalf("expected imported movie")
	}
	if imported.Poster == nil || *imported.Poster != "https://img.example/poster.jpg" {
		t.Fatalf("imported poster = %v", imported.Poster)
	}

	deletedStale, err := repository.GetMovieByTMDBId(db, 999)
	if err != nil {
		t.Fatalf("GetMovieByTMDBId(stale) error = %v", err)
	}
	if deletedStale != nil {
		t.Fatalf("expected stale movie to be deleted, got %+v", deletedStale)
	}
}
