package repository

import (
	"database/sql"
	"time"

	"indexarr/internal/models"
)

// InsertMovie inserts a movie with its cast and media info in a transaction
func InsertMovie(db *sql.DB, movie *models.Movie) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert movie
	result, err := tx.Exec(`
		       INSERT INTO movies (title, year, duration, synopsis, genres, rating, popularity, status, file_size, file_path, container, date_added, last_scanned, tmdb_id, imdb_id, poster)
		       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	       `, movie.Title, movie.Year, movie.Duration, movie.Synopsis, movie.Genres, movie.Rating, movie.Popularity,
		movie.Status, movie.FileSize, movie.FilePath, movie.Container, movie.DateAdded, time.Now().Format(time.RFC3339),
		movie.TMDBId, movie.IMDbId, movie.Poster)
	if err != nil {
		return 0, err
	}

	movieID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert cast
	for _, cast := range movie.Cast {
		_, err := tx.Exec(`
			INSERT INTO cast (movie_id, name, role, avatar)
			VALUES (?, ?, ?, ?)
		`, movieID, cast.Name, cast.Role, cast.Avatar)
		if err != nil {
			return 0, err
		}
	}

	// Insert media info
	if movie.MediaInfo != nil {
		for _, vt := range movie.MediaInfo.VideoTracks {
			_, err := tx.Exec(`
				INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, movieID, vt.Codec, vt.Resolution, vt.FPS, vt.Bitrate, vt.HDR, vt.ColorSpace)
			if err != nil {
				return 0, err
			}
		}

		for _, at := range movie.MediaInfo.AudioTracks {
			_, err := tx.Exec(`
				INSERT INTO audio_tracks (movie_id, codec, channels, language, sample_rate, bitrate)
				VALUES (?, ?, ?, ?, ?, ?)
			`, movieID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
			if err != nil {
				return 0, err
			}
		}

		for _, st := range movie.MediaInfo.SubtitleTracks {
			_, err := tx.Exec(`
				INSERT INTO subtitle_tracks (movie_id, language, format)
				VALUES (?, ?, ?)
			`, movieID, st.Language, st.Format)
			if err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return movieID, nil
}

// InsertEpisode inserts an episode with its media info
func InsertEpisode(db *sql.DB, episode *models.Episode) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, episode.SeriesID, episode.SeasonNum, episode.EpisodeNum, episode.Title, episode.Duration,
		episode.Status, episode.FileSize, episode.FilePath, episode.DateAdded, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}

	episodeID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert media info
	if episode.MediaInfo != nil {
		for _, vt := range episode.MediaInfo.VideoTracks {
			_, err := tx.Exec(`
				INSERT INTO video_tracks (episode_id, codec, resolution, fps, bitrate, hdr, color_space)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, episodeID, vt.Codec, vt.Resolution, vt.FPS, vt.Bitrate, vt.HDR, vt.ColorSpace)
			if err != nil {
				return 0, err
			}
		}

		for _, at := range episode.MediaInfo.AudioTracks {
			_, err := tx.Exec(`
				INSERT INTO audio_tracks (episode_id, codec, channels, language, sample_rate, bitrate)
				VALUES (?, ?, ?, ?, ?, ?)
			`, episodeID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
			if err != nil {
				return 0, err
			}
		}

		for _, st := range episode.MediaInfo.SubtitleTracks {
			_, err := tx.Exec(`
				INSERT INTO subtitle_tracks (episode_id, language, format)
				VALUES (?, ?, ?)
			`, episodeID, st.Language, st.Format)
			if err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return episodeID, nil
}

// InsertSeries inserts a series (without episodes)
func InsertSeries(db *sql.DB, series *models.Series) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO series (title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id, poster)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, series.Title, series.YearStart, series.YearEnd, series.SeasonCount, series.EpisodeCount,
		series.Synopsis, series.Genres, series.Rating, series.Popularity, series.Status,
		series.FileSize, series.DateAdded, series.TVDBId, series.IMDbId, series.Poster)
	if err != nil {
		return 0, err
	}

	seriesID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert cast
	for _, cast := range series.Cast {
		_, err := tx.Exec(`
			INSERT INTO cast (series_id, name, role, avatar)
			VALUES (?, ?, ?, ?)
		`, seriesID, cast.Name, cast.Role, cast.Avatar)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return seriesID, nil
}

// GetOrCreateSeason returns existing season or creates a new one
func GetOrCreateSeason(db *sql.DB, seriesID int64, seasonNum int) (int64, error) {
	var seasonID int64
	err := db.QueryRow(`SELECT id FROM seasons WHERE series_id = ? AND number = ?`, seriesID, seasonNum).Scan(&seasonID)
	if err == nil {
		return seasonID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new season
	result, err := db.Exec(`INSERT INTO seasons (series_id, number, file_size) VALUES (?, ?, 0)`, seriesID, seasonNum)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetEpisodeBySeriesSeasonEpisode finds an episode by series ID, season, and episode number
func GetEpisodeBySeriesSeasonEpisode(db *sql.DB, seriesID int64, seasonNum, episodeNum int) (*models.Episode, error) {
	var episode models.Episode
	err := db.QueryRow(`
		SELECT id, series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned
		FROM episodes
		WHERE series_id = ? AND season_num = ? AND episode_num = ?
	`, seriesID, seasonNum, episodeNum).Scan(&episode.ID, &episode.SeriesID, &episode.SeasonNum, &episode.EpisodeNum,
		&episode.Title, &episode.Duration, &episode.Status, &episode.FileSize, &episode.FilePath, &episode.DateAdded, &episode.LastScanned)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

// UpdateEpisode updates an existing episode
func UpdateEpisode(db *sql.DB, episode *models.Episode) error {
	_, err := db.Exec(`
		UPDATE episodes
		SET title = ?, duration = ?, status = ?, file_size = ?, file_path = ?, last_scanned = ?
		WHERE id = ?
	`, episode.Title, episode.Duration, episode.Status, episode.FileSize, episode.FilePath, time.Now().Format(time.RFC3339), episode.ID)
	return err
}

// GetSeriesByTitle finds a series by title (case-insensitive)
func GetSeriesByTitle(db *sql.DB, title string) (*models.Series, error) {
	var series models.Series
	var poster sql.NullString
	err := db.QueryRow(`
		SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id, poster
		FROM series WHERE LOWER(title) = LOWER(?)
	`, title).Scan(&series.ID, &series.Title, &series.YearStart, &series.YearEnd, &series.SeasonCount, &series.EpisodeCount,
		&series.Synopsis, &series.Genres, &series.Rating, &series.Popularity, &series.Status,
		&series.FileSize, &series.DateAdded, &series.TVDBId, &series.IMDbId, &poster)
	if poster.Valid {
		series.Poster = &poster.String
	} else {
		series.Poster = nil
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &series, nil
}

// GetSeriesByTVDBId finds a series by TVDB ID
func GetSeriesByTVDBId(db *sql.DB, tvdbID int64) (*models.Series, error) {
	var series models.Series
	var poster sql.NullString
	err := db.QueryRow(`
		SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id, poster
		FROM series WHERE tvdb_id = ?
	`, tvdbID).Scan(&series.ID, &series.Title, &series.YearStart, &series.YearEnd, &series.SeasonCount, &series.EpisodeCount,
		&series.Synopsis, &series.Genres, &series.Rating, &series.Popularity, &series.Status,
		&series.FileSize, &series.DateAdded, &series.TVDBId, &series.IMDbId, &poster)
	if poster.Valid {
		series.Poster = &poster.String
	} else {
		series.Poster = nil
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &series, nil
}

// UpdateSeriesCounts updates season_count and episode_count for a series
func UpdateSeriesCounts(db *sql.DB, seriesID int64) error {
	_, err := db.Exec(`
		UPDATE series SET
			season_count = (SELECT COUNT(DISTINCT season_num) FROM episodes WHERE series_id = ?),
			episode_count = (SELECT COUNT(*) FROM episodes WHERE series_id = ?)
		WHERE id = ?
	`, seriesID, seriesID, seriesID)
	return err
}

// Scan status operations

// GetScanStatus returns the current scan status (creates if not exists)
func GetScanStatus(db *sql.DB) (*models.ScanStatus, error) {
	var status models.ScanStatus
	var startedAt, completedAt, errorMsg sql.NullString

	err := db.QueryRow(`SELECT id, status, started_at, completed_at, files_found, files_processed, error_message FROM scan_status WHERE id = 1`).
		Scan(&status.ID, &status.Status, &startedAt, &completedAt, &status.FilesFound, &status.FilesProcessed, &errorMsg)

	if err == sql.ErrNoRows {
		// Initialize default status
		_, err := db.Exec(`INSERT INTO scan_status (id, status, files_found, files_processed) VALUES (1, 'idle', 0, 0)`)
		if err != nil {
			return nil, err
		}
		return &models.ScanStatus{ID: 1, Status: "idle"}, nil
	}
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		status.StartedAt = startedAt.String
	}
	if completedAt.Valid {
		status.CompletedAt = completedAt.String
	}
	if errorMsg.Valid {
		status.ErrorMessage = errorMsg.String
	}

	return &status, nil
}

// UpdateScanStatus updates the scan status
func UpdateScanStatus(db *sql.DB, status *models.ScanStatus) error {
	_, err := db.Exec(`
		UPDATE scan_status SET status = ?, started_at = ?, completed_at = ?, files_found = ?, files_processed = ?, error_message = ?
		WHERE id = 1
	`, status.Status, nullString(status.StartedAt), nullString(status.CompletedAt), status.FilesFound, status.FilesProcessed, nullString(status.ErrorMessage))
	return err
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// MovieExistsByPath checks if a movie with the given file path exists
func MovieExistsByPath(db *sql.DB, filePath string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM movies WHERE file_path = ?`, filePath).Scan(&count)
	return count > 0, err
}

// EpisodeExistsByPath checks if an episode with the given file path exists
func EpisodeExistsByPath(db *sql.DB, filePath string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM episodes WHERE file_path = ?`, filePath).Scan(&count)
	return count > 0, err
}

// PurgeDatabase deletes all content from the database
func PurgeDatabase(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete in reverse order of foreign key dependencies
	tables := []string{
		"subtitle_tracks",
		"audio_tracks",
		"video_tracks",
		"cast",
		"episodes",
		"seasons",
		"series",
		"movies",
	}

	for _, table := range tables {
		_, err := tx.Exec("DELETE FROM " + table)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
