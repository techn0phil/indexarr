# Indexarr Backend (Go)

This is the backend for Indexarr, a media library management application. It provides a RESTful API for managing movies, tv-shows, episodes, and technical metadata.

## Features
- RESTful API for movies, tv-shows, episodes
- Multi-criteria filtering and search
- Media file scanning and metadata extraction
- Statistics service (totals, disk usage, 4K %)
- Integration with TMDB/TVDB for metadata
- Unit and integration tests

## Project Structure
```
backend/go/
├── cmd/server/           # Entry point (main.go)
├── internal/
│   ├── api/              # HTTP handlers and routes
│   ├── config/           # Configuration
│   ├── models/           # Data models (Movie, Series, Episode, MediaInfo)
│   ├── repository/       # Database layer
│   └── services/         # Business logic (catalog, scanner, stats)
├── go.mod                # Go module
├── go.sum                # Dependencies
└── README.md             # This file
```

## Getting Started
1. Install dependencies:
   ```bash
   go mod download && go mod tidy
   ```
2. Run the server:
   ```bash
   go run ./cmd/server
   ```
3. Run tests:
   ```bash
   go test ./...
   ```

## API Endpoints
- `GET /api/movies` — List movies (with filters)
- `GET /api/movies/:id` — Movie details
- `GET /api/series` — List series
- `GET /api/series/:id` — Series details
- `GET /api/series/:id/seasons/:season/episodes` — List episodes
- `POST /api/scan` — Trigger media scan
- `GET /api/stats` — Library statistics

## Conventions
- Packages: lowercase, no underscores
- Types: PascalCase
- Functions: camelCase
- Constants: UPPER_CASE_WITH_UNDERSCORES

## License
GPL v3 — see root LICENSE
