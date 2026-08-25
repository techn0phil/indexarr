package repository

import "testing"

func TestGetOrCreateSeasonAndUpdateSeriesCounts(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	seriesID := insertTestSeries(t, db, "Counted Series", "ongoing", 5001, 6001)

	seasonID, err := GetOrCreateSeason(db, seriesID, 1)
	if err != nil {
		t.Fatalf("GetOrCreateSeason create returned error: %v", err)
	}
	seasonID2, err := GetOrCreateSeason(db, seriesID, 1)
	if err != nil {
		t.Fatalf("GetOrCreateSeason fetch returned error: %v", err)
	}
	if seasonID != seasonID2 {
		t.Fatalf("expected same season id on repeated call, got %d vs %d", seasonID, seasonID2)
	}

	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, 'S1E1', 45, 'available', 1024, '/counted/s01e01.mkv', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert available episode: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 2, 'S1E2', 45, 'missing', 0, '', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert missing episode: %v", err)
	}

	if err := UpdateSeriesCounts(db, seriesID); err != nil {
		t.Fatalf("UpdateSeriesCounts returned error: %v", err)
	}

	var seasonCount, episodeCount, missingCount int
	var fileSize int64
	if err := db.QueryRow("SELECT season_count, episode_count, missing_episode_count, file_size FROM series WHERE id = ?", seriesID).Scan(&seasonCount, &episodeCount, &missingCount, &fileSize); err != nil {
		t.Fatalf("failed to read updated series counts: %v", err)
	}
	if seasonCount != 1 || episodeCount != 2 || missingCount != 1 || fileSize != 1024 {
		t.Fatalf("unexpected series aggregate values: seasons=%d episodes=%d missing=%d fileSize=%d", seasonCount, episodeCount, missingCount, fileSize)
	}
}

func TestDeleteEmptySeasonsAndDeleteEmptySeries(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	seriesWithEpisodes := insertTestSeries(t, db, "With Episodes", "complete", 5101, 6101)
	emptySeries := insertTestSeries(t, db, "Empty Series", "complete", 5102, 6102)

	if _, err := db.Exec("INSERT INTO seasons (series_id, number, file_size) VALUES (?, 1, 0), (?, 2, 0)", seriesWithEpisodes, seriesWithEpisodes); err != nil {
		t.Fatalf("failed to insert seasons: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, 'S1E1', 45, 'available', 100, '/with/s01e01.mkv', '2026-01-01', '2026-01-01')", seriesWithEpisodes); err != nil {
		t.Fatalf("failed to insert episode: %v", err)
	}

	if err := DeleteEmptySeasons(db, seriesWithEpisodes); err != nil {
		t.Fatalf("DeleteEmptySeasons returned error: %v", err)
	}
	var seasonsCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM seasons WHERE series_id = ?", seriesWithEpisodes).Scan(&seasonsCount); err != nil {
		t.Fatalf("failed to count seasons: %v", err)
	}
	if seasonsCount != 1 {
		t.Fatalf("expected only one non-empty season to remain, got %d", seasonsCount)
	}

	if err := DeleteEmptySeries(db); err != nil {
		t.Fatalf("DeleteEmptySeries returned error: %v", err)
	}
	var remaining int
	if err := db.QueryRow("SELECT COUNT(*) FROM series WHERE id = ?", emptySeries).Scan(&remaining); err != nil {
		t.Fatalf("failed to check empty series deletion: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected empty series to be deleted")
	}
}

func TestDeleteHelpersAndPurgeDatabase(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	movieID := insertTestMovie(t, db, "Delete Me", "available", 5201, "/library/movies/delete-me.mkv")
	movieByTMDB := insertTestMovie(t, db, "Delete By TMDB", "available", 5202, "/library/movies/tmdb.mkv")

	seriesID := insertTestSeries(t, db, "Delete Series", "complete", 5301, 6301)
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, 'S1E1', 45, 'available', 100, '/library/series/delete-series/s01e01.mkv', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert episode for delete test: %v", err)
	}

	if err := DeleteMovie(db, movieID); err != nil {
		t.Fatalf("DeleteMovie returned error: %v", err)
	}
	if err := DeleteMovieByTMDBId(db, 5202); err != nil {
		t.Fatalf("DeleteMovieByTMDBId returned error: %v", err)
	}
	_ = movieByTMDB

	seriesByPath := insertTestSeries(t, db, "Series By Path", "complete", 5302, 6302)
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, 'S1E1', 45, 'available', 100, '/library/series/by-path/s01e01.mkv', '2026-01-01', '2026-01-01')", seriesByPath); err != nil {
		t.Fatalf("failed to insert episode for path delete test: %v", err)
	}
	if err := DeleteEpisodeByPath(db, "%/by-path/%"); err != nil {
		t.Fatalf("DeleteEpisodeByPath returned error: %v", err)
	}

	if err := DeleteSeriesBySonarrID(db, 6301); err != nil {
		t.Fatalf("DeleteSeriesBySonarrID returned error: %v", err)
	}

	if err := PurgeDatabase(db); err != nil {
		t.Fatalf("PurgeDatabase returned error: %v", err)
	}

	tables := []string{"movies", "series", "seasons", "episodes", "cast", "video_tracks", "audio_tracks", "subtitle_tracks"}
	for _, table := range tables {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatalf("failed counting table %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected table %s to be empty after purge, found %d rows", table, count)
		}
	}
}
