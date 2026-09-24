package repository

import "testing"

func TestMovieExistsByFilePath(t *testing.T) {
	db := setupTestDB(t)

	path := "/movies/sample.mkv"
	insertTestMovie(t, db, "Sample Movie", "available", 3001, path)

	exists, err := MovieExistsByFilePath(db, path)
	if err != nil {
		t.Fatalf("MovieExistsByFilePath returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected movie to exist")
	}

	notExists, err := MovieExistsByFilePath(db, "/movies/missing.mkv")
	if err != nil {
		t.Fatalf("MovieExistsByFilePath returned error for missing path: %v", err)
	}
	if notExists {
		t.Fatalf("expected movie not to exist")
	}
}

func TestEpisodeExistsByFilePath(t *testing.T) {
	db := setupTestDB(t)

	result, err := db.Exec(`
		INSERT INTO series (
			title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity,
			status, file_size, date_added, tvdb_id, imdb_id, poster
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "Sample Series", 2020, 2020, 1, 1, "", "", 7.0, 6.0, "complete", 2048, "2026-01-01", 4001, "tt7654321", "")
	if err != nil {
		t.Fatalf("failed to insert series: %v", err)
	}
	seriesID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get series id: %v", err)
	}

	episodePath := "/series/sample/S01E01.mkv"
	if _, err := db.Exec(`
		INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, seriesID, 1, 1, "Pilot", 45, "available", 1024, episodePath, "2026-01-01", "2026-01-01"); err != nil {
		t.Fatalf("failed to insert episode: %v", err)
	}

	exists, err := EpisodeExistsByFilePath(db, episodePath)
	if err != nil {
		t.Fatalf("EpisodeExistsByFilePath returned error: %v", err)
	}
	if !exists {
		t.Fatalf("expected episode to exist")
	}

	notExists, err := EpisodeExistsByFilePath(db, "/series/sample/S01E99.mkv")
	if err != nil {
		t.Fatalf("EpisodeExistsByFilePath returned error for missing path: %v", err)
	}
	if notExists {
		t.Fatalf("expected episode not to exist")
	}
}
