package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"indexarr/internal/models"
)

// buildOrClause creates an OR condition from comma-separated values
// Example: "3840,1920" returns "(resolution LIKE '%3840%' OR resolution LIKE '%1920%')"
func buildOrClause(fieldName, filterValue string) string {
	if filterValue == "" {
		return ""
	}

	values := strings.Split(filterValue, ",")
	if len(values) == 0 {
		return ""
	}

	var conditions []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", fieldName, v))
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return "(" + strings.Join(conditions, " OR ") + ")"
}

func GetMovies(db *sql.DB, filters *models.FilterCriteria) ([]models.Movie, int64, error) {
	// Default pagination
	if filters.PageSize <= 0 {
		filters.PageSize = 50
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	offset := (filters.Page - 1) * filters.PageSize

	// Build WHERE clause
	where := "1=1"
	if filters.Status != "" {
		where += fmt.Sprintf(" AND status='%s'", filters.Status)
	}
	if filters.Search != "" {
		where += fmt.Sprintf(" AND title LIKE '%%%s%%'", filters.Search)
	}

	// Filters requiring joins to related tables - support comma-separated values with OR logic
	if filters.Resolution != "" {
		resolutionClause := buildOrClause("resolution", filters.Resolution)
		if resolutionClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM video_tracks WHERE movie_id = movies.id AND %s)", resolutionClause)
		}
	}
	if filters.Codec != "" {
		codecClause := buildOrClause("codec", filters.Codec)
		if codecClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM video_tracks WHERE movie_id = movies.id AND %s)", codecClause)
		}
	}
	if filters.Audio != "" {
		audioClause := buildOrClause("codec", filters.Audio)
		if audioClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM audio_tracks WHERE movie_id = movies.id AND %s)", audioClause)
		}
	}
	if filters.HDR != "" {
		hdrClause := buildOrClause("hdr", filters.HDR)
		if hdrClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM video_tracks WHERE movie_id = movies.id AND %s)", hdrClause)
		}
	}

	// Count total
	var total int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM movies WHERE %s", where)).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query movies
	query := fmt.Sprintf(`SELECT id, title, year, duration, synopsis, genres, rating, popularity, status, file_size, file_path, container, date_added, tmdb_id, imdb_id, poster FROM movies WHERE %s ORDER BY title LIMIT ? OFFSET ?`, where)
	rows, err := db.Query(query, filters.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		var poster sql.NullString
		err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Duration, &m.Synopsis, &m.Genres, &m.Rating, &m.Popularity, &m.Status, &m.FileSize, &m.FilePath, &m.Container, &m.DateAdded, &m.TMDBId, &m.IMDbId, &poster)
		if poster.Valid {
			m.Poster = &poster.String
		} else {
			m.Poster = nil
		}
		if err != nil {
			return nil, 0, err
		}
		// Load related data
		m.Cast, _ = GetCastForMovie(db, m.ID)
		m.MediaInfo, _ = GetMediaInfoForMovie(db, m.ID)
		movies = append(movies, m)
	}

	return movies, total, nil
}

func GetMovieByID(db *sql.DB, id int64) (*models.Movie, error) {
	var m models.Movie
	var poster sql.NullString
	err := db.QueryRow(`SELECT id, title, year, duration, synopsis, genres, rating, popularity, status, file_size, file_path, container, date_added, tmdb_id, imdb_id, poster FROM movies WHERE id=?`, id).Scan(&m.ID, &m.Title, &m.Year, &m.Duration, &m.Synopsis, &m.Genres, &m.Rating, &m.Popularity, &m.Status, &m.FileSize, &m.FilePath, &m.Container, &m.DateAdded, &m.TMDBId, &m.IMDbId, &poster)
	if poster.Valid {
		m.Poster = &poster.String
	} else {
		m.Poster = nil
	}
	if err != nil {
		return nil, err
	}
	m.Cast, _ = GetCastForMovie(db, m.ID)
	m.MediaInfo, _ = GetMediaInfoForMovie(db, m.ID)
	return &m, nil
}

func GetCastForMovie(db *sql.DB, movieID int64) ([]models.Cast, error) {
	rows, err := db.Query("SELECT id, name, role, avatar FROM cast WHERE movie_id=? ORDER BY id", movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cast []models.Cast
	for rows.Next() {
		var c models.Cast
		err := rows.Scan(&c.ID, &c.Name, &c.Role, &c.Avatar)
		if err != nil {
			return nil, err
		}
		cast = append(cast, c)
	}
	return cast, nil
}

func GetMediaInfoForMovie(db *sql.DB, movieID int64) (*models.MediaInfo, error) {
	// Get video tracks
	videoRows, err := db.Query("SELECT id, codec, resolution, fps, bitrate, hdr, color_space FROM video_tracks WHERE movie_id=?", movieID)
	if err != nil {
		return nil, err
	}
	defer videoRows.Close()

	mi := &models.MediaInfo{}
	for videoRows.Next() {
		var vt models.VideoTrack
		var id int64
		err := videoRows.Scan(&id, &vt.Codec, &vt.Resolution, &vt.FPS, &vt.Bitrate, &vt.HDR, &vt.ColorSpace)
		if err != nil {
			return nil, err
		}
		mi.VideoTracks = append(mi.VideoTracks, vt)
	}

	// Get audio tracks
	audioRows, err := db.Query("SELECT id, codec, channels, language, sample_rate, bitrate FROM audio_tracks WHERE movie_id=?", movieID)
	if err != nil {
		return nil, err
	}
	defer audioRows.Close()

	for audioRows.Next() {
		var at models.AudioTrack
		var id int64
		err := audioRows.Scan(&id, &at.Codec, &at.Channels, &at.Language, &at.SampleRate, &at.Bitrate)
		if err != nil {
			return nil, err
		}
		mi.AudioTracks = append(mi.AudioTracks, at)
	}

	// Get subtitle tracks
	subRows, err := db.Query("SELECT id, language, format FROM subtitle_tracks WHERE movie_id=?", movieID)
	if err != nil {
		return nil, err
	}
	defer subRows.Close()

	for subRows.Next() {
		var st models.SubtitleTrack
		var id int64
		err := subRows.Scan(&id, &st.Language, &st.Format)
		if err != nil {
			return nil, err
		}
		mi.SubtitleTracks = append(mi.SubtitleTracks, st)
	}

	return mi, nil
}

func GetSeries(db *sql.DB, filters *models.FilterCriteria) ([]models.Series, int64, error) {
	if filters.PageSize <= 0 {
		filters.PageSize = 50
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	offset := (filters.Page - 1) * filters.PageSize

	where := "1=1"
	if filters.Status != "" {
		where += fmt.Sprintf(" AND status='%s'", filters.Status)
	}
	if filters.Search != "" {
		where += fmt.Sprintf(" AND title LIKE '%%%s%%'", filters.Search)
	}

	// Filters requiring joins to related tables (episodes -> video/audio tracks) - support comma-separated values with OR logic
	if filters.Resolution != "" {
		resolutionClause := buildOrClause("vt.resolution", filters.Resolution)
		if resolutionClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM episodes e JOIN video_tracks vt ON e.id = vt.episode_id WHERE e.series_id = series.id AND %s)", resolutionClause)
		}
	}
	if filters.Codec != "" {
		codecClause := buildOrClause("vt.codec", filters.Codec)
		if codecClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM episodes e JOIN video_tracks vt ON e.id = vt.episode_id WHERE e.series_id = series.id AND %s)", codecClause)
		}
	}
	if filters.Audio != "" {
		audioClause := buildOrClause("at.codec", filters.Audio)
		if audioClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM episodes e JOIN audio_tracks at ON e.id = at.episode_id WHERE e.series_id = series.id AND %s)", audioClause)
		}
	}
	if filters.HDR != "" {
		hdrClause := buildOrClause("vt.hdr", filters.HDR)
		if hdrClause != "" {
			where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM episodes e JOIN video_tracks vt ON e.id = vt.episode_id WHERE e.series_id = series.id AND %s)", hdrClause)
		}
	}

	var total int64
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM series WHERE %s", where)).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id FROM series WHERE %s ORDER BY title LIMIT ? OFFSET ?`, where)
	rows, err := db.Query(query, filters.PageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var series []models.Series
	for rows.Next() {
		var s models.Series
		err := rows.Scan(&s.ID, &s.Title, &s.YearStart, &s.YearEnd, &s.SeasonCount, &s.EpisodeCount, &s.Synopsis, &s.Genres, &s.Rating, &s.Popularity, &s.Status, &s.FileSize, &s.DateAdded, &s.TVDBId, &s.IMDbId)
		if err != nil {
			return nil, 0, err
		}
		s.Cast, _ = GetCastForSeries(db, s.ID)
		series = append(series, s)
	}

	return series, total, nil
}

func GetSeriesByID(db *sql.DB, id int64) (*models.Series, error) {
	var s models.Series
	err := db.QueryRow(`SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id FROM series WHERE id=?`, id).Scan(&s.ID, &s.Title, &s.YearStart, &s.YearEnd, &s.SeasonCount, &s.EpisodeCount, &s.Synopsis, &s.Genres, &s.Rating, &s.Popularity, &s.Status, &s.FileSize, &s.DateAdded, &s.TVDBId, &s.IMDbId)
	if err != nil {
		return nil, err
	}
	s.Cast, _ = GetCastForSeries(db, s.ID)
	s.Seasons, _ = GetSeasonsForSeries(db, s.ID)
	return &s, nil
}

func GetCastForSeries(db *sql.DB, seriesID int64) ([]models.Cast, error) {
	rows, err := db.Query("SELECT id, name, role, avatar FROM cast WHERE series_id=? ORDER BY id", seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cast []models.Cast
	for rows.Next() {
		var c models.Cast
		err := rows.Scan(&c.ID, &c.Name, &c.Role, &c.Avatar)
		if err != nil {
			return nil, err
		}
		cast = append(cast, c)
	}
	return cast, nil
}

func GetSeasonsForSeries(db *sql.DB, seriesID int64) ([]models.Season, error) {
	rows, err := db.Query("SELECT id, series_id, number, file_size FROM seasons WHERE series_id=? ORDER BY number", seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasons []models.Season
	for rows.Next() {
		var s models.Season
		err := rows.Scan(&s.ID, &s.SeriesID, &s.Number, &s.FileSize)
		if err != nil {
			return nil, err
		}
		s.Episodes, _ = GetEpisodesForSeason(db, seriesID, s.Number)
		// Calculate available/missing
		for _, ep := range s.Episodes {
			if ep.Status == "available" {
				s.AvailableEps++
			} else if ep.Status == "missing" {
				s.MissingEps++
			}
		}
		seasons = append(seasons, s)
	}

	return seasons, nil
}

func GetEpisodesForSeason(db *sql.DB, seriesID int64, seasonNum int) ([]models.Episode, error) {
	rows, err := db.Query(`SELECT id, series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added FROM episodes WHERE series_id=? AND season_num=? ORDER BY episode_num`, seriesID, seasonNum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var episodes []models.Episode
	for rows.Next() {
		var e models.Episode
		err := rows.Scan(&e.ID, &e.SeriesID, &e.SeasonNum, &e.EpisodeNum, &e.Title, &e.Duration, &e.Status, &e.FileSize, &e.FilePath, &e.DateAdded)
		if err != nil {
			return nil, err
		}
		if e.Status == "available" {
			e.MediaInfo, _ = GetMediaInfoForEpisode(db, e.ID)
		}
		episodes = append(episodes, e)
	}

	return episodes, nil
}

func GetMediaInfoForEpisode(db *sql.DB, episodeID int64) (*models.MediaInfo, error) {
	mi := &models.MediaInfo{}

	videoRows, err := db.Query("SELECT id, codec, resolution, fps, bitrate, hdr, color_space FROM video_tracks WHERE episode_id=?", episodeID)
	if err != nil {
		return nil, err
	}
	defer videoRows.Close()

	for videoRows.Next() {
		var vt models.VideoTrack
		var id int64
		err := videoRows.Scan(&id, &vt.Codec, &vt.Resolution, &vt.FPS, &vt.Bitrate, &vt.HDR, &vt.ColorSpace)
		if err != nil {
			return nil, err
		}
		mi.VideoTracks = append(mi.VideoTracks, vt)
	}

	audioRows, err := db.Query("SELECT id, codec, channels, language, sample_rate, bitrate FROM audio_tracks WHERE episode_id=?", episodeID)
	if err != nil {
		return nil, err
	}
	defer audioRows.Close()

	for audioRows.Next() {
		var at models.AudioTrack
		var id int64
		err := audioRows.Scan(&id, &at.Codec, &at.Channels, &at.Language, &at.SampleRate, &at.Bitrate)
		if err != nil {
			return nil, err
		}
		mi.AudioTracks = append(mi.AudioTracks, at)
	}

	return mi, nil
}

func GetStats(db *sql.DB) (*models.StatsResponse, error) {
	stats := &models.StatsResponse{Success: true}

	// Total movies
	db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&stats.TotalMovies)

	// Total series
	db.QueryRow("SELECT COUNT(*) FROM series").Scan(&stats.TotalSeries)

	// Total episodes
	db.QueryRow("SELECT COUNT(*) FROM episodes").Scan(&stats.TotalEpisodes)

	// Available/missing counts
	db.QueryRow("SELECT COUNT(*) FROM movies WHERE status='available'").Scan(&stats.AvailMovies)
	db.QueryRow("SELECT COUNT(*) FROM movies WHERE status='missing'").Scan(&stats.MissingMovies)
	db.QueryRow("SELECT COUNT(*) FROM episodes WHERE status='available'").Scan(&stats.AvailEpisodes)
	db.QueryRow("SELECT COUNT(*) FROM episodes WHERE status='missing'").Scan(&stats.MissingEpisodes)

	// Problems count (missing files)
	stats.ProblemsCount = stats.MissingMovies + stats.MissingEpisodes

	// Disk space in GB
	var totalBytes int64
	db.QueryRow("SELECT COALESCE(SUM(file_size), 0) FROM movies WHERE status='available' UNION ALL SELECT COALESCE(SUM(file_size), 0) FROM episodes WHERE status='available'").Scan(&totalBytes)
	stats.DiskSpaceGB = float64(totalBytes) / (1024 * 1024 * 1024)

	// 4K count
	query := `
		SELECT COUNT(DISTINCT CASE WHEN vt.resolution LIKE '3840%' THEN m.id END) FROM movies m
		LEFT JOIN video_tracks vt ON m.id = vt.movie_id
		WHERE m.status='available'
	`
	db.QueryRow(query).Scan(&stats.FourKCount)
	if stats.TotalMovies > 0 {
		stats.FourKPercent = float64(stats.FourKCount) / float64(stats.TotalMovies) * 100
	}

	return stats, nil
}
