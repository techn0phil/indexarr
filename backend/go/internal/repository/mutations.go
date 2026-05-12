package repository

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"indexarr/internal/models"
)

// retryOnLock retries a function up to 3 times if database is locked
func retryOnLock(fn func() error) error {
	var lastErr error
	backoffs := []time.Duration{100 * time.Millisecond, 300 * time.Millisecond, 1 * time.Second}

	for attempt := 0; attempt < len(backoffs)+1; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Check if it's a "database is locked" error
		if strings.Contains(err.Error(), "database is locked") {
			lastErr = err
			if attempt < len(backoffs) {
				log.Printf("Database locked, retrying in %v (attempt %d/3)", backoffs[attempt], attempt+1)
				time.Sleep(backoffs[attempt])
				continue
			}
		}

		// Non-lock error or max retries reached
		return err
	}

	return lastErr
}

// InsertMovie inserts a movie with its cast and media info in a transaction
func InsertMovie(db *sql.DB, movie *models.Movie) (int64, error) {
	var movieID int64

	err := retryOnLock(func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
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
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		movieID = id

		// Insert cast
		for _, cast := range movie.Cast {
			_, err := tx.Exec(`
				INSERT INTO cast (movie_id, name, role, avatar)
				VALUES (?, ?, ?, ?)
			`, movieID, cast.Name, cast.Role, cast.Avatar)
			if err != nil {
				return err
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
					return err
				}
			}

			for _, at := range movie.MediaInfo.AudioTracks {
				_, err := tx.Exec(`
					INSERT INTO audio_tracks (movie_id, codec, channels, language, sample_rate, bitrate)
					VALUES (?, ?, ?, ?, ?, ?)
				`, movieID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
				if err != nil {
					return err
				}
			}

			for _, st := range movie.MediaInfo.SubtitleTracks {
				_, err := tx.Exec(`
					INSERT INTO subtitle_tracks (movie_id, language, format)
					VALUES (?, ?, ?)
				`, movieID, st.Language, st.Format)
				if err != nil {
					return err
				}
			}
		}

		return tx.Commit()
	})

	return movieID, err
}

// InsertEpisode inserts an episode with its media info
func InsertEpisode(db *sql.DB, episode *models.Episode) (int64, error) {
	var episodeID int64

	err := retryOnLock(func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		result, err := tx.Exec(`
			INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added, last_scanned)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, episode.SeriesID, episode.SeasonNum, episode.EpisodeNum, episode.Title, episode.Duration,
			episode.Status, episode.FileSize, episode.FilePath, episode.DateAdded, time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		episodeID = id

		// Insert media info
		if episode.MediaInfo != nil {
			for _, vt := range episode.MediaInfo.VideoTracks {
				_, err := tx.Exec(`
					INSERT INTO video_tracks (episode_id, codec, resolution, fps, bitrate, hdr, color_space)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`, episodeID, vt.Codec, vt.Resolution, vt.FPS, vt.Bitrate, vt.HDR, vt.ColorSpace)
				if err != nil {
					return err
				}
			}

			for _, at := range episode.MediaInfo.AudioTracks {
				_, err := tx.Exec(`
					INSERT INTO audio_tracks (episode_id, codec, channels, language, sample_rate, bitrate)
					VALUES (?, ?, ?, ?, ?, ?)
				`, episodeID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
				if err != nil {
					return err
				}
			}

			for _, st := range episode.MediaInfo.SubtitleTracks {
				_, err := tx.Exec(`
					INSERT INTO subtitle_tracks (episode_id, language, format)
					VALUES (?, ?, ?)
				`, episodeID, st.Language, st.Format)
				if err != nil {
					return err
				}
			}
		}

		return tx.Commit()
	})

	return episodeID, err
}

// InsertSeries inserts a series (without episodes)
func InsertSeries(db *sql.DB, series *models.Series) (int64, error) {
	var seriesID int64

	err := retryOnLock(func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		result, err := tx.Exec(`
			INSERT INTO series (title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id, poster)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, series.Title, series.YearStart, series.YearEnd, series.SeasonCount, series.EpisodeCount,
			series.Synopsis, series.Genres, series.Rating, series.Popularity, series.Status,
			series.FileSize, series.DateAdded, series.TVDBId, series.IMDbId, series.Poster)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		seriesID = id

		// Insert cast
		for _, cast := range series.Cast {
			_, err := tx.Exec(`
				INSERT INTO cast (series_id, name, role, avatar)
				VALUES (?, ?, ?, ?)
			`, seriesID, cast.Name, cast.Role, cast.Avatar)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})

	return seriesID, err
}

// GetOrCreateSeason returns existing season or creates a new one (race-condition safe)
func GetOrCreateSeason(db *sql.DB, seriesID int64, seasonNum int) (int64, error) {
	var seasonID int64

	err := retryOnLock(func() error {
		// First, try to get existing season without transaction
		var id int64
		scanErr := db.QueryRow(`SELECT id FROM seasons WHERE series_id = ? AND number = ?`, seriesID, seasonNum).Scan(&id)
		if scanErr == nil {
			seasonID = id
			return nil
		}
		if scanErr != sql.ErrNoRows {
			return scanErr
		}

		// Season doesn't exist, use transaction to ensure we don't create duplicates
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Check again inside transaction (another goroutine might have created it)
		var existingID int64
		txErr := tx.QueryRow(`SELECT id FROM seasons WHERE series_id = ? AND number = ?`, seriesID, seasonNum).Scan(&existingID)
		if txErr == nil {
			seasonID = existingID
			return tx.Commit()
		}
		if txErr != sql.ErrNoRows {
			return txErr
		}

		// Create new season
		result, err := tx.Exec(`INSERT INTO seasons (series_id, number, file_size) VALUES (?, ?, 0)`, seriesID, seasonNum)
		if err != nil {
			return err
		}

		insertID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		seasonID = insertID

		return tx.Commit()
	})

	return seasonID, err
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
	return retryOnLock(func() error {
		_, err := db.Exec(`
			UPDATE episodes
			SET title = ?, duration = ?, status = ?, file_size = ?, file_path = ?, last_scanned = ?
			WHERE id = ?
		`, episode.Title, episode.Duration, episode.Status, episode.FileSize, episode.FilePath, time.Now().Format(time.RFC3339), episode.ID)
		return err
	})
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
	return retryOnLock(func() error {
		_, err := db.Exec(`
			UPDATE series SET
				season_count = (SELECT COUNT(DISTINCT season_num) FROM episodes WHERE series_id = ?),
				episode_count = (SELECT COUNT(*) FROM episodes WHERE series_id = ?)
			WHERE id = ?
		`, seriesID, seriesID, seriesID)
		return err
	})
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
	return retryOnLock(func() error {
		_, err := db.Exec(`
			UPDATE scan_status SET status = ?, started_at = ?, completed_at = ?, files_found = ?, files_processed = ?, error_message = ?
			WHERE id = 1
		`, status.Status, nullString(status.StartedAt), nullString(status.CompletedAt), status.FilesFound, status.FilesProcessed, nullString(status.ErrorMessage))
		return err
	})
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
