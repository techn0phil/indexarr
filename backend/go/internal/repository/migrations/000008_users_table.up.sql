-- Users table for local authentication
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'guest' CHECK (role IN ('admin', 'guest')),
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for username lookups during login
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Index for listing enabled users
CREATE INDEX IF NOT EXISTS idx_users_enabled ON users(enabled);
