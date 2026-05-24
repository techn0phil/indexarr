---
name: db-migration
description: "Create, review, and test SQLite schema migrations for Indexarr using golang-migrate. Handles forward (.up.sql) and rollback (.down.sql) migrations with proper naming, indexing, and constraints. Use when: creating database migrations, adding columns, modifying schema, reviewing migration safety."
mode: tool
tools:
  - read_file
  - create_file
  - grep_search
  - file_search
  - run_in_terminal
---

# Database Migration Agent

Specialized agent for creating and managing SQLite schema migrations in the Indexarr project.

## Workflow

When invoked, this agent will:

1. **Understand the change**: Ask clarifying questions about the desired schema modification
2. **Check current schema**: Review existing migrations to understand current state
3. **Generate migration version**: Create sequential migration number (000XXX)
4. **Write .up.sql**: Create forward migration with appropriate SQL
5. **Write .down.sql**: Create rollback migration (handles SQLite's limited ALTER TABLE)
6. **Validate syntax**: Check SQL syntax and migration pair consistency
7. **Test migrations**: Run up/down migrations on test database
8. **Document**: Update MIGRATIONS.md if needed

## Migration Patterns

### Naming Convention

```
internal/repository/migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000007_add_sonarr_fields.up.sql
└── 000007_add_sonarr_fields.down.sql
```

**Format**: `0000XX_description.{up,down}.sql`
- Use sequential numbering (check last migration number)
- Use snake_case for description
- Keep descriptions concise but meaningful

### Common Migration Types

#### 1. Add Column

**Forward (up.sql)**:
```sql
-- Add column_name to table_name
ALTER TABLE table_name ADD COLUMN column_name TYPE DEFAULT_VALUE;

-- Create index if column will be queried
CREATE INDEX idx_table_column ON table_name(column_name);
```

**Rollback (down.sql)**:
```sql
-- Remove column_name from table_name (SQLite limitation)
-- Must recreate table without the column

-- 1. Create new table without the column
CREATE TABLE table_name_new (
  id INTEGER PRIMARY KEY,
  existing_col1 TYPE,
  existing_col2 TYPE
  -- (omit the removed column)
);

-- 2. Copy data from old table (excluding new column)
INSERT INTO table_name_new SELECT id, existing_col1, existing_col2 FROM table_name;

-- 3. Drop old table
DROP TABLE table_name;

-- 4. Rename new table
ALTER TABLE table_name_new RENAME TO table_name;

-- 5. Drop the index
DROP INDEX IF EXISTS idx_table_column;
```

#### 2. Add Foreign Key or Constraint

**Forward (up.sql)**:
```sql
-- Add foreign key constraint (requires table recreation in SQLite)
PRAGMA foreign_keys=off;

BEGIN TRANSACTION;

CREATE TABLE table_name_new (
  id INTEGER PRIMARY KEY,
  existing_col TYPE,
  new_fk_col INTEGER,
  FOREIGN KEY(new_fk_col) REFERENCES other_table(id) ON DELETE CASCADE
);

INSERT INTO table_name_new SELECT * FROM table_name;
DROP TABLE table_name;
ALTER TABLE table_name_new RENAME TO table_name;

COMMIT;

PRAGMA foreign_keys=on;
```

**Rollback (down.sql)**:
```sql
-- Reverse the constraint addition
-- (Similar table recreation without the constraint)
```

#### 3. Add Index

**Forward (up.sql)**:
```sql
-- Create index for efficient lookups
CREATE INDEX idx_table_column ON table_name(column_name);
```

**Rollback (down.sql)**:
```sql
-- Remove index
DROP INDEX IF EXISTS idx_table_column;
```

#### 4. Create Table

**Forward (up.sql)**:
```sql
-- Create new_table with schema
CREATE TABLE IF NOT EXISTS new_table (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  column1 TYPE NOT NULL,
  column2 TYPE,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(column1) REFERENCES other_table(id)
);

-- Create indexes
CREATE INDEX idx_new_table_column1 ON new_table(column1);
```

**Rollback (down.sql)**:
```sql
-- Drop table and indexes
DROP INDEX IF EXISTS idx_new_table_column1;
DROP TABLE IF EXISTS new_table;
```

## SQLite Migration Gotchas

### Limited ALTER TABLE Support

SQLite has restricted ALTER TABLE capabilities:
- ✅ **Can**: `ADD COLUMN`, `RENAME COLUMN`, `RENAME TABLE`
- ❌ **Cannot**: `DROP COLUMN`, `ADD CONSTRAINT`, `MODIFY COLUMN`

**Solution**: For unsupported operations, recreate the table:
1. Create new table with desired schema
2. Copy data from old table
3. Drop old table
4. Rename new table

### Foreign Key Constraints

- Must disable foreign keys during table recreation: `PRAGMA foreign_keys=off`
- Re-enable after: `PRAGMA foreign_keys=on`
- Wrap in transaction to maintain consistency

### Index Creation

- Always check if index exists: `CREATE INDEX IF NOT EXISTS`
- Drop with safety check: `DROP INDEX IF EXISTS`
- Index names must be globally unique (not per-table)

### Default Values

- Default values only apply to new rows after migration
- Existing rows keep NULL (or previous default)
- Use UPDATE statement if existing rows need default value

## Testing Migrations

### Step 1: Check Migration Number

```bash
cd backend/go/internal/repository/migrations
ls -1 *.up.sql | tail -1
# Output: 000007_add_sonarr_fields.up.sql
# Next migration: 000008_your_description.up.sql
```

### Step 2: Apply Forward Migration

```bash
cd backend/go
go run ./cmd/server/main.go
# Migrations auto-apply on startup
```

**Or manually**:
```bash
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" up 1
```

### Step 3: Verify Schema

```bash
sqlite3 indexarr.db ".schema table_name"
```

### Step 4: Test Rollback

```bash
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" down 1
```

### Step 5: Verify Rollback

```bash
sqlite3 indexarr.db ".schema table_name"
# Should match schema before migration
```

### Step 6: Reapply Forward

```bash
migrate -path internal/repository/migrations -database "sqlite3://indexarr.db" up 1
```

## Pre-Migration Checklist

Before creating a migration, verify:

- [ ] Current schema state reviewed (`internal/repository/schema.sql` or latest migration)
- [ ] Migration number is sequential (one higher than last migration)
- [ ] Description is clear and follows snake_case convention
- [ ] Forward migration (.up.sql) contains complete SQL
- [ ] Rollback migration (.down.sql) reverses all changes
- [ ] Indexes added for foreign keys and frequently queried columns
- [ ] Comments explain the purpose of each statement
- [ ] SQLite limitations considered (table recreation for DROP COLUMN, etc.)
- [ ] Foreign key constraints handled properly (PRAGMA foreign_keys)
- [ ] Transaction boundaries set where needed (BEGIN/COMMIT)

## Post-Migration Checklist

After creating migration files:

- [ ] Migration tested on development database
- [ ] Rollback tested successfully
- [ ] Re-apply forward migration tested
- [ ] No data loss during up/down cycle
- [ ] Related Go models updated if needed (`internal/models/`)
- [ ] Repository queries updated if columns/tables changed
- [ ] MIGRATIONS.md updated with notes if complex migration

## Example: Complete Migration

**Scenario**: Add `radarr_id` column to movies table for Radarr integration.

**000008_add_radarr_id_to_movies.up.sql**:
```sql
-- Add Radarr integration field to movies table
ALTER TABLE movies ADD COLUMN radarr_id INTEGER;

-- Create index for Radarr ID lookups (will be queried frequently)
CREATE INDEX idx_movies_radarr_id ON movies(radarr_id);
```

**000008_add_radarr_id_to_movies.down.sql**:
```sql
-- Remove Radarr integration field from movies table

-- Drop index first
DROP INDEX IF EXISTS idx_movies_radarr_id;

-- SQLite limitation: must recreate table without radarr_id column
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
  -- radarr_id OMITTED
);

-- Copy data (excluding radarr_id)
INSERT INTO movies_new 
SELECT id, title, year, duration, synopsis, genres, rating, popularity, 
       status, file_size, file_path, container, date_added, last_scanned, 
       tmdb_id, imdb_id, poster 
FROM movies;

-- Replace old table
DROP TABLE movies;
ALTER TABLE movies_new RENAME TO movies;
```

## Safety Guidelines

1. **Always test migrations on development database first**
2. **Never modify existing migration files** (create new migration instead)
3. **Backup production database before running migrations**
4. **Test rollback immediately after applying forward migration**
5. **Use transactions for multi-statement migrations**
6. **Document complex migrations in MIGRATIONS.md**
7. **Update Go models to match schema changes**

## Related Files

- **Migration Guide**: [backend/go/MIGRATIONS.md](../../backend/go/MIGRATIONS.md)
- **Schema Definition**: [backend/go/internal/repository/schema.sql](../../backend/go/internal/repository/schema.sql)
- **Database Init**: [backend/go/internal/repository/db.go](../../backend/go/internal/repository/db.go)
- **Data Models**: [backend/go/internal/models/](../../backend/go/internal/models/)

## When to Create New Migration

Create a new migration when:
- Adding/removing columns from tables
- Adding/removing indexes
- Modifying constraints (foreign keys, unique, check)
- Creating/dropping tables
- Changing column types or defaults
- Adding/removing triggers (if used)

Do NOT create migration for:
- Data updates (use seed scripts or API endpoints)
- Temporary schema experiments (use separate test database)
- Changes to non-schema config (use environment variables)
