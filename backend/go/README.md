
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


## Authentication & User Management

Indexarr backend supports authentication and user management with two modes:

- **local**: Local users stored in the database. Admin user can be bootstrapped via environment variables.
- **oidc**: Login via an external OIDC provider (e.g., Auth0, Google, Keycloak). Local users are disabled in this mode.
- **disabled**: No authentication (open access, not recommended for production).

Set the mode with the `AUTH_MODE` environment variable.

### User management

- **Admin user**: The first admin user can be created via environment variables (`AUTH_ADMIN_USERNAME`, `AUTH_ADMIN_PASSWORD`) or via the API if no users exist.
- **User roles**: Each user has a role (either `admin` or `guest`). Only admins can manage users and perform destructive actions (purge, create/delete users, etc).
- **User management API**: Admins can create, update, enable/disable, or delete users via the `/api/users` endpoints.
- **Password management**: Users can change their own password; admins can reset any user's password.

### API endpoints (auth & users)

**Public (no auth required):**
- `GET /api/auth/config` — Get current authentication mode/config
- `POST /api/auth/login` — Login (returns JWT)
- `POST /api/auth/logout` — Logout (client-side only)
- `GET /api/auth/oidc/login` — Start OIDC login (if enabled)
- `GET /api/auth/oidc/callback` — OIDC callback (if enabled)

**Authenticated:**
- `GET /api/auth/me` — Get current user info
- `POST /api/auth/change-password` — Change own password

**Admin only:**
- `GET /api/users` — List all users
- `POST /api/users` — Create user
- `PUT /api/users/{id}` — Update user (username, role, enabled)
- `DELETE /api/users/{id}` — Delete user
- `POST /api/users/{id}/password` — Set/reset user password

### Environment variables (auth)

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_MODE` | local | Authentication mode: `disabled`, `local`, or `oidc` |
| `AUTH_ADMIN_USERNAME` | - | Bootstrap admin username (only used if no users exist) |
| `AUTH_ADMIN_PASSWORD` | - | Bootstrap admin password (only used if no users exist) |
| `AUTH_SESSION_SECRET` | - | Secret for signing JWT sessions (required for auth) |
| `AUTH_SESSION_MAX_AGE` | 168 | Session maximum duration (7 days by default) |
| `OIDC_ISSUER_URL` | - | OIDC issuer URL (if using OIDC) |
| `OIDC_CLIENT_ID` | - | OIDC client ID (if using OIDC) |
| `OIDC_CLIENT_SECRET` | - | OIDC client secret (if using OIDC) |
| `OIDC_REDIRECT_URL` | - | OIDC redirect/callback URL (if using OIDC) |
| `OIDC_SCOPES` | - | OIDC scopes (if using OIDC) |
| `OIDC_ROLES_CLAIM` | - | OIDC claim containing roles (if using OIDC) |
| `OIDC_ADMIN_ROLE_VALUE` | - | Value of the admin role (if using OIDC) |
| `OIDC_USERNAME_CLAIM` | - | OIDC claim containing the username (if using OIDC) |
| `OIDC_AUTO_CREATE_USER` | - | Whether user should be auto-created in local database (if using OIDC) |


See `.env.example` for all options and details.


## Conventions
- Packages: lowercase, no underscores
- Types: PascalCase
- Functions: camelCase
- Constants: UPPER_CASE_WITH_UNDERSCORES

## License
GPL v3 — see root LICENSE
