package services

import (
	"os"
	"path/filepath"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
)

func TestScannerPathMatchingHelpers(t *testing.T) {
	s := &Scanner{config: &config.Config{
		MediaLibraryPaths:  []string{"/library/media"},
		MoviesLibraryPaths: []string{"/library/movies"},
		SeriesLibraryPaths: []string{"/library/series"},
	}}

	if !s.pathsMatchAnyLibrary([]string{"/library/movies/"}) {
		t.Fatalf("expected movie library path to match")
	}
	if s.pathsMatchAnyLibrary([]string{"/other/path"}) {
		t.Fatalf("expected non-library path not to match")
	}

	if !s.pathsMatchMovieOrMediaLibrary([]string{"/library/media"}) {
		t.Fatalf("expected media path to match movie-or-media helper")
	}
	if s.pathsMatchMovieOrMediaLibrary([]string{"/library/series"}) {
		t.Fatalf("expected series-only path not to match movie-or-media helper")
	}

	if !s.pathsMatchSeriesOrMediaLibrary([]string{"/library/series"}) {
		t.Fatalf("expected series path to match series-or-media helper")
	}
	if s.pathsMatchSeriesOrMediaLibrary([]string{"/library/movies"}) {
		t.Fatalf("expected movie-only path not to match series-or-media helper")
	}
}

func TestFindCommonParentFolder(t *testing.T) {
	folders := []string{
		"/library/series/Show A/Season 01",
		"/library/series/Show A/Season 02",
	}
	if got := findCommonParentFolder(folders); got != "/library/series/Show A" {
		t.Fatalf("findCommonParentFolder() = %q, want %q", got, "/library/series/Show A")
	}

	if got := findCommonParentFolder(nil); got != "" {
		t.Fatalf("findCommonParentFolder(nil) = %q, want empty", got)
	}
}

func TestFilesystemScannersCountAndStatus(t *testing.T) {
	db := setupServicesTestDBWithMigrations(t)

	root := t.TempDir()
	movieDir := filepath.Join(root, "movies")
	seriesDir := filepath.Join(root, "series")
	if err := os.MkdirAll(movieDir, 0o755); err != nil {
		t.Fatalf("failed creating movie dir: %v", err)
	}
	if err := os.MkdirAll(seriesDir, 0o755); err != nil {
		t.Fatalf("failed creating series dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(movieDir, "Movie.2025.mkv"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed writing movie file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seriesDir, "Show.S01E01.mp4"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed writing series file: %v", err)
	}

	cfg := &config.Config{
		MediainfoPath:      "mediainfo",
		ScanTimeout:        1,
		MoviesLibraryPaths: []string{movieDir},
		SeriesLibraryPaths: []string{seriesDir},
	}

	movieScanner := NewFilesystemMovieScanner(db, cfg, nil)
	seriesScanner := NewFilesystemSeriesScanner(db, cfg, nil)

	movieCount, err := movieScanner.GetPendingFileCount()
	if err != nil {
		t.Fatalf("movie GetPendingFileCount() error = %v", err)
	}
	if movieCount != 1 {
		t.Fatalf("movie count = %d, want 1", movieCount)
	}

	seriesCount, err := seriesScanner.GetPendingFileCount()
	if err != nil {
		t.Fatalf("series GetPendingFileCount() error = %v", err)
	}
	if seriesCount != 1 {
		t.Fatalf("series count = %d, want 1", seriesCount)
	}

	status, err := movieScanner.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status == "" {
		t.Fatalf("expected non-empty status")
	}
}

func TestScannerRemoveStaleMediaFilesDeletesMissingMovieAndEpisode(t *testing.T) {
	db := setupServicesTestDBWithMigrations(t)

	root := t.TempDir()
	movieExisting := filepath.Join(root, "movies", "Keep.2024.mkv")
	movieMissing := filepath.Join(root, "movies", "Missing.2020.mkv")
	episodeExisting := filepath.Join(root, "series", "Show", "S01E01.mkv")
	episodeMissing := filepath.Join(root, "series", "Show", "S01E02.mkv")

	for _, p := range []string{movieExisting, episodeExisting} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("failed to create parent dir for %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("failed to create file %s: %v", p, err)
		}
	}

	if _, err := repository.InsertMovie(db, &models.Movie{
		Title:     "Keep Movie",
		Year:      2024,
		Status:    "available",
		FilePath:  movieExisting,
		DateAdded: "2024-01-01T00:00:00Z",
		TMDBId:    101,
	}); err != nil {
		t.Fatalf("InsertMovie(existing) error = %v", err)
	}
	if _, err := repository.InsertMovie(db, &models.Movie{
		Title:     "Missing Movie",
		Year:      2020,
		Status:    "available",
		FilePath:  movieMissing,
		DateAdded: "2024-01-01T00:00:00Z",
		TMDBId:    102,
	}); err != nil {
		t.Fatalf("InsertMovie(missing) error = %v", err)
	}

	seriesID, err := repository.InsertSeries(db, &models.Series{
		Title:             "Show",
		Slug:              "show",
		YearStart:         2022,
		SeasonCount:       1,
		EpisodeCount:      2,
		TotalSeasonCount:  1,
		TotalEpisodeCount: 2,
		Status:            "ongoing",
		DateAdded:         "2024-01-01T00:00:00Z",
		TMDBId:            1,
		TVDBId:            2,
	})
	if err != nil {
		t.Fatalf("InsertSeries() error = %v", err)
	}

	if _, err := repository.InsertEpisode(db, &models.Episode{
		SeriesID:   seriesID,
		SeasonNum:  1,
		EpisodeNum: 1,
		Title:      "Keep Episode",
		Status:     "available",
		FilePath:   episodeExisting,
		DateAdded:  "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("InsertEpisode(existing) error = %v", err)
	}
	if _, err := repository.InsertEpisode(db, &models.Episode{
		SeriesID:   seriesID,
		SeasonNum:  1,
		EpisodeNum: 2,
		Title:      "Missing Episode",
		Status:     "available",
		FilePath:   episodeMissing,
		DateAdded:  "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("InsertEpisode(missing) error = %v", err)
	}

	s := &Scanner{db: db, config: &config.Config{MediaLibraryPaths: []string{root}}}
	deletedMovies, deletedEpisodes, err := s.removeStaleMediaFiles([]string{root}, []string{movieExisting, episodeExisting})
	if err != nil {
		t.Fatalf("removeStaleMediaFiles() error = %v", err)
	}
	if deletedMovies != 1 {
		t.Fatalf("deletedMovies = %d, want 1", deletedMovies)
	}
	if deletedEpisodes != 1 {
		t.Fatalf("deletedEpisodes = %d, want 1", deletedEpisodes)
	}

	movieStillExists, err := repository.MovieExistsByPath(db, movieExisting)
	if err != nil {
		t.Fatalf("MovieExistsByPath(existing) error = %v", err)
	}
	if !movieStillExists {
		t.Fatalf("expected existing movie to remain")
	}

	movieMissingExists, err := repository.MovieExistsByPath(db, movieMissing)
	if err != nil {
		t.Fatalf("MovieExistsByPath(missing) error = %v", err)
	}
	if movieMissingExists {
		t.Fatalf("expected missing movie to be removed")
	}

	episodeStillExists, err := repository.EpisodeExistsByPath(db, episodeExisting)
	if err != nil {
		t.Fatalf("EpisodeExistsByPath(existing) error = %v", err)
	}
	if !episodeStillExists {
		t.Fatalf("expected existing episode to remain")
	}

	episodeMissingExists, err := repository.EpisodeExistsByPath(db, episodeMissing)
	if err != nil {
		t.Fatalf("EpisodeExistsByPath(missing) error = %v", err)
	}
	if episodeMissingExists {
		t.Fatalf("expected missing episode to be removed")
	}
}

func TestScannerRemoveStaleMediaFilesSkipsWhenPathsDoNotMatchLibraries(t *testing.T) {
	db := setupServicesTestDBWithMigrations(t)

	root := t.TempDir()
	movieMissing := filepath.Join(root, "movies", "Missing.2020.mkv")

	if _, err := repository.InsertMovie(db, &models.Movie{
		Title:     "Missing Movie",
		Year:      2020,
		Status:    "available",
		FilePath:  movieMissing,
		DateAdded: "2024-01-01T00:00:00Z",
		TMDBId:    202,
	}); err != nil {
		t.Fatalf("InsertMovie() error = %v", err)
	}

	s := &Scanner{db: db, config: &config.Config{MediaLibraryPaths: []string{filepath.Join(root, "library")}}}
	deletedMovies, deletedEpisodes, err := s.removeStaleMediaFiles([]string{filepath.Join(root, "other")}, nil)
	if err != nil {
		t.Fatalf("removeStaleMediaFiles() error = %v", err)
	}
	if deletedMovies != 0 || deletedEpisodes != 0 {
		t.Fatalf("unexpected deletions when path mismatch: movies=%d episodes=%d", deletedMovies, deletedEpisodes)
	}

	exists, err := repository.MovieExistsByPath(db, movieMissing)
	if err != nil {
		t.Fatalf("MovieExistsByPath() error = %v", err)
	}
	if !exists {
		t.Fatalf("expected movie to remain when scanned paths do not match library")
	}
}
