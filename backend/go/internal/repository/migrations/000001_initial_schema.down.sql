-- Drop all indexes
DROP INDEX IF EXISTS idx_movies_status;
DROP INDEX IF EXISTS idx_movies_title;
DROP INDEX IF EXISTS idx_series_status;
DROP INDEX IF EXISTS idx_series_title;
DROP INDEX IF EXISTS idx_episodes_series;
DROP INDEX IF EXISTS idx_video_tracks_movie;
DROP INDEX IF EXISTS idx_video_tracks_episode;
DROP INDEX IF EXISTS idx_audio_tracks_movie;
DROP INDEX IF EXISTS idx_audio_tracks_episode;

-- Drop all tables
DROP TABLE IF EXISTS cast;
DROP TABLE IF EXISTS subtitle_tracks;
DROP TABLE IF EXISTS audio_tracks;
DROP TABLE IF EXISTS video_tracks;
DROP TABLE IF EXISTS scan_status;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS seasons;
DROP TABLE IF EXISTS series;
DROP TABLE IF EXISTS movies;
