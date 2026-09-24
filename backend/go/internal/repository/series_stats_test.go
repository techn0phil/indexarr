package repository

import (
	"testing"

	"indexarr/internal/models"
)

func TestGetSeries_DefaultPaginationAndStatusFilter(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	insertTestSeries(t, db, "Alpha Series", "ongoing", 3001, 4001)
	insertTestSeries(t, db, "Beta Series", "complete", 3002, 4002)

	filters := &models.FilterCriteria{}
	list, total, err := GetSeries(db, filters)
	if err != nil {
		t.Fatalf("GetSeries returned error: %v", err)
	}
	if filters.Page != 1 || filters.PageSize != 50 {
		t.Fatalf("expected default pagination page=1 pageSize=50, got page=%d pageSize=%d", filters.Page, filters.PageSize)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(list) != 2 {
		t.Fatalf("expected len(series)=2, got %d", len(list))
	}

	filtered, filteredTotal, err := GetSeries(db, &models.FilterCriteria{Status: "ongoing"})
	if err != nil {
		t.Fatalf("GetSeries with status filter returned error: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 {
		t.Fatalf("expected one ongoing series, got total=%d len=%d", filteredTotal, len(filtered))
	}
	if filtered[0].Title != "Alpha Series" {
		t.Fatalf("expected Alpha Series, got %q", filtered[0].Title)
	}
}

func TestGetSeries_FilterByEpisodeResolution(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	series4K := insertTestSeries(t, db, "Series 4K", "complete", 3101, 4101)
	seriesHD := insertTestSeries(t, db, "Series HD", "complete", 3102, 4102)

	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, ?, 45, 'available', 1000, ?, '2026-01-01', '2026-01-01')", series4K, "Pilot 4K", "/series4k/s01e01.mkv"); err != nil {
		t.Fatalf("failed to insert 4k episode: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, ?, 45, 'available', 1000, ?, '2026-01-01', '2026-01-01')", seriesHD, "Pilot HD", "/serieshd/s01e01.mkv"); err != nil {
		t.Fatalf("failed to insert hd episode: %v", err)
	}

	var ep4KID, epHDID int64
	if err := db.QueryRow("SELECT id FROM episodes WHERE series_id = ?", series4K).Scan(&ep4KID); err != nil {
		t.Fatalf("failed to load 4k episode id: %v", err)
	}
	if err := db.QueryRow("SELECT id FROM episodes WHERE series_id = ?", seriesHD).Scan(&epHDID); err != nil {
		t.Fatalf("failed to load hd episode id: %v", err)
	}

	if _, err := db.Exec("INSERT INTO video_tracks (episode_id, codec, resolution, fps, bitrate, hdr, color_space) VALUES (?, 'H.265', '3840x2160', 24.0, '15000 kb/s', 'HDR10+', 'BT.2020')", ep4KID); err != nil {
		t.Fatalf("failed to insert 4k track: %v", err)
	}
	if _, err := db.Exec("INSERT INTO video_tracks (episode_id, codec, resolution, fps, bitrate, hdr, color_space) VALUES (?, 'H.264', '1920x1080', 24.0, '8000 kb/s', '', 'BT.709')", epHDID); err != nil {
		t.Fatalf("failed to insert hd track: %v", err)
	}

	series, total, err := GetSeries(db, &models.FilterCriteria{Resolution: "3840"})
	if err != nil {
		t.Fatalf("GetSeries with resolution filter returned error: %v", err)
	}
	if total != 1 || len(series) != 1 {
		t.Fatalf("expected one 4k series, got total=%d len=%d", total, len(series))
	}
	if series[0].Title != "Series 4K" {
		t.Fatalf("expected Series 4K, got %q", series[0].Title)
	}
}

func TestGetSeriesByID_LoadsSeasonsAndEpisodes(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	seriesID := insertTestSeries(t, db, "Detailed Series", "ongoing", 3201, 4201)
	if _, err := db.Exec("INSERT INTO seasons (series_id, number, file_size) VALUES (?, 1, 2000)", seriesID); err != nil {
		t.Fatalf("failed to insert season: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, 'Episode 1', 42, 'available', 1000, '/detailed/s01e01.mkv', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert available episode: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 2, 'Episode 2', 42, 'missing', 0, '', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert missing episode: %v", err)
	}

	var seasonNumber int
	if err := db.QueryRow("SELECT number FROM seasons WHERE series_id = ?", seriesID).Scan(&seasonNumber); err != nil {
		t.Fatalf("failed to read inserted season number: %v", err)
	}
	episodesForSeason, err := GetEpisodesForSeason(db, seriesID, seasonNumber)
	if err != nil {
		t.Fatalf("GetEpisodesForSeason returned error: %v", err)
	}
	if len(episodesForSeason) != 2 {
		t.Fatalf("expected two episodes from GetEpisodesForSeason, got %d", len(episodesForSeason))
	}

	series, err := GetSeriesByID(db, seriesID)
	if err != nil {
		t.Fatalf("GetSeriesByID returned error: %v", err)
	}
	if series.Title != "Detailed Series" {
		t.Fatalf("unexpected title: got %q", series.Title)
	}
	if len(series.Seasons) != 1 {
		t.Fatalf("expected one season, got %d", len(series.Seasons))
	}
	if series.Seasons[0].Number != 1 {
		t.Fatalf("expected loaded season number 1, got %d", series.Seasons[0].Number)
	}
}

func TestGetStats_ComputesExpectedAggregates(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	movieAvailID := insertTestMovie(t, db, "Movie A", "available", 7001, "/movies/a.mkv")
	insertTestMovie(t, db, "Movie B", "missing", 7002, "/movies/b.mkv")
	if _, err := db.Exec("UPDATE movies SET file_size = 1073741824 WHERE id = ?", movieAvailID); err != nil {
		t.Fatalf("failed to update movie size: %v", err)
	}
	if _, err := db.Exec("INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space) VALUES (?, 'H.265', '3840x2160', 24.0, '20000 kb/s', 'HDR10', 'BT.2020')", movieAvailID); err != nil {
		t.Fatalf("failed to insert movie video track: %v", err)
	}

	seriesID := insertTestSeries(t, db, "Series Stats", "complete", 7301, 8301)
	if _, err := db.Exec("UPDATE series SET total_episode_count = 2 WHERE id = ?", seriesID); err != nil {
		t.Fatalf("failed to set series episode total: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 1, 'S1E1', 45, 'available', 2147483648, '/series/s01e01.mkv', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert available episode: %v", err)
	}
	if _, err := db.Exec("INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned) VALUES (?, 1, 2, 'S1E2', 45, 'missing', 0, '', '2026-01-01', '2026-01-01')", seriesID); err != nil {
		t.Fatalf("failed to insert missing episode: %v", err)
	}

	stats, err := GetStats(db)
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}

	if stats.TotalMovies != 2 || stats.TotalSeries != 1 || stats.TotalEpisodes != 2 {
		t.Fatalf("unexpected totals: movies=%d series=%d episodes=%d", stats.TotalMovies, stats.TotalSeries, stats.TotalEpisodes)
	}
	if stats.AvailMovies != 1 || stats.MissingMovies != 1 {
		t.Fatalf("unexpected movie availability counts: avail=%d missing=%d", stats.AvailMovies, stats.MissingMovies)
	}
	if stats.AvailEpisodes != 1 || stats.MissingEpisodes != 1 {
		t.Fatalf("unexpected episode availability counts: avail=%d missing=%d", stats.AvailEpisodes, stats.MissingEpisodes)
	}
	if stats.ProblemsCount != 2 {
		t.Fatalf("expected problems count 2, got %d", stats.ProblemsCount)
	}
	if stats.FourKCount != 1 || stats.FourKPercent != 50 {
		t.Fatalf("unexpected 4k stats: count=%d percent=%f", stats.FourKCount, stats.FourKPercent)
	}
	if stats.MoviesDiskSpaceGB != 1 || stats.SeriesDiskSpaceGB != 2 || stats.DiskSpaceGB != 3 {
		t.Fatalf("unexpected disk stats: movies=%f series=%f total=%f", stats.MoviesDiskSpaceGB, stats.SeriesDiskSpaceGB, stats.DiskSpaceGB)
	}
}
