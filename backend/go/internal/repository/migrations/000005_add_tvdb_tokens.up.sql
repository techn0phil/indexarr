-- Add tvdb_tokens table for storing TVDB API bearer tokens
CREATE TABLE IF NOT EXISTS tvdb_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tvdb_tokens_singleton ON tvdb_tokens (id);
