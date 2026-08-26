package repository

import (
	"database/sql"
	_ "embed"
	"os"
	"path/filepath"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var testSchemaSQL string

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if _, err := db.Exec(testSchemaSQL); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	return db
}

func setupTestDBWithMigrations(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	migrationFiles, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		t.Fatalf("failed to list migration files: %v", err)
	}
	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", migrationFile, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("failed to execute migration %s: %v", migrationFile, err)
		}
	}

	return db
}

func insertTestMovie(t *testing.T, db *sql.DB, title string, status string, tmdbID int64, filePath string) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO movies (
			title, year, duration, synopsis, genres, rating, popularity, status,
			file_size, file_path, container, date_added, last_scanned, tmdb_id, imdb_id, poster
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, title, 2020, 100, "synopsis", "genre", 8.0, 7.0, status, 1024, filePath, "mkv", "2026-01-01", "2026-01-01", tmdbID, "tt1234567", "")
	if err != nil {
		t.Fatalf("failed to insert movie: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get movie last insert id: %v", err)
	}
	return id
}

func insertTestSeries(t *testing.T, db *sql.DB, title string, status string, tmdbID int64, sonarrID int64) int64 {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO series (
			title, year_start, year_end, season_count, episode_count, missing_episode_count,
			total_season_count, total_episode_count, synopsis, genres, rating, popularity,
			status, file_size, date_added, tmdb_id, tvdb_id, imdb_id, poster, slug, sonarr_id, title_slug
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, title, 2021, 2024, 0, 0, 0, 0, 0, "series synopsis", "drama", 8.2, 10.0, status, 0, "2026-01-01", tmdbID, tmdbID+1000, "tt7654321", "", "", sonarrID, "")
	if err != nil {
		t.Fatalf("failed to insert series: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get series last insert id: %v", err)
	}
	return id
}
