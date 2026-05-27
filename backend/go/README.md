
# Indexarr Backend (Go)

This is the backend for Indexarr, a media library management application. It provides a RESTful API for managing movies, TV series, episodes, and technical metadata.

## Features
- RESTful API for movies, TV series, episodes
- Multi-criteria filtering and search
- Media file scanning and metadata extraction
- Statistics service (totals, disk usage, 4K %)
- Integration with TMDB/TVDB for metadata
- Database migrations with golang-migrate
- Environment-based configuration (`.env`)
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
│   │   ├── handlers.go           # HTTP handlers
│   │   ├── routes.go             # Route definitions
│   │   ├── scan_handlers.go      # Scan API handlers
│   │   └── websocket.go          # WebSocket for scan progress
│   ├── config/
│   │   └── config.go             # Loads env/config
│   ├── models/
│   │   ├── movie.go
│   │   ├── series.go
│   │   ├── episode.go
│   │   ├── mediainfo.go
│   │   ├── scan.go
│   │   └── filter.go
│   ├── repository/
│   │   ├── db.go                 # DB connection
│   │   ├── queries.go            # Query helpers
│   │   ├── mutations.go          # Write helpers
│   │   ├── exists.go             # Existence checks
│   │   ├── schema.sql            # Schema reference
│   │   ├── seed.go               # Seed data
│   │   └── migrations/
│   │       ├── 000001_initial_schema.up.sql
│   │       ├── ...               # Other migration scripts
│   └── services/
│       ├── scanner.go            # File/metadata scanner
│       ├── filesystem_scanner.go # Filesystem import logic
│       ├── radarr_client.go      # Radarr API client
│       ├── radarr_importer.go    # Radarr import logic
│       ├── sonarr_client.go      # Sonarr API client
│       ├── sonarr_importer.go    # Sonarr import logic
│       ├── tmdb.go               # TMDB metadata
│       ├── tvdb.go               # TVDB metadata
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

Key variables:
- `TMDB_API_KEY`, `TVDB_API_KEY` — Required for metadata
- `RADARR_URL`, `RADARR_API_KEY` — For Radarr integration (optional)
- `SONARR_URL`, `SONARR_API_KEY` — For Sonarr integration (optional)
- `MOVIES_LIBRARY_PATHS`, `SERIES_LIBRARY_PATHS` — For filesystem scanning
- `SCAN_INTERVAL` — Scan schedule (hours)

See `.env.example` for all options.

## Database & Migrations

- SQLite database: `indexarr.db` (created in `backend/go/`)
- Schema migrations managed with [golang-migrate](https://github.com/golang-migrate/migrate)
- Migration scripts in `internal/repository/migrations/`
- See [MIGRATIONS.md](MIGRATIONS.md) for migration workflow

## API Endpoints (Main)
- `GET /api/movies` — List/filter movies
- `GET /api/movies/:id` — Movie details
- `GET /api/series` — List/filter series
- `GET /api/series/:id` — Series details
- `GET /api/series/:id/seasons/:season/episodes` — List episodes
- `POST /api/scan` — Trigger media scan
- `GET /api/scan/status` — Scan status
- `POST /api/scan/stop` — Stop scan
- `GET /api/stats` — Library statistics
- `GET /api/config` — Current config
- `POST /api/purge` — Purge all data

## Conventions
- Packages: lowercase, no underscores
- Types: PascalCase
- Functions: camelCase
- Constants: UPPER_CASE_WITH_UNDERSCORES

## License
GPL v3 — see root LICENSE
