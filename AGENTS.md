# Indexarr — AI Agent Guide

**Indexarr** is a media library management application inspired by Sonarr and Radarr. It provides centralized catalog management for movies and TV series with detailed tracking of media file properties, library statistics, and advanced filtering.

**Status**: Production-ready (~95% complete). Core features implemented, authentication and multi-language support added, remaining work is polish and optimization.

**Stack**: React + TypeScript (frontend) + Go (backend) + SQLite (database) + Docker (deployment)

---

## Quick Links

- **Main README**: [README.md](README.md) — Installation, features, project structure
- **Implementation Plan**: [plan.md](plan.md) — Detailed phase-by-phase implementation status
- **Docker Guide**: [DOCKER.md](DOCKER.md) — Container management, common commands
- **Design System**: [Design System](#design-system) — Colors, CSS variables, badges
- **UI Mockups**: [docs/prototype/medialib_v5.html](docs/prototype/medialib_v5.html) — Complete design reference
- **Backend Docs**: [backend/go/README.md](backend/go/README.md)
- **Frontend Docs**: [frontend/react/README.md](frontend/react/README.md)
- **Database Issues**: [/memories/repo/database-locking-analysis.md](/memories/repo/database-locking-analysis.md) — SQLite locking gotchas

---

## Quick Start for AI Agents

### Development Commands

```bash
# Backend (from backend/go/)
go run ./cmd/server              # Dev server on :8080

# Frontend (from frontend/react/)
npm run dev                      # Dev server on :5173 (proxies API to :8080)

# Docker (from root)
docker compose up -d             # Full stack on :8787

# Tests
go test ./...                    # Backend tests
npm test                         # Frontend tests (when implemented)
```

### Environment Setup

Copy `.env.example` and configure:

**Media Import (choose at least one)**:
- **Mode 1 (Radarr/Sonarr)**: Set `RADARR_URL` + `RADARR_API_KEY` and/or `SONARR_URL` + `SONARR_API_KEY`
- **Mode 2 (Filesystem)**: Set `MOVIES_LIBRARY_PATHS` and/or `SERIES_LIBRARY_PATHS`
- **Both modes can be mixed**: e.g., Radarr for movies + filesystem for series

**Metadata APIs (required)**:
- `TMDB_API_KEY` — The Movie Database API key
- `TVDB_API_KEY` — TV Database API key

**Authentication (optional, defaults to no auth)**:
- `AUTH_MODE` — Authentication mode: `none` (default), `simple`, or `oidc`
- `AUTH_ADMIN_USERNAME` — Admin username (required if `AUTH_MODE=simple`)
- `AUTH_ADMIN_PASSWORD` — Admin password (required if `AUTH_MODE=simple`)
- `AUTH_SESSION_SECRET` — Secret key for JWT signing (generate random string, required if auth enabled)
- `AUTH_SESSION_MAX_AGE` — Session duration in hours (default: 168 = 7 days)

**Internationalization**:
- Frontend supports 5 languages: English, Deutsch, Español, Français, Italiano
- Language preference saved in localStorage, defaults to browser language
- Switch language via LanguageToggle component in sidebar

### Database

- **SQLite** with WAL mode enabled
- Location: `backend/go/indexarr.db` (dev) or `/app/data/indexarr.db` (Docker)
- Migrations: Auto-run on startup via `golang-migrate`
- **Purge endpoint**: `POST /api/purge` (keeps schema, deletes all data)

---

## What's Implemented

### Backend (Go) — ~98% Complete

**✅ Core Services**:
- **Dual Import Architecture**: Radarr/Sonarr API clients + filesystem scanner (mix and match)
- **Media Catalog**: Movie/series CRUD with filtering, search, pagination
- **File Scanner**: Periodic/manual scans with WebSocket progress broadcast
- **Mediainfo Parser**: Extracts codec, resolution, HDR, bitrate, audio/subtitle tracks using mediainfo CLI
- **TMDB/TVDB Clients**: Metadata enrichment with per-scan caching
- **Statistics Service**: Library totals, disk usage, 4K percentage, problem counts
- **Scheduler**: Configurable scan intervals (cron-like)
- **Authentication Service**: JWT-based auth with support for "none", "simple", and "oidc" modes
- **User Management**: User creation, password changes, role-based access control

**✅ API Endpoints**:
```
# Public endpoints
GET  /api/auth/config                   # Get auth mode configuration
POST /api/auth/login                    # Login with credentials

# Protected endpoints
GET  /api/auth/me                       # Get current user info
POST /api/auth/change-password          # Change user password
GET  /api/auth/users                    # List users (admin only)
POST /api/auth/users                    # Create user (admin only)
DELETE /api/auth/users/:id              # Delete user (admin only)
POST /api/auth/logout                   # Logout (clear auth cookie)

# Media endpoints
GET  /api/movies                        # List with filters (status, resolution, codec, audio, HDR, search)
GET  /api/movies/:id                    # Movie details + cast + mediainfo
POST /api/movies/:id/refresh            # Refresh single movie metadata
GET  /api/series                        # List series with filters
GET  /api/series/:id                    # Series details + seasons + episodes
POST /api/series/:id/refresh            # Refresh single series metadata
POST /api/scan                          # Trigger full scan (all media)
POST /api/scan/movies                   # Trigger movies-only scan
POST /api/scan/series                   # Trigger series-only scan
GET  /api/scan/status                   # Current scan status
POST /api/scan/stop                     # Stop running scan
GET  /api/scan/ws                       # WebSocket for real-time scan progress
GET  /api/stats                         # Library statistics
GET  /api/config                        # Get configuration (import mode, library paths)
POST /api/purge                         # Purge all data (keep schema)
GET  /health                            # Health check endpoint
```

**✅ Database Schema**:
- **movies**: Metadata, file info, external IDs, poster, status
- **series**, **seasons**, **episodes**: Full TV series tracking
- **video_tracks**, **audio_tracks**, **subtitle_tracks**: Technical metadata per file
- **cast**: Actor names, roles, avatars
- **scan_status**: Scan progress tracking
- **Indexes**: Optimized for common queries

### Frontend (React) — ~93% Complete

**✅ Pages**:
- **ListFilms**: Grid/list view, infinite scroll, multi-filter chips, stat cards, search
- **ListSeries**: Grid/list view, infinite scroll, filters
- **MovieDetail**: Hero section, cast grid, mediainfo tracks table, refresh button
- **SeriesDetail**: Hero section, season tabs, episode list with technical details
- **LoginPage**: User authentication with JWT-based login
- **UsersPage**: User management interface (admin only)

**✅ Components**:
- **Sidebar**: Fixed 210px navigation with active state, badge counts
- **Topbar**: Search bar, breadcrumb support, user menu
- **MovieCard/SeriesCard**: Poster placeholders, status bar, technical badges
- **StatCard**: Library statistics
- **FilterChip/FilterModal**: Multi-select filters with "Clear" and "Apply" buttons
- **ViewToggle**: Grid/list switcher with localStorage persistence
- **ScanStatusCard**: Real-time scan progress via WebSocket
- **ThemeToggle**: Light/dark mode toggle
- **LanguageToggle**: Multi-language support (en, de, es, fr, it)
- **UserMenu**: User profile and logout menu

**✅ Features**:
- Infinite scroll pagination (custom `useInfiniteList` hook)
- Multi-criteria filtering with comma-separated OR logic (e.g., `?resolution=3840,1920`)
- Real-time WebSocket updates for scan progress with exponential backoff reconnection
- Dark mode with CSS variables throughout
- Type-safe API client and interfaces
- JWT-based authentication with cookie persistence
- Multi-language support (i18n) with 5 languages
- Language preference saved in localStorage
- User management interface for admins
- Protected routes based on authentication status

**⚠️ Missing/Incomplete**:
- Filter persistence — not saved in URL or localStorage (resets on page change)
- Error boundaries — no React error boundaries
- Accessibility — limited ARIA labels and keyboard navigation

### DevOps — ~95% Complete

**✅ Docker**:
- Multi-stage build: Node (frontend) + Go (backend) + Alpine runtime
- Nginx reverse proxy: Frontend static files + backend API proxy on port 80
- Health checks, volume mounts, environment variables, non-root user
- CI/CD: GitHub Actions workflow builds multi-arch images → GitHub Container Registry

**⚠️ Missing**:
- Monitoring/metrics (no Prometheus, Grafana, etc.)

---

## Architecture Overview

### Import Mode Architecture

**Unique Feature**: Flexible dual import mode — choose per media type.

**Mode 1: Radarr/Sonarr Integration**
- Backend calls Radarr/Sonarr API to fetch existing library
- Extracts mediainfo from files Radarr/Sonarr already knows about
- Supports path mapping for Docker mounts (`RADARR_PATH_MAPPING`, `SONARR_PATH_MAPPING`)
- Automatically removes stale items deleted from Radarr/Sonarr

**Mode 2: Filesystem Scanner**
- Backend walks directory trees looking for video files (mkv, mp4, avi, etc.)
- Extracts mediainfo directly from discovered files
- Enriches metadata via TMDB/TVDB APIs
- No dependency on external services

**Mix and Match**:
- Movies via Radarr + Series via filesystem scanner
- Movies via filesystem + Series via Sonarr
- Enable both modes simultaneously

**Configuration**:
```bash
# Radarr integration
RADARR_URL=http://radarr:7878
RADARR_API_KEY=your_key
RADARR_PATH_MAPPING=/downloads:/mnt/media  # Optional path translation

# Sonarr integration
SONARR_URL=http://sonarr:8989
SONARR_API_KEY=your_key
SONARR_PATH_MAPPING=/downloads:/mnt/media

# Filesystem scanning
MOVIES_LIBRARY_PATHS=/data/movies,/data/movies2
SERIES_LIBRARY_PATHS=/data/series

# Metadata APIs (required)
TMDB_API_KEY=your_key
TVDB_API_KEY=your_key

# Scheduling
SCAN_INTERVAL=24  # Hours between scans (0=disabled)
```

### Data Flow

1. **Scan Trigger** (manual via API or scheduled via cron)
2. **Importer** (Radarr/Sonarr API call OR filesystem walk)
3. **Mediainfo Extraction** (run mediainfo CLI, parse JSON output)
4. **Metadata Enrichment** (TMDB/TVDB API calls with per-scan caching)
5. **Database Write** (SQLite with transactions)
6. **WebSocket Broadcast** (real-time progress to connected clients)
7. **Frontend Update** (React components re-render with new data)
1. **Media Catalog Service** — Movie/Series CRUD, metadata from TMDB/TVDB
2. **File Scanner** — Discover media files and trigger metadata extraction
3. **Mediainfo Parser** — Extract codec, resolution, HDR, bitrate, audio tracks from files
4. **Search & Filter Engine** — Multi-criteria filtering by status, resolution, codec, audio, HDR
5. **Statistics Service** — Library totals, disk usage, 4K percentage, health metrics
6. **API Layer** — RESTful endpoints for frontend consumption

**Data Model**:
```
Movies
  ├─ Metadata (title, year, duration, genres, TMDB rating, synopsis)
  ├─ Technical (file path, codec, resolution, HDR, audio tracks, bitrate)
  ├─ Status (available, missing, problem)
  └─ Cast information

Series
  ├─ Metadata (title, years, seasons count, TVDB rating)
  ├─ Seasons
  │   ├─ Episodes (number, title, status, file info, size)
  │   └─ Aggregated stats (available/missing count)
  └─ File metadata
```

---

## Frontend — React Conventions

### Naming Conventions

**Components**: PascalCase
- `MovieCard.tsx` — Single movie card in grid/list
- `FilterChip.tsx` — Filter chip with dropdown modal
- `DetailHero.tsx` — Hero section with poster and metadata
- `StatCard.tsx` — Statistic card (total, disk usage, 4K %, problems)
- `EpisodeRow.tsx` — Episode list row with status and details
- `Sidebar.tsx`, `Topbar.tsx`, `SearchBar.tsx`

**Files & Directories**:
- Components: `src/components/` — React components
- Pages: `src/pages/` — Full page components (ListFilms, ListSeries, MovieDetail, SeriesDetail)
- Utilities: `src/utils/` — Helper functions
- Hooks: `src/hooks/` — Custom React hooks (useFilter, usePagination, etc.)
- Styles: Component-scoped CSS modules (*.module.css)
- API: `src/api/` — API client functions
- Types: `src/types/` — TypeScript interfaces

**CSS Classes**: kebab-case
- `.detail-hero`, `.nav-item`, `.stat-card`, `.filter-chip`, `.episode-row`
- Avoid nested deep selectors; prefer CSS modules for scoping

**Variables & Constants**: camelCase
- `filterOptions`, `statusColors`, `apiBaseUrl`

### State Management

**Current Implementation**: Context API via `useAppContext` hook
- Manages navigation state (current page, back navigation)
- Filter state is local to components (not persisted)
- View preference stored in localStorage (films-view, series-view)
- Theme managed via CSS variables and localStorage

**Future Consideration**: If filter state becomes complex, consider migrating to Zustand or Redux for better state persistence and URL synchronization.

### Component Patterns

**Filter Chips**:
- Default state: `border: 0.5px solid var(--color-border-tertiary)`, background secondary
- Active state (has value): Green background `#E1F5EE`, border `#5DCAA5`, text `#085041`, with badge counter
- Click opens modal with multi-select options, "Clear" and "Apply" buttons

**Cards**:
- Movie/Series cards: Poster placeholder, title, badges, status bar at bottom
- Stat cards: Label (uppercase), value (bold), sub-label
- Info cards: White/dark background, subtle borders, contained sections

**Detail Pages**:
- Hero section: Poster (left), title + rating + metadata + synopsis (right), action buttons
- Two-column layout: Main content (left, larger), sidebar (right, 260px)
- Section headers: Uppercase, small font, muted color

**Badges**:
- 4K: `#FAEEDA` background / `#633806` text
- 1080p: `#EAF3DE` background / `#27500A` text
- Dolby Vision: `#EEEDFE` background / `#3C3489` text
- HDR10+: `#E6F1FB` background / `#0C447C` text
- Missing: `#FCEBEB` background / `#791F1F` text
- Codec: Secondary background with tertiary border

### Key Interactions

- **Search hotkey**: "/" focuses search input (Material Design pattern)
- **Filter modal**: Click chip → modal opens → select options → click "Apply"
- **View toggle**: Grid (default) ↔ List view persisted in state
- **Back navigation**: Detail page back button returns to list with filters preserved
- **Dynamic breadcrumbs**: Updates based on current page/detail item
- **Responsive layout**: Sidebar fixed (210px), content area flex with overflow handling

---

## Backend — Go Conventions

### Package Structure

```
backend/go/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── services/
│   │   ├── auth.go              # Authentication Service (JWT)
│   │   ├── broadcaster.go       # WebSocket broadcaster
│   │   ├── extractor.go         # Mediainfo extraction
│   │   ├── filesystem_scanner.go # Filesystem scanning
│   │   ├── importer.go          # Media import orchestration
│   │   ├── parser.go            # JSON/metadata parsing
│   │   ├── radarr_client.go     # Radarr API client
│   │   ├── radarr_importer.go   # Radarr importer
│   │   ├── scanner.go           # Main scanner service
│   │   ├── scheduler.go         # Scheduled scans
│   │   ├── sonarr_client.go     # Sonarr API client
│   │   ├── sonarr_importer.go   # Sonarr importer
│   │   ├── tmdb.go              # TMDB API client
│   │   └── tvdb.go              # TVDB API client
│   ├── models/
│   │   ├── episode.go
│   │   ├── filter.go
│   │   ├── mediainfo.go
│   │   ├── movie.go
│   │   ├── scan.go
│   │   ├── series.go
│   │   └── user.go
│   ├── api/
│   │   ├── auth_handlers.go     # Auth HTTP handlers
│   │   ├── handlers.go          # Media HTTP handlers
│   │   ├── middleware.go        # Auth middleware
│   │   ├── routes.go            # Route definitions
│   │   └── websocket.go         # WebSocket handling
│   ├── config/                  # Configuration management
│   ├── repository/              # Database/persistence layer
│   │   └── migrations/          # SQL migrations
│   └── utils/                   # Helper functions
├── docs/
│   └── MIGRATIONS.md            # Database migration guide
├── go.mod
├── go.sum
└── README.md
```

### Naming Conventions

**Packages**: lowercase, no underscores
- `services`, `models`, `api`, `repository`, `config`

**Types & Interfaces**: PascalCase
- `type Movie struct`, `type MediaService interface`, `type FilterCriteria struct`

**Functions**: camelCase
- `func (m *Movie) GetTechSpecs()`, `func scanMediaDirectory(path string)`

**Constants**: UPPER_CASE_WITH_UNDERSCORES
- `const DEFAULT_PAGE_SIZE = 50`, `const CODEC_H265 = "H.265"`

### API Patterns

**RESTful Endpoints** (examples):
- `GET /api/movies` — List movies with filtering
- `GET /api/movies/:id` — Get movie details
- `GET /api/series` — List series
- `GET /api/series/:id` — Get series details
- `GET /api/series/:id/seasons/:season/episodes` — List episodes in season
- `POST /api/scan` — Trigger media file scan
- `GET /api/stats` — Get library statistics

**Request/Response Pattern**:
```go
type ListRequest struct {
    Page        int    `query:"page" default:"1"`
    PageSize    int    `query:"page_size" default:"50"`
    Status      string `query:"status"`      // available, missing, problem
    Resolution  string `query:"resolution"`  // 4K, 1080p, etc.
    Codec       string `query:"codec"`       // H.264, H.265, etc.
    Audio       string `query:"audio"`       // codec name
    HDR         string `query:"hdr"`         // Dolby Vision, HDR10+, etc.
    Sort        string `query:"sort"`        // title, year, added, etc.
}

type ApiResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

### Data Models

**Movie**:
- ID, Title, Year, Duration, Synopsis, Genres
- Rating (TMDB), Popularity
- FileInfo (path, codec, resolution, HDR, bitrate, audio tracks)
- Status (available, missing, problem)
- External IDs (TMDB, IMDb)
- Cast (actor name, role, avatar)

**Series**:
- ID, Title, Years (start-end), Season count, Episode count
- Rating, Popularity
- Status (complete, ongoing, upcoming)
- External IDs (TVDB, IMDb)
- Seasons → Episodes

**Episode**:
- SeasonNumber, EpisodeNumber, Title, Duration
- FileInfo (codec, resolution, HDR, audio, bitrate, size)
- Status (available, missing)

**MediaInfo** (file-level metadata):
- VideoTracks: codec, resolution, HDR, bitrate, fps, color space
- AudioTracks: codec, channels, sample rate, bitrate, language
- SubtitleTracks: language, format (SRT, ASS, etc.)

### Testing

**Unit Tests**:
- Mediainfo parser: Parse test samples, verify codec/resolution/HDR extraction
- Filter logic: Test filtering by status, resolution, codec combinations
- Stats calculation: Verify aggregations (total size, 4K count, etc.)

**Integration Tests**:
- API endpoints: Test HTTP request/response with mock data
- Database operations: CRUD operations on media records

**Test Files**: `*_test.go` in same package

---

## Design System

### Colors

**Primary Brand**:
- Primary: `#1D9E75` (teal, active state)
- Light variant: `#E1F5EE` (background for active chips)
- Accent: `#5DCAA5` (borders for active chips)
- Dark accent: `#085041` (text for active chips)
- Badge: `#9FE1CB` (nav badge background when active)

**Status Indicators**:
- OK / Available: `#1D9E75` (green)
- Warning / Partial: `#EF9F27` (orange)
- Missing / Problem: `#E24B4A` (red)

**Technical Badges** (background / text):
- 4K: `#FAEEDA` / `#633806` (amber)
- 1080p: `#EAF3DE` / `#27500A` (green)
- Dolby Vision: `#EEEDFE` / `#3C3489` (violet)
- HDR10+: `#E6F1FB` / `#0C447C` (blue)
- Missing: `#FCEBEB` / `#791F1F` (red)

### CSS Variables

Use CSS custom properties for light/dark mode compatibility. Define in `:root`:

```css
:root {
  /* Backgrounds */
  --color-background-primary: #FFFFFF;
  --color-background-secondary: #F5F5F5;
  --color-background-tertiary: #EBEBEB;

  /* Text */
  --color-text-primary: #000000;
  --color-text-secondary: #606060;
  --color-text-tertiary: #999999;

  /* Borders */
  --color-border-tertiary: #D9D9D9;
  --color-border-secondary: #B3B3B3;

  /* Dimensions */
  --border-radius-md: 6px;
  --border-radius-lg: 8px;

  /* Typography */
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-mono: "Courier New", monospace;
}

@media (prefers-color-scheme: dark) {
  :root {
    --color-background-primary: #1a1a1a;
    --color-background-secondary: #2a2a2a;
    --color-background-tertiary: #3a3a3a;
    --color-text-primary: #FFFFFF;
    --color-text-secondary: #CCCCCC;
    --color-text-tertiary: #999999;
    --color-border-tertiary: #404040;
    --color-border-secondary: #606060;
  }
}
```

### Typography

**Font**: System font stack (`-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`)

**Font Sizes & Weights**:
- 22px / weight 500 — Page titles (hero section)
- 15px / weight 500 — Logo name
- 13px / weight 400 — Body text (default)
- 12px / weight 500 — Section headers, nav items, card titles
- 11px / weight 500 — Chips, stat labels, badges
- 10px / weight 500 — Uppercase labels, secondary text

**Restrictions**:
- Use only weight 400 (regular) and 500 (medium)
- No gradients, no box shadows, no emojis

### Borders & Spacing

**Borders**: `0.5px solid var(--color-border-tertiary)` throughout

**Border Radius**:
- Elements (buttons, chips, inputs): `var(--border-radius-md)` = 6px
- Cards, containers: `var(--border-radius-lg)` = 8px
- Search pill, badges: `99px` (fully rounded)

**Spacing** (from mockup):
- Sidebar width: 210px
- Main sidebar item padding: 7px 18px
- Card gaps: 10-16px
- Content padding: 16-24px

---

## Directory Structure & File Organization

```
indexarr/
├── .github/                     # GitHub resources
│   ├── agents/                  # Specialized AI agents for development
│   ├── skills/                  # Custom AI skills (go-testing, react-testing, etc.)
│   └── workflows/               # GitHub Actions CI/CD
├── backend/                     # Go backend
│   └── go/
│       ├── cmd/server/          # Server entry point
│       ├── internal/
│       │   ├── api/             # HTTP handlers, routes, WebSocket, auth
│       │   ├── config/          # Configuration management
│       │   ├── models/          # Data models (Movie, Series, Episode, User)
│       │   ├── repository/      # Database layer (SQLite)
│       │   │   └── migrations/  # SQL migrations (golang-migrate format)
│       │   └── services/        # Business logic (auth, scanner, importer, APIs)
│       ├── docs/
│       │   └── MIGRATIONS.md    # Database migration guide
│       ├── go.mod               # Go module (Go 1.25+)
│       └── README.md            # Backend documentation
├── frontend/                    # React frontend
│   └── react/
│       ├── src/
│       │   ├── components/      # UI components (cards, filters, layout, etc.)
│       │   ├── pages/           # Page components (ListFilms, MovieDetail, etc.)
│       │   ├── api/             # API client with JWT auth support
│       │   ├── hooks/           # Custom hooks (useInfiniteList, useAppContext)
│       │   ├── i18n/            # Internationalization configuration
│       │   ├── styles/          # CSS modules with design system variables
│       │   ├── types/           # TypeScript interfaces (auth, user, media)
│       │   └── App.tsx          # Root component
│       ├── public/locales/      # i18n language files (en, de, es, fr, it)
│       ├── package.json         # Node.js dependencies (React 19, Vite, i18next)
│       └── README.md            # Frontend documentation
├── samples/                     # Sample data for testing
│   ├── tmdb/                    # TheMovieDB API samples
│   │   ├── movies/              # Movie metadata samples
│   │   └── series/              # Series metadata samples
│   ├── tvdb/                    # TheTVDB API samples
│   │   ├── episodes/            # Episode metadata samples
│   │   ├── movies/              # Movie metadata samples
│   │   └── series/              # Series metadata samples
│   ├── fake-movies.sh           # Generate fake movie files for testing
│   ├── fake-series.sh           # Generate fake series files for testing
│   └── mediainfo-output.json    # Sample mediainfo extraction output
├── docs/
│   └── DOCKER.md                # Docker container management guide
├── ux-ui/                       # UI/UX design and specifications
│   └── medialib_v5.html         # Complete HTML/CSS design mockup
├── AGENTS.md                    # AI agent and skill customization guide
├── docker-compose.yml           # Production Docker Compose configuration
├── docker-compose.dev.yml       # Development Docker Compose configuration
├── Dockerfile                   # Multi-stage build (Node + Go + Alpine runtime)
├── nginx.conf                   # Nginx reverse proxy configuration
└── LICENSE                      # GPL v3
```

---

## Build & Development Commands

### Backend (Go)

Always run go commands from `backend/go` folder.

```bash
# Navigate to backend
cd backend/go

# Initialize module (if not exists)
go mod init indexarr

# Install dependencies
go mod download
go mod tidy

# Run server (development)
go run ./cmd/server

# Build executable
go build -o indexarr-backend ./cmd/server

# Run tests
go test ./...

# Run with verbose output
go run -v ./cmd/server

# Format code
go fmt ./...

# Lint code
golangci-lint run ./...
```

### Frontend (React)

```bash
# Navigate to frontend
cd frontend/react

# Initialize with Vite (if not exists)
npm create vite@latest . -- --template react

# Install dependencies
npm install

# Development server (hot reload)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run tests
npm test

# Format/lint code
npm run lint
npm run format
```

---

## Key Gotchas & Patterns

### Frontend (React)

1. **Authentication Flow**: `useAppContext` manages auth state (`authMode`, `isAuthenticated`, `authLoading`). LoginPage redirects to main app on successful auth. Protected routes check `isAuthenticated` before rendering.

2. **WebSocket Reconnection**: WebSocket uses exponential backoff (up to 5s between retries) and only connects when `isAuthenticated`. Manual reconnection fallback required if connection drops.

3. **i18n Integration**: Language preference persists in localStorage and syncs with i18next. All UI strings use `useTranslation()` hook with namespace keys (e.g., `t('sidebar.movies')`). Language files in `public/locales/{lang}/`.

4. **Filter State Persistence**: When user applies filters, badge counter updates and filtered results display. When returning to list from detail page, filters should be preserved.

5. **Fixed Sidebar Layout**: Sidebar is 210px fixed; main content area uses `flex: 1` with overflow handling. Ensure responsive behavior on smaller screens.

6. **Dynamic Breadcrumbs**: Breadcrumb updates based on current page and detail item. On list pages, breadcrumb is empty; on detail pages, shows "Context / Item Name".

7. **Search Hotkey**: "/" triggers focus on search input (Material Design pattern). Prevent form submission on Enter in search input.

8. **Dark Mode via CSS Variables**: All colors use CSS variables. Test both light and dark themes during development. No hardcoded color values.

9. **View Toggle State**: Grid/list toggle affects card layout significantly. Persist view preference in localStorage.

10. **Modal Outside Click**: Filter chips open modals that should close on outside click or "Cancel" button.

11. **Image Placeholders**: Movie/series posters use placeholder with initials (first letter(s) of title) on secondary background. Posters load from TMDB later.

### Backend (Go)

1. **Authentication Modes**: Three modes supported:
   - `none`: No authentication (default for backward compatibility)
   - `simple`: Username/password with JWT tokens (requires `AUTH_ADMIN_USERNAME`, `AUTH_ADMIN_PASSWORD`, `AUTH_SESSION_SECRET`)
   - `oidc`: OpenID Connect (future implementation)
   
   Auth middleware checks JWT cookie (`auth_token`) and sets user context. Disables auth checks if mode is `none`.

2. **JWT Token Management**: Tokens signed with `AUTH_SESSION_SECRET`, expiration configurable via `AUTH_SESSION_MAX_AGE` (default: 168 hours = 7 days). Token stored in HttpOnly cookie, cleared on logout.

3. **SQLite Locking**: See [/memories/repo/database-locking-analysis.md](/memories/repo/database-locking-analysis.md) for detailed analysis. Key points:
   - WAL mode enabled with 5s busy timeout
   - Connection pool limited to avoid contention
   - Long transactions can cause "database is locked" errors
   - Batch scan status updates instead of per-file updates

4. **Path Mapping**: Radarr/Sonarr paths may differ from local filesystem (Docker mounts). Use `RADARR_PATH_MAPPING` and `SONARR_PATH_MAPPING` to translate paths (e.g., `/downloads:/mnt/media`).

5. **Per-Scan Caching**: Scanner caches TVDB lookups per scan to avoid redundant API calls. Cache is cleared at start of each scan.

6. **Memory Management**: Extractor calls `unix.Fadvise(FADV_DONTNEED)` after reading files to clear Linux page cache and avoid memory bloat during large scans.

7. **Nil Service Checks**: Importers can be `nil` if not configured (e.g., no Radarr URL). Always check `if movieImporter != nil` before calling methods.

8. **Mediainfo Timeouts**: Mediainfo extraction has 30s timeout per file. Large files or network mounts may timeout.

9. **Status Calculation**: Status (available, missing, problem) derived from file existence. Check happens during scan, not on-demand.

10. **Filter Combinations**: Backend supports comma-separated OR logic: `?resolution=3840,1920` means "3840 OR 1920".

11. **WebSocket Broadcast**: All connected clients receive scan progress updates. No per-client tracking. Broadcast happens without auth gating (public endpoint).

12. **Database Migrations**: Uses golang-migrate format (000001_description.up.sql / .down.sql). Migrations auto-run on startup. User table added in migration 000009.

---

## Implementation Specifications

For detailed UI/UX specifications, design mockups, and full implementation guidance, see:

- **HTML/CSS Mockup** (complete design with all pages): [docs/prototype/medialib_v5.html](docs/prototype/medialib_v5.html)

---

## Known Issues & Future Work

### Frontend
- [ ] Filter persistence (URL or localStorage)
- [ ] No React error boundaries
- [ ] Accessibility (ARIA labels, keyboard navigation)
- [ ] Loading skeleton screens
- [ ] OIDC authentication support
- [ ] Admin panel for scan scheduling and configuration

### Backend
- [ ] Full JSON structured logging (currently uses zerolog with mixed format)
- [ ] Unit and integration tests for auth, scanner services
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Metrics/monitoring (Prometheus endpoints)
- [ ] OIDC provider integration

### DevOps
- [ ] Health check on actual database connectivity (not just stats endpoint)
- [ ] Database backup/restore procedures
- [ ] Secrets management for AUTH_SESSION_SECRET in Docker deployment

## Related Resources

- **Repository Memory**: [/memories/repo/](/memories/repo/) — Critical gotchas and lessons learned
- **Implementation Plan**: [plan.md](plan.md) — Phase status and completion tracking
- **Design Mockups**: [docs/prototype/medialib_v5.html](docs/prototype/medialib_v5.html) — Complete UI reference

---

**Created**: 5 mai 2026  
**License**: GPL v3 ([LICENSE](LICENSE))
