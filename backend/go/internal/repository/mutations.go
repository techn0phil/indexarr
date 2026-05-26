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

// UpdateMovie updates an existing movie and its related data
func UpdateMovie(db *sql.DB, movie *models.Movie) error {
	return retryOnLock(func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Update movie
		_, err = tx.Exec(`
			UPDATE movies
			SET title = ?, year = ?, duration = ?, synopsis = ?, genres = ?, rating = ?, popularity = ?, status = ?, file_size = ?, file_path = ?, container = ?, last_scanned = ?, tmdb_id = ?, imdb_id = ?, poster = ?
			WHERE id = ?
		`, movie.Title, movie.Year, movie.Duration, movie.Synopsis, movie.Genres, movie.Rating, movie.Popularity,
			movie.Status, movie.FileSize, movie.FilePath, movie.Container, time.Now().Format(time.RFC3339),
			movie.TMDBId, movie.IMDbId, movie.Poster, movie.ID)
		if err != nil {
			return err
		}

		// Delete existing cast and media info (simpler than diffing)
		_, err = tx.Exec(`DELETE FROM cast WHERE movie_id = ?`, movie.ID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`DELETE FROM video_tracks WHERE movie_id = ?`, movie.ID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`DELETE FROM audio_tracks WHERE movie_id = ?`, movie.ID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`DELETE FROM subtitle_tracks WHERE movie_id = ?`, movie.ID)
		if err != nil {
			return err
		}

		// Re-insert cast
		for _, cast := range movie.Cast {
			_, err := tx.Exec(`
				INSERT INTO cast (movie_id, name, role, avatar)
				VALUES (?, ?, ?, ?)
			`, movie.ID, cast.Name, cast.Role, cast.Avatar)
			if err != nil {
				return err
			}
		}

		// Re-insert media info
		if movie.MediaInfo != nil {
			for _, vt := range movie.MediaInfo.VideoTracks {
				_, err := tx.Exec(`
					INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space)
					VALUES (?, ?, ?, ?, ?, ?, ?)
				`, movie.ID, vt.Codec, vt.Resolution, vt.FPS, vt.Bitrate, vt.HDR, vt.ColorSpace)
				if err != nil {
					return err
				}
			}

			for _, at := range movie.MediaInfo.AudioTracks {
				_, err := tx.Exec(`
					INSERT INTO audio_tracks (movie_id, codec, channels, language, sample_rate, bitrate)
					VALUES (?, ?, ?, ?, ?, ?)
				`, movie.ID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
				if err != nil {
					return err
				}
			}

			for _, st := range movie.MediaInfo.SubtitleTracks {
				_, err := tx.Exec(`
					INSERT INTO subtitle_tracks (movie_id, language, format)
					VALUES (?, ?, ?)
				`, movie.ID, st.Language, st.Format)
				if err != nil {
					return err
				}
			}
		}

		return tx.Commit()
	})
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
			INSERT INTO series (title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tmdb_id, tvdb_id, imdb_id, poster, slug, sonarr_id, title_slug)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, series.Title, series.YearStart, series.YearEnd, series.SeasonCount, series.EpisodeCount,
			series.Synopsis, series.Genres, series.Rating, series.Popularity, series.Status,
			series.FileSize, series.DateAdded, series.TMDBId, series.TVDBId, series.IMDbId, series.Poster, series.Slug, series.SonarrID, series.TitleSlug)
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

// GetSeriesByTitleAndYear finds a series by title and year (case-insensitive)
func GetSeriesByTitleAndYear(db *sql.DB, title string, year int) (*models.Series, error) {
	var series models.Series
	var poster sql.NullString
	err := db.QueryRow(`
		SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tmdb_id, tvdb_id, imdb_id, poster, slug
		FROM series WHERE LOWER(title) = LOWER(?) AND year_start = ?
	`, title, year).Scan(&series.ID, &series.Title, &series.YearStart, &series.YearEnd, &series.SeasonCount, &series.EpisodeCount,
		&series.Synopsis, &series.Genres, &series.Rating, &series.Popularity, &series.Status,
		&series.FileSize, &series.DateAdded, &series.TMDBId, &series.TVDBId, &series.IMDbId, &poster, &series.Slug)
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

// GetSeriesByTMDBId finds a series by TMDB ID
func GetSeriesByTMDBId(db *sql.DB, tmdbID int64) (*models.Series, error) {
	var series models.Series
	var poster sql.NullString
	var sonarrID sql.NullInt64
	var titleSlug sql.NullString
	err := db.QueryRow(`
		SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tmdb_id, tvdb_id, imdb_id, poster, slug, sonarr_id, title_slug
		FROM series WHERE tmdb_id = ?
	`, tmdbID).Scan(&series.ID, &series.Title, &series.YearStart, &series.YearEnd, &series.SeasonCount, &series.EpisodeCount,
		&series.Synopsis, &series.Genres, &series.Rating, &series.Popularity, &series.Status,
		&series.FileSize, &series.DateAdded, &series.TMDBId, &series.TVDBId, &series.IMDbId, &poster, &series.Slug, &sonarrID, &titleSlug)
	if poster.Valid {
		series.Poster = &poster.String
	} else {
		series.Poster = nil
	}
	if sonarrID.Valid {
		series.SonarrID = sonarrID.Int64
	}
	if titleSlug.Valid {
		series.TitleSlug = titleSlug.String
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &series, nil
}

// GetSeriesBySonarrID finds a series by Sonarr ID
func GetSeriesBySonarrID(db *sql.DB, sonarrID int64) (*models.Series, error) {
	var series models.Series
	var poster sql.NullString
	var sonarrIDVal sql.NullInt64
	var titleSlug sql.NullString
	err := db.QueryRow(`
		SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tmdb_id, tvdb_id, imdb_id, poster, slug, sonarr_id, title_slug
		FROM series WHERE sonarr_id = ?
	`, sonarrID).Scan(&series.ID, &series.Title, &series.YearStart, &series.YearEnd, &series.SeasonCount, &series.EpisodeCount,
		&series.Synopsis, &series.Genres, &series.Rating, &series.Popularity, &series.Status,
		&series.FileSize, &series.DateAdded, &series.TMDBId, &series.TVDBId, &series.IMDbId, &poster, &series.Slug, &sonarrIDVal, &titleSlug)
	if poster.Valid {
		series.Poster = &poster.String
	} else {
		series.Poster = nil
	}
	if sonarrIDVal.Valid {
		series.SonarrID = sonarrIDVal.Int64
	}
	if titleSlug.Valid {
		series.TitleSlug = titleSlug.String
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

// DeleteMovie deletes a movie by ID
func DeleteMovie(db *sql.DB, movieID int64) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM movies WHERE id = ?`, movieID)
		return err
	})
}

// DeleteMovieByPath deletes movies matching a file path pattern (for handling moved/deleted files)
func DeleteMovieByPath(db *sql.DB, pathPattern string) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM movies WHERE file_path LIKE ?`, pathPattern)
		return err
	})
}

// DeleteEpisode deletes an episode by ID (cascade constraints handle tracks)
func DeleteEpisode(db *sql.DB, episodeID int64) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM episodes WHERE id = ?`, episodeID)
		return err
	})
}

// DeleteEpisodeByPath deletes episodes matching a file path pattern (for handling moved/deleted files)
func DeleteEpisodeByPath(db *sql.DB, pathPattern string) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM episodes WHERE file_path LIKE ?`, pathPattern)
		return err
	})
}

// DeleteSeries deletes a series by ID (cascade constraints handle episodes, seasons, cast, and tracks)
func DeleteSeries(db *sql.DB, seriesID int64) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM series WHERE id = ?`, seriesID)
		return err
	})
}

func DeleteEmptySeasons(db *sql.DB, seriesID int64) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`
			DELETE FROM seasons
			WHERE series_id = ? AND number NOT IN (SELECT DISTINCT season_num FROM episodes WHERE series_id = ?)
		`, seriesID, seriesID)
		return err
	})
}

// DeleteEmptySeries deletes series that have no episodes (used after scanning to clean up)
func DeleteEmptySeries(db *sql.DB) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`
			DELETE FROM series
			WHERE id NOT IN (SELECT DISTINCT series_id FROM episodes)
		`)
		return err
	})
}

// UpdateSeries updates an existing series and its cast
func UpdateSeries(db *sql.DB, series *models.Series) error {
	return retryOnLock(func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Update series
		_, err = tx.Exec(`
			UPDATE series
			SET title = ?, year_start = ?, year_end = ?, synopsis = ?, genres = ?, rating = ?, popularity = ?, status = ?, file_size = ?, tmdb_id = ?, tvdb_id = ?, imdb_id = ?, poster = ?, slug = ?, sonarr_id = ?, title_slug = ?
			WHERE id = ?
		`, series.Title, series.YearStart, series.YearEnd, series.Synopsis, series.Genres, series.Rating, series.Popularity,
			series.Status, series.FileSize, series.TMDBId, series.TVDBId, series.IMDbId, series.Poster, series.Slug, series.SonarrID, series.TitleSlug, series.ID)
		if err != nil {
			return err
		}

		// Delete existing cast and re-insert (simpler than diffing)
		_, err = tx.Exec(`DELETE FROM cast WHERE series_id = ?`, series.ID)
		if err != nil {
			return err
		}

		// Re-insert cast
		for _, cast := range series.Cast {
			_, err := tx.Exec(`
				INSERT INTO cast (series_id, name, role, avatar)
				VALUES (?, ?, ?, ?)
			`, series.ID, cast.Name, cast.Role, cast.Avatar)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}

// GetTVDBToken retrieves the stored TVDB bearer token (if exists and not expired)
func GetTVDBToken(db *sql.DB) (string, time.Time, error) {
	var token string
	var expiresAt string

	err := db.QueryRow(`SELECT token, expires_at FROM tvdb_tokens WHERE id = 1`).Scan(&token, &expiresAt)
	if err == sql.ErrNoRows {
		return "", time.Time{}, nil // No token stored
	}
	if err != nil {
		return "", time.Time{}, err
	}

	// Parse expiry time
	expiry, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiry, nil
}

// SaveTVDBToken stores the TVDB bearer token with expiry time
func SaveTVDBToken(db *sql.DB, token string, expiresAt time.Time) error {
	return retryOnLock(func() error {
		// Use INSERT OR REPLACE to ensure only one token exists
		_, err := db.Exec(`
			INSERT OR REPLACE INTO tvdb_tokens (id, token, expires_at, created_at)
			VALUES (1, ?, ?, ?)
		`, token, expiresAt.Format(time.RFC3339), time.Now().Format(time.RFC3339))
		return err
	})
}

// GetMovieByTMDBId finds a movie by TMDB ID
func GetMovieByTMDBId(db *sql.DB, tmdbID int64) (*models.Movie, error) {
	var m models.Movie
	var poster sql.NullString
	err := db.QueryRow(`
		SELECT id, title, year, duration, synopsis, genres, rating, popularity, status, file_size, file_path, container, date_added, tmdb_id, imdb_id, poster
		FROM movies WHERE tmdb_id = ?
	`, tmdbID).Scan(&m.ID, &m.Title, &m.Year, &m.Duration, &m.Synopsis, &m.Genres, &m.Rating, &m.Popularity, &m.Status, &m.FileSize, &m.FilePath, &m.Container, &m.DateAdded, &m.TMDBId, &m.IMDbId, &poster)
	if poster.Valid {
		m.Poster = &poster.String
	} else {
		m.Poster = nil
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.Cast, _ = GetCastForMovie(db, m.ID)
	m.MediaInfo, _ = GetMediaInfoForMovie(db, m.ID)
	return &m, nil
}

// GetAllMovieTMDBIds returns all TMDB IDs of movies in the database
func GetAllMovieTMDBIds(db *sql.DB) ([]int64, error) {
	rows, err := db.Query(`SELECT tmdb_id FROM movies WHERE tmdb_id > 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetAllMovieFilePaths returns all file paths of movies in the database
func GetAllMovieFilePaths(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT file_path FROM movies WHERE file_path IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// DeleteMovieByTMDBId deletes a movie by its TMDB ID
func DeleteMovieByTMDBId(db *sql.DB, tmdbID int64) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM movies WHERE tmdb_id = ?`, tmdbID)
		return err
	})
}

// GetAllSeriesSonarrIDs returns all Sonarr IDs of series in the database
func GetAllSeriesSonarrIDs(db *sql.DB) ([]int64, error) {
	rows, err := db.Query(`SELECT sonarr_id FROM series WHERE sonarr_id > 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetAllEpisodeFilePaths returns all file paths of episodes in the database
func GetAllEpisodeFilePaths(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT file_path FROM episodes WHERE file_path IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// DeleteSeriesBySonarrID deletes a series by its Sonarr ID
func DeleteSeriesBySonarrID(db *sql.DB, sonarrID int64) error {
	return retryOnLock(func() error {
		_, err := db.Exec(`DELETE FROM series WHERE sonarr_id = ?`, sonarrID)
		return err
	})
}
