-- Add DELETE CASCADE constraints to ensure media info and cast data are automatically removed when a movie is deleted
-- Do not drop and recreate tables to preserve existing data; instead, use ALTER TABLE to add constraints

PRAGMA foreign_keys = OFF;

-- video_tracks
ALTER TABLE video_tracks RENAME TO video_tracks_old;

CREATE TABLE video_tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  episode_id INTEGER,
  codec TEXT,
  resolution TEXT,
  fps REAL,
  bitrate TEXT,
  hdr TEXT,
  color_space TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id) ON DELETE CASCADE,
  FOREIGN KEY(episode_id) REFERENCES episodes(id) ON DELETE CASCADE
);

INSERT INTO video_tracks (
  id, movie_id, episode_id, codec, resolution, fps, bitrate, hdr, color_space
)
SELECT
  id, movie_id, episode_id, codec, resolution, fps, bitrate, hdr, color_space
FROM video_tracks_old;

DROP TABLE video_tracks_old;

CREATE INDEX IF NOT EXISTS idx_video_tracks_movie ON video_tracks(movie_id);
CREATE INDEX IF NOT EXISTS idx_video_tracks_episode ON video_tracks(episode_id);


-- audio_tracks
ALTER TABLE audio_tracks RENAME TO audio_tracks_old;

CREATE TABLE audio_tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  episode_id INTEGER,
  codec TEXT,
  channels TEXT,
  language TEXT,
  sample_rate TEXT,
  bitrate TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id) ON DELETE CASCADE,
  FOREIGN KEY(episode_id) REFERENCES episodes(id) ON DELETE CASCADE
);

INSERT INTO audio_tracks (
  id, movie_id, episode_id, codec, channels, language, sample_rate, bitrate
)
SELECT
  id, movie_id, episode_id, codec, channels, language, sample_rate, bitrate
FROM audio_tracks_old;

DROP TABLE audio_tracks_old;

CREATE INDEX IF NOT EXISTS idx_audio_tracks_movie ON audio_tracks(movie_id);
CREATE INDEX IF NOT EXISTS idx_audio_tracks_episode ON audio_tracks(episode_id);


-- subtitle_tracks
ALTER TABLE subtitle_tracks RENAME TO subtitle_tracks_old;

CREATE TABLE subtitle_tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  episode_id INTEGER,
  language TEXT,
  format TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id) ON DELETE CASCADE,
  FOREIGN KEY(episode_id) REFERENCES episodes(id) ON DELETE CASCADE
);

INSERT INTO subtitle_tracks (
  id, movie_id, episode_id, language, format
)
SELECT
  id, movie_id, episode_id, language, format
FROM subtitle_tracks_old;

DROP TABLE subtitle_tracks_old;

CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_movie ON subtitle_tracks(movie_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_tracks_episode ON subtitle_tracks(episode_id);


-- cast
ALTER TABLE "cast" RENAME TO cast_old;

CREATE TABLE "cast" (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  series_id INTEGER,
  name TEXT,
  role TEXT,
  avatar TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id) ON DELETE CASCADE,
  FOREIGN KEY(series_id) REFERENCES series(id) ON DELETE CASCADE
);

INSERT INTO "cast" (
  id, movie_id, series_id, name, role, avatar
)
SELECT
  id, movie_id, series_id, name, role, avatar
FROM cast_old;

DROP TABLE cast_old;


PRAGMA foreign_keys = ON;
