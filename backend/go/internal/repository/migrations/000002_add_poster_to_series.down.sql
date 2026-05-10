-- Remove poster column from series table (rollback)
-- Note: SQLite has limited ALTER TABLE support; this approach preserves data
CREATE TABLE series_new (
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
  imdb_id TEXT
);

INSERT INTO series_new SELECT id, title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id FROM series;
DROP TABLE series;
ALTER TABLE series_new RENAME TO series;
