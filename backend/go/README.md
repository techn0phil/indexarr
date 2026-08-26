
# Indexarr Backend (Go)

This is the backend for Indexarr, a media library management application. It provides a RESTful API for managing movies, TV series, episodes, and technical metadata.

## Features
- RESTful API for movies, TV series, episodes
- Multi-criteria filtering and search
- Media file scanning and metadata extraction
- Statistics service (totals, disk usage, 4K %)
- Integration with TMDB/TVDB for metadata
- Authentication service (JWT-based, supports none/simple/oidc modes)
- User management (create, update, delete, password change)
- Database migrations with golang-migrate
- Environment-based configuration (`.env`)
- Real-time WebSocket updates for scan progress
- Unit and integration tests

## Project Structure
```
backend/go/
├── .env.example                  # Example environment config
├── .env                          # Local environment config (not committed)
├── cmd/
│   └── server/
│       └── main.go               # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go           # Media HTTP handlers
│   │   ├── auth_handlers.go      # Authentication HTTP handlers
│   │   ├── routes.go             # Route definitions
│   │   ├── middleware.go         # Auth middleware
│   │   └── websocket.go          # WebSocket for scan progress
│   ├── config/
│   │   └── config.go             # Loads env/config
│   ├── models/
│   │   ├── movie.go
│   │   ├── series.go
│   │   ├── episode.go
│   │   ├── mediainfo.go
│   │   ├── scan.go
│   │   ├── filter.go
│   │   └── user.go
│   ├── repository/
│   │   ├── db.go                 # DB connection
│   │   ├── queries.go            # Query helpers
│   │   ├── mutations.go          # Write helpers
│   │   ├── exists.go             # Existence checks
│   │   ├── user_repository.go    # User CRUD operations
│   │   ├── schema.sql            # Schema reference
│   │   ├── seed.go               # Seed data
│   │   └── migrations/
│   │       ├── 000001_initial_schema.up.sql
│   │       ├── ...               # Migration scripts
│   │       └── 000009_create_users.up.sql
│   └── services/
│       ├── auth.go               # Authentication service (JWT)
│       ├── scanner.go            # File/metadata scanner orchestration
│       ├── filesystem_scanner.go # Filesystem import logic
│       ├── radarr_client.go      # Radarr API client
│       ├── radarr_importer.go    # Radarr import logic
│       ├── sonarr_client.go      # Sonarr API client
│       ├── sonarr_importer.go    # Sonarr import logic
│       ├── importer.go           # Media import orchestration
│       ├── tmdb.go               # TMDB metadata client
│       ├── tvdb.go               # TVDB metadata client
│       ├── extractor.go          # MediaInfo extraction
│       ├── parser.go             # Parsing helpers
│       ├── broadcaster.go        # WebSocket broadcast
│       └── scheduler.go          # Scan scheduler
├── go.mod                        # Go module
├── go.sum                        # Dependencies
├── MIGRATIONS.md                 # Migration workflow guide
└── README.md                     # This file
```

## Getting Started
1. Copy `.env.example` to `.env` and set required environment variables (see below).
2. Install dependencies:
   ```bash
   go mod download && go mod tidy
   ```
3. Run the server (from `backend/go/`):
   ```bash
   go run ./cmd/server
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

## Configuration

All configuration is via environment variables (see `.env.example`).

### Import Modes

Indexarr supports a **flexible dual import architecture** — mix and match modes per media type:

**Mode 1: Radarr/Sonarr Integration**
- Calls Radarr/Sonarr API to fetch existing library
- Extracts mediainfo from files Radarr/Sonarr knows about
- Supports path mapping for Docker mounts (e.g., `/downloads:/mnt/media`)
- Automatically removes items deleted from Radarr/Sonarr

**Mode 2: Filesystem Scanner**
- Walks directory trees looking for video files (mkv, mp4, avi, etc.)
- Extracts mediainfo directly from discovered files
- Enriches metadata via TMDB/TVDB APIs
- No dependency on external services

**Mix and Match Examples**:
- Movies via Radarr + Series via filesystem scanner
- Movies via filesystem + Series via Sonarr
- Enable both modes simultaneously

### Key Environment Variables

- **Metadata APIs** (required):
  - `TMDB_API_KEY` — The Movie Database API key
  - `TVDB_API_KEY` — TV Database API key

- **Radarr Integration** (optional):
  - `RADARR_URL` — Radarr instance URL
  - `RADARR_API_KEY` — Radarr API key
  - `RADARR_PATH_MAPPING` — Path translation (e.g., `/downloads:/mnt/media`)

- **Sonarr Integration** (optional):
  - `SONARR_URL` — Sonarr instance URL
  - `SONARR_API_KEY` — Sonarr API key
  - `SONARR_PATH_MAPPING` — Path translation (e.g., `/downloads:/mnt/media`)

- **Filesystem Scanning** (optional):
  - `MOVIES_LIBRARY_PATHS` — Comma-separated movie directory paths
  - `SERIES_LIBRARY_PATHS` — Comma-separated series directory paths

- **Scheduling & Detection**:
  - `SCAN_INTERVAL` — Hours between automatic scans (0=disabled)
  - `DETECTION_LANGUAGE` — For media detection during scanning
  - `METADATA_LANGUAGE` — For metadata fetching from APIs

See `.env.example` for all options.

## Database & Migrations

- SQLite database: `indexarr.db` (created in `backend/go/`)
- Schema migrations managed with [golang-migrate](https://github.com/golang-migrate/migrate)
- Migration scripts in `internal/repository/migrations/`
- See [MIGRATIONS.md](MIGRATIONS.md) for migration workflow

## API Endpoints

### Public (no authentication required)
- `GET /api/auth/config` — Get authentication mode configuration
- `POST /api/auth/login` — Login with credentials (if auth enabled)
- `GET /health` — Health check

### Protected (authentication required if enabled)

**Movies & Series**:
- `GET /api/movies` — List/filter movies
- `GET /api/movies/:id` — Movie details with cast and mediainfo
- `POST /api/movies/:id/refresh` — Refresh single movie metadata
- `GET /api/series` — List/filter series
- `GET /api/series/:id` — Series details with seasons and episodes
- `POST /api/series/:id/refresh` — Refresh single series metadata

**Scanning**:
- `POST /api/scan` — Trigger full media scan
- `POST /api/scan/movies` — Trigger movies-only scan
- `POST /api/scan/series` — Trigger series-only scan
- `GET /api/scan/status` — Current scan status
- `POST /api/scan/stop` — Stop running scan
- `GET /api/scan/ws` — WebSocket for real-time scan progress

**Statistics & Configuration**:
- `GET /api/stats` — Library statistics (totals, disk usage, 4K %)
- `GET /api/config` — Current configuration (import mode, library paths)
- `POST /api/purge` — Purge all media data (keeps schema)

**Authentication & User Management**:
- `GET /api/auth/me` — Get current user info
- `POST /api/auth/change-password` — Change current user password
- `POST /api/auth/logout` — Logout (clear auth cookie)
- `GET /api/users` — List all users (admin only, simple auth mode)
- `POST /api/users` — Create new user (admin only, simple auth mode)
- `PUT /api/users/:id` — Update user details (admin only, simple auth mode)
- `DELETE /api/users/:id` — Delete user (admin only, simple auth mode)
- `POST /api/users/:id/password` — Admin set user password (admin only, simple auth mode)

## Authentication

Three authentication modes are supported:
- **`none`** (default): No authentication required
- **`simple`**: Username/password with JWT tokens. Requires `AUTH_ADMIN_USERNAME`, `AUTH_ADMIN_PASSWORD`, and `AUTH_SESSION_SECRET` environment variables.
- **`oidc`**: OpenID Connect (future implementation)

When auth is enabled, tokens are stored in HttpOnly cookies and validated by the `AuthMiddleware` on protected endpoints. See `.env.example` for configuration.

## Key Gotchas & Patterns

### Database & Concurrency
- **WAL Mode**: SQLite uses Write-Ahead Logging (WAL) with 5s busy timeout to minimize locking
- **Connection Pool**: Limited connection pool prevents database contention
- **Long Transactions**: Avoid long-running transactions; they can trigger "database is locked" errors
- **Batch Updates**: Scan status updates are batched rather than per-file to reduce contention
- For detailed analysis, see `/memories/repo/database-locking-analysis.md`

### Import Architecture
- **Nil Service Checks**: Importers can be `nil` if not configured (e.g., no Radarr URL). Always check `if movieImporter != nil` before calling methods.
- **Path Mapping**: Radarr/Sonarr paths may differ from local filesystem in Docker. Use `RADARR_PATH_MAPPING` and `SONARR_PATH_MAPPING` to translate paths.
- **Per-Scan Caching**: TVDB lookups are cached per scan to avoid redundant API calls. Cache clears at scan start.

### File Processing
- **Mediainfo Timeouts**: Extraction has 30s timeout per file. Large files or network mounts may timeout.
- **Memory Management**: Extractor calls `unix.Fadvise(FADV_DONTNEED)` after reading files to clear Linux page cache and prevent memory bloat during large scans.
- **Status Calculation**: Status (available, missing, problem) is derived from file existence checks during scan, not on-demand.

### Filtering
- **OR Logic**: Backend supports comma-separated OR queries: `?resolution=3840,1920` means "3840 OR 1920"
- **Multiple Criteria**: Combine multiple filter types with AND logic: `?resolution=3840&codec=H.265&audio=FLAC`

### WebSocket & Real-time Updates
- **Broadcast**: All connected clients receive scan progress updates; no per-client tracking
- **Public Endpoint**: WebSocket broadcast happens without auth gating
- **Reconnection**: Frontend handles reconnection with exponential backoff

## Testing

Run all backend tests:
```bash
go test ./...
```

Run targeted package suites:
```bash
go test ./internal/repository ./internal/services ./internal/api
```

Run with race detector:
```bash
go test ./... -race
```

Run with verbose output:
```bash
go test -v ./...
```

Generate coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

Testing notes:
- Unit tests must be deterministic and should not call real external services or processes.
- External API behavior should be tested with `httptest` servers and local doubles.
- Repository and service tests should use migration-aware in-memory SQLite setup helpers.

### Test Categories

**Unit Tests**:
- Mediainfo parser: Parse test samples, verify codec/resolution/HDR extraction
- Filter logic: Test filtering by status, resolution, codec combinations
- Stats calculation: Verify aggregations (total size, 4K count, etc.)

**Integration Tests**:
- API endpoints: Test HTTP request/response with mock data
- Database operations: CRUD operations on media records

**Test Files**: Name tests `*_test.go` in the same package as the code being tested.

## Conventions
- Packages: lowercase, no underscores
- Types: PascalCase
- Functions: camelCase
- Constants: UPPER_CASE_WITH_UNDERSCORES

## Troubleshooting

### Database Locked Errors
If you see "database is locked" errors:
1. Check for long-running transactions that hold locks
2. Ensure `SCAN_INTERVAL` isn't too aggressive (default: 24 hours)
3. Stop concurrent scans (only one scan should run at a time)
4. Review [database-locking-analysis.md](/memories/repo/database-locking-analysis.md) for deeper insights

### Mediainfo Not Found
Ensure `mediainfo` CLI is installed on the system:
```bash
# Ubuntu/Debian
sudo apt-get install mediainfo

# macOS
brew install mediainfo

# Docker
# The Dockerfile includes mediainfo installation
```

### API Returns 401 Unauthorized
If authentication is enabled (`AUTH_MODE=simple` or `oidc`):
1. Verify you've provided a valid JWT token in the `auth_token` cookie or `Authorization: Bearer <token>` header
2. Check that token hasn't expired (default max age: 168 hours = 7 days)
3. Ensure `AUTH_SESSION_SECRET` is set consistently across restarts

### Scan Hangs or Times Out
- Mediainfo extraction timeout per file: 30 seconds
- Check for network mounts that are slow or unavailable
- Review server logs for extraction errors on specific files

## License
GPL v3 — see root LICENSE
