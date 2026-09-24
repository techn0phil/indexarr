# Schema Migrations Guide

This project uses **golang-migrate** for managing database schema changes in a versioned, reversible way. This guide explains how to create, test, and apply migrations.

---

## Quick Reference

### Current Migration Status

**From `backend/go/` directory:**

```bash
cd backend/go
sqlite3 indexarr.db "SELECT * FROM schema_migrations ORDER BY version;"
```

**From project root (absolute path):**

```bash
sqlite3 backend/go/indexarr.db "SELECT * FROM schema_migrations ORDER BY version;"
```

**Expected output (current):**
```
version | dirty
--------|-------
      1 |     0
      2 |     0
      3 |     0
      4 |     0
      5 |     0
      6 |     0
      7 |     0
      8 |     0
      9 |     0
     10 |     0
     11 |     0
```

If the command returns nothing, ensure you're running from `backend/go/` or using the full path `backend/go/indexarr.db`.

---

## Creating a New Migration

### Step 1: Generate migration files

Migration files must follow the naming convention: `NNNNN_description.{up,down}.sql` where `NNNNN` is the next sequential version number (6 digits, padded with zeros).

**Example**: To add a new column to movies table (migration 12):

```bash
cd backend/go/internal/repository/migrations

# Create empty migration files
touch 000012_add_release_date_to_movies.up.sql
touch 000012_add_release_date_to_movies.down.sql
```

**Important**: Always use the next sequential version number (currently starting at 000012). Use kebab-case for the description.

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
- Avoid using `IF NOT EXISTS` clauses — migrations should be deterministic and idempotent within their version

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

### Automatic (Application Startup) — RECOMMENDED

Migrations run automatically when the application starts using embedded migrations:

```bash
cd backend/go
go run ./cmd/server/main.go
```

The application calls `InitDB()` which:
1. Opens the SQLite database
2. Embeds and runs pending migrations from `internal/repository/migrations/` via `go:embed`
3. Applies migrations using `golang-migrate` (embedded in the binary)
4. Creates `schema_migrations` table to track versions
5. Configures SQLite pragmas (WAL mode, busy timeout, cache size, etc.)
6. Returns the database connection

**This is the recommended and only approach for production deployments.** Migrations are embedded in the binary, so no external files are required.

### Manual Testing with CLI Tool (Optional)

To manually test migrations during development, install the `migrate` CLI:

```bash
go install -tags 'sqlite3' github.com/golang-migrate/migrate/cmd/migrate@latest
```

Check migration status (run from `backend/go/`):

```bash
cd backend/go
~/go/bin/migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" version
```

**Or from project root:**

```bash
~/go/bin/migrate -path backend/go/internal/repository/migrations -database "sqlite3://backend/go/indexarr.db" version
```

Run pending migrations:

```bash
~/go/bin/migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" up
```

Rollback the last migration:

```bash
~/go/bin/migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" down 1
```

Rollback to a specific version (e.g., v1):

```bash
~/go/bin/migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" goto 1
```

> **Note**: 
> - All `migrate` commands above assume you're in the `backend/go/` directory
> - The CLI tool is optional — migrations always run automatically on app startup
> - Adjust paths accordingly if running from elsewhere

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

2. **Find the next version number**:
   ```bash
   cd backend/go/internal/repository/migrations
   ls -1 *.up.sql | tail -1  # Shows current highest version
   ```

3. **Create migration files** using the next sequential version + kebab-case:
   ```bash
   touch 000012_add_release_date_to_movies.{up,down}.sql
   ```

4. **Write the SQL** in both files (forward and rollback)

5. **Test locally** (see "Testing Migrations" section)

6. **Commit and push**:
   ```bash
   git add backend/go/internal/repository/migrations/
   git commit -m "Add release_date column to movies table"
   git push
   ```

7. **On production deployment**:
   - Rebuild the application binary (`docker build` or `go build`)
   - Application startup automatically runs pending migrations
   - No manual intervention needed
   - All instances can be updated simultaneously

### Preventing Conflicts

**Use sequential version numbers** (000001, 000002, etc.) with 6-digit padding:
- Always use the next available sequential number
- Don't reuse version numbers
- Check `ls -1 *.up.sql | tail -1` to find the latest version
- Coordinate with teammates to avoid merge conflicts in version ordering

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

## Embedded Migrations Architecture

This project uses **embedded migrations** via Go's `go:embed` feature. Migration files are compiled into the binary, eliminating external file dependencies.

### How It Works

1. **Compile-time embedding** (`db.go`):
   ```go
   //go:embed migrations/*.sql
   var migrationsFS embed.FS
   ```

2. **Runtime execution**:
   - Application calls `InitDB(dbPath)` on startup
   - `runMigrations()` uses `iofs.New()` to access embedded filesystem
   - `golang-migrate` applies any pending migrations
   - Database continues operating normally

### Benefits

- ✅ Zero external dependencies at runtime
- ✅ Atomic deployment (binary includes schema changes)
- ✅ No risk of schema drift (migrations always deployed together with code)
- ✅ Simpler Docker deployments (no volume mounts for migration files)

### SQLite Configuration

On database initialization, the following pragmas are set:

```sql
PRAGMA journal_mode=WAL;              -- Write-Ahead Logging for concurrent reads
PRAGMA busy_timeout=5000;             -- 5-second retry on lock contention
PRAGMA cache_size=-64000;             -- 64MB in-memory cache
PRAGMA synchronous=NORMAL;            -- Balance between speed and safety
PRAGMA foreign_keys=on;               -- Enable foreign key constraints
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

### Important: Embedded Migrations

Migrations are **embedded in the binary** at build time. This means:
- Migration files in `internal/repository/migrations/` are baked into the executable
- Changes to migration files require a **rebuild** (`go build` or `docker build`)
- The `migrate` CLI tool operates on the filesystem, not the embedded filesystem
- In production, migrations run automatically on app startup — no manual CLI invocation

If you've edited a migration file and the changes aren't applied, rebuild the binary and restart the application.

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

## Recent Migrations Summary

- **Migration 9**: User authentication table with username/password/role (admin/guest)
- **Migration 10**: Episode and season total count tracking
- **Migration 11**: i18n support with language code standardization (en, fr, es, de, it, ja, ko, zh)

---

**Last Updated**: 2026-08-24  
**Current Schema Version**: 11 (i18n support with language codes)
