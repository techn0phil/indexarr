# Schema Migrations Guide

This project uses **golang-migrate** for managing database schema changes in a versioned, reversible way. This guide explains how to create, test, and apply migrations.

---

## Quick Reference

### Current Migration Status

**From `backend/go/` directory:**

```bash
cd backend/go
sqlite3 indexarr.db "SELECT * FROM schema_migrations;"
```

**From project root (absolute path):**

```bash
sqlite3 backend/go/indexarr.db "SELECT * FROM schema_migrations;"
```

**Expected output:**
```
version | dirty
--------|-------
      1 |     0
      2 |     0
```

If the command returns nothing, ensure you're running from `backend/go/` or using the full path `backend/go/indexarr.db`.

---

## Creating a New Migration

### Step 1: Generate migration files

Migration files must follow the naming convention: `YYYYMMDDHHMMSS_description.{up,down}.sql`

**Example**: To add a new column to movies table:

```bash
cd backend/go/internal/repository/migrations

# Create empty migration files
touch 20260512120000_add_release_date_to_movies.up.sql
touch 20260512120000_add_release_date_to_movies.down.sql
```

Use the current timestamp for the version number. Use kebab-case for the description.

### Step 2: Write the forward migration (`.up.sql`)

This file contains the SQL that applies the change:

```sql
-- Add release_date column to movies table
ALTER TABLE movies ADD COLUMN release_date TEXT;

-- Create index for efficient querying
CREATE INDEX idx_movies_release_date ON movies(release_date);
```

**Best Practices**:
- Start with a comment explaining the change
- One logical change per migration (e.g., add one column, not multiple unrelated changes)
- Add indexes if the new column will be queried frequently
- Include appropriate constraints (NOT NULL, DEFAULT, etc.)

### Step 3: Write the rollback migration (`.down.sql`)

This file reverses the `.up.sql` changes:

```sql
-- Remove release_date column from movies table
DROP INDEX IF EXISTS idx_movies_release_date;

-- SQLite has limited ALTER TABLE, so we must recreate the table
CREATE TABLE movies_new (
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

INSERT INTO movies_new SELECT id, title, year, duration, synopsis, genres, rating, popularity, status, file_size, file_path, container, date_added, last_scanned, tmdb_id, imdb_id, poster FROM movies;
DROP TABLE movies;
ALTER TABLE movies_new RENAME TO movies;
```

**Important**: Make sure the rollback migration recreates all columns EXCEPT the one being removed. Copy the current schema from `migrations/000001_initial_schema.up.sql` if needed.

---

## Applying Migrations

### Automatic (Application Startup)

Migrations run automatically when the application starts:

```bash
cd backend/go
go run ./cmd/server/main.go
```

The application calls `InitDB()` which:
1. Opens the SQLite database
2. Runs `golang-migrate` to apply any pending migrations
3. Creates `schema_migrations` table to track versions
4. Returns the database connection

**This is the recommended approach for production deployments.**

### Manual (CLI Tool)

Install the `migrate` CLI:

```bash
go install -tags 'sqlite3' github.com/golang-migrate/migrate/cmd/migrate@latest
```

Check migration status (run from `backend/go/`):

```bash
cd backend/go
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" version
```

**Or from project root:**

```bash
migrate -path backend/go/internal/repository/migrations -database "sqlite3://backend/go/indexarr.db" version
```

Run pending migrations:

```bash
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" up
```

Rollback the last migration:

```bash
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" down 1
```

Rollback to a specific version (e.g., v1):

```bash
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" goto 1
```

> **Note**: All `migrate` commands above assume you're in the `backend/go/` directory. Adjust paths accordingly if running from elsewhere.

---

## Testing Migrations

### Before Committing

1. **Test the forward migration** (from `backend/go/`):
   ```bash
   rm -f indexarr.db
   go run ./cmd/server/main.go
   # Wait for server to start (ctrl+c to stop)
   sqlite3 indexarr.db ".schema" | grep -i "your_change"
   ```

2. **Test the rollback**:
   ```bash
   migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" down
   sqlite3 indexarr.db ".schema" | grep "your_change" && echo "ERROR: Column not removed" || echo "OK: Rollback successful"
   ```

3. **Test re-applying**:
   ```bash
   migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" up
   sqlite3 indexarr.db ".schema" | grep -i "your_change"
   ```

### Verification Queries

**From `backend/go/` directory:**

Check current schema version:

```bash
sqlite3 indexarr.db "SELECT * FROM schema_migrations ORDER BY version;"
```

List all tables:

```bash
sqlite3 indexarr.db ".tables"
```

Inspect a table structure:

```bash
sqlite3 indexarr.db "PRAGMA table_info(series);"
```

**From project root:**

Replace `indexarr.db` with `backend/go/indexarr.db` in all commands above.

---

## Migration Workflow (Team Collaboration)

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/add-release-dates
   ```

2. **Create migration files** using the timestamp + kebab-case convention:
   ```bash
   touch migrations/20260512120000_add_release_date_to_movies.{up,down}.sql
   ```

3. **Write the SQL** in both files (forward and rollback)

4. **Test locally** (see "Testing Migrations" section)

5. **Commit and push**:
   ```bash
   git add backend/go/internal/repository/migrations/
   git commit -m "Add release_date column to movies table"
   git push
   ```

6. **On production deployment**:
   - Application startup automatically runs pending migrations
   - No manual intervention needed
   - All instances can be updated simultaneously

### Preventing Conflicts

**Use ISO 8601 timestamps** (YYYYMMDDHHMMSS) to ensure unique migration versions:
- Don't reuse version numbers
- If multiple developers create migrations, timestamps prevent conflicts
- Example: `20260512120000`, `20260512123045`, etc.

---

## Common Patterns

### Adding a Column

```sql
-- up
ALTER TABLE movies ADD COLUMN budget INTEGER DEFAULT 0;

-- down (see "Writing Rollback Migrations" above for full table recreation)
```

### Adding a Unique Index

```sql
-- up
CREATE UNIQUE INDEX idx_movies_imdb_id ON movies(imdb_id);

-- down
DROP INDEX idx_movies_imdb_id;
```

### Modifying Data (with Schema Change)

```sql
-- up
ALTER TABLE movies ADD COLUMN status_v2 TEXT DEFAULT 'available';
UPDATE movies SET status_v2 = status;
-- (optional: drop old column if needed, see table recreation pattern)

-- down
-- Reverse the data transformation
```

### Creating a New Table

```sql
-- up
CREATE TABLE ratings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  movie_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  score REAL NOT NULL,
  created_at TEXT,
  FOREIGN KEY(movie_id) REFERENCES movies(id)
);

-- down
DROP TABLE IF EXISTS ratings;
```

---

## SQLite Limitations

SQLite has limited `ALTER TABLE` support. For complex changes (renaming columns, removing columns, etc.), use the "recreate table" pattern:

```sql
-- CREATE new table with desired structure
CREATE TABLE movies_new ( ... );

-- Copy data from old table
INSERT INTO movies_new SELECT ... FROM movies;

-- Drop old table
DROP TABLE movies;

-- Rename new table
ALTER TABLE movies_new RENAME TO movies;

-- Recreate indexes
CREATE INDEX ... ON movies(...);
```

All migration examples in this project follow this pattern for safe rollbacks.

---

## Troubleshooting

### Migrations fail to apply

**Error**: `failed to lookup series: no such column: poster`

**Cause**: The application code tries to query a column that hasn't been migrated yet.

**Solution**: Ensure migrations run before seeding/querying. In `db.go`, `InitDB()` calls `runMigrations()` before returning the database connection.

### Database is locked

**Error**: `database is locked`

**Cause**: Multiple processes accessing SQLite simultaneously.

**Solution**: 
- Stop all instances of the application
- Let the current migration finish
- Restart the application

### Migration version mismatch

**Error**: `error: pgx connect failed: connect refused` or similar

**Solution**: 
```bash
# Check current version
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" version

# Reset to a known state if needed
rm -f indexarr.db  # WARNING: deletes data
go run ./cmd/server/main.go  # Re-apply all migrations from scratch
```

### Command returns nothing / blank output

**Error**: `sqlite3 indexarr.db "SELECT * FROM schema_migrations;"` returns no results or is blank

**Cause**: The command is being run from the wrong directory. The database file is at `backend/go/indexarr.db`, not in the current working directory.

**Solution**:
- **Option 1** (Recommended): Change to the correct directory
  ```bash
  cd backend/go
  sqlite3 indexarr.db "SELECT * FROM schema_migrations;"
  ```

- **Option 2**: Use the full path from anywhere
  ```bash
  sqlite3 backend/go/indexarr.db "SELECT * FROM schema_migrations;"
  ```

- **Option 3**: Use the migrate CLI (directory-independent)
  ```bash
  cd backend/go
  migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" version
  ```

---

## References

- **golang-migrate docs**: https://github.com/golang-migrate/migrate
- **SQLite ALTER TABLE**: https://www.sqlite.org/lang_altertable.html
- **Migrations in Go**: https://pkg.go.dev/github.com/golang-migrate/migrate/v4

---

**Last Updated**: 2026-05-11  
**Current Schema Version**: 2 (series table with poster column)
