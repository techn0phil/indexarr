-- Movies table
CREATE TABLE IF NOT EXISTS movies (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  year INTEGER,
  duration INTEGER,
  synopsis TEXT,
  genres TEXT,
  rating REAL,
  popularity REAL,
  status TEXT DEFAULT 'available',
  file_size INTEGER,
  file_path TEXT,
  container TEXT,
  date_added TEXT,
  last_scanned TEXT,
  tmdb_id INTEGER,
  imdb_id TEXT,
  poster TEXT
);

-- Series table
CREATE TABLE IF NOT EXISTS series (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  year_start INTEGER,
  year_end INTEGER,
  season_count INTEGER,
  episode_count INTEGER,
  synopsis TEXT,
  genres TEXT,
  rating REAL,
  popularity REAL,
  status TEXT DEFAULT 'complete',
  file_size INTEGER,
  date_added TEXT,
  tvdb_id INTEGER UNIQUE,
  imdb_id TEXT,
  poster TEXT
);

-- Seasons table
CREATE TABLE IF NOT EXISTS seasons (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  series_id INTEGER NOT NULL,
  number INTEGER NOT NULL,
  file_size INTEGER,
  FOREIGN KEY(series_id) REFERENCES series(id),
  UNIQUE(series_id, number)
);

-- Episodes table
CREATE TABLE IF NOT EXISTS episodes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  series_id INTEGER NOT NULL,
  season_num INTEGER NOT NULL,
  episode_num INTEGER NOT NULL,
  title TEXT,
  duration INTEGER,
  status TEXT DEFAULT 'available',
  file_size INTEGER,
  file_path TEXT,
  date_added TEXT,
  last_scanned TEXT,
  FOREIGN KEY(series_id) REFERENCES series(id),
  UNIQUE(series_id, season_num, episode_num)
);

-- Scan status table (tracks scan progress)
CREATE TABLE IF NOT EXISTS scan_status (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  status TEXT DEFAULT 'idle',
  started_at TEXT,
  completed_at TEXT,
  files_found INTEGER DEFAULT 0,
  files_processed INTEGER DEFAULT 0,
  error_message TEXT
);

-- Video tracks table
CREATE TABLE IF NOT EXISTS video_tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  episode_id INTEGER,
  codec TEXT,
  resolution TEXT,
  fps REAL,
  bitrate TEXT,
  hdr TEXT,
  color_space TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id),
  FOREIGN KEY(episode_id) REFERENCES episodes(id)
);

-- Audio tracks table
CREATE TABLE IF NOT EXISTS audio_tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  episode_id INTEGER,
  codec TEXT,
  channels TEXT,
  language TEXT,
  sample_rate TEXT,
  bitrate TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id),
  FOREIGN KEY(episode_id) REFERENCES episodes(id)
);

-- Subtitle tracks table
CREATE TABLE IF NOT EXISTS subtitle_tracks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  episode_id INTEGER,
  language TEXT,
  format TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id),
  FOREIGN KEY(episode_id) REFERENCES episodes(id)
);

-- Cast table
CREATE TABLE IF NOT EXISTS cast (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER,
  series_id INTEGER,
  name TEXT,
  role TEXT,
  avatar TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id),
  FOREIGN KEY(series_id) REFERENCES series(id)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_movies_status ON movies(status);
CREATE INDEX IF NOT EXISTS idx_movies_title ON movies(title);
CREATE INDEX IF NOT EXISTS idx_series_status ON series(status);
CREATE INDEX IF NOT EXISTS idx_series_title ON series(title);
CREATE INDEX IF NOT EXISTS idx_episodes_series ON episodes(series_id);
CREATE INDEX IF NOT EXISTS idx_video_tracks_movie ON video_tracks(movie_id);
CREATE INDEX IF NOT EXISTS idx_video_tracks_episode ON video_tracks(episode_id);
CREATE INDEX IF NOT EXISTS idx_audio_tracks_movie ON audio_tracks(movie_id);
CREATE INDEX IF NOT EXISTS idx_audio_tracks_episode ON audio_tracks(episode_id);
