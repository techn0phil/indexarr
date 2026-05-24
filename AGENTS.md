# Indexarr — AI Agent Guide

**Indexarr** is a media library management application inspired by Sonarr and Radarr. It provides centralized catalog management for movies and TV series with detailed tracking of media file properties, library statistics, and advanced filtering.

**Status**: Production-ready (~90% complete). Core features implemented, polish and optimization ongoing.

**Stack**: React + TypeScript (frontend) + Go (backend) + SQLite (database) + Docker (deployment)

---

## Quick Links

- **Main README**: [README.md](README.md) — Installation, features, project structure
- **Implementation Plan**: [plan.md](plan.md) — Detailed phase-by-phase implementation status
- **Docker Guide**: [DOCKER.md](DOCKER.md) — Container management, common commands
- **Design System**: [Design System](#design-system) — Colors, CSS variables, badges
- **UI Mockups**: [ux-ui/medialib_v4_detail_pages.html](ux-ui/medialib_v4_detail_pages.html) — Complete design reference
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
- **Import Mode 1 (Radarr/Sonarr)**: Set `RADARR_URL` + `RADARR_API_KEY` and/or `SONARR_URL` + `SONARR_API_KEY`
- **Import Mode 2 (Filesystem)**: Set `MOVIES_LIBRARY_PATHS` and/or `SERIES_LIBRARY_PATHS`
- **Both modes can be mixed**: e.g., Radarr for movies + filesystem for series
- **Required**: `TMDB_API_KEY` and `TVDB_API_KEY` for metadata enrichment

### Database

- **SQLite** with WAL mode enabled
- Location: `backend/go/indexarr.db` (dev) or `/app/data/indexarr.db` (Docker)
- Migrations: Auto-run on startup via `golang-migrate`
- **Purge endpoint**: `POST /api/purge` (keeps schema, deletes all data)

---

## What's Implemented

### Backend (Go) — ~95% Complete

**✅ Core Services**:
- **Dual Import Architecture**: Radarr/Sonarr API clients + filesystem scanner (mix and match)
- **Media Catalog**: Movie/series CRUD with filtering, search, pagination
- **File Scanner**: Periodic/manual scans with WebSocket progress broadcast
- **Mediainfo Parser**: Extracts codec, resolution, HDR, bitrate, audio/subtitle tracks using mediainfo CLI
- **TMDB/TVDB Clients**: Metadata enrichment with per-scan caching
- **Statistics Service**: Library totals, disk usage, 4K percentage, problem counts
- **Scheduler**: Configurable scan intervals (cron-like)

**✅ API Endpoints**:
```
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

### Frontend (React) — ~85% Complete

**✅ Pages**:
- **ListFilms**: Grid/list view, infinite scroll, multi-filter chips, stat cards, search
- **ListSeries**: Grid/list view, infinite scroll, filters
- **MovieDetail**: Hero section, cast grid, mediainfo tracks table, refresh button
- **SeriesDetail**: Hero section, season tabs, episode list with technical details

**✅ Components**:
- **Sidebar**: Fixed 210px navigation with active state, badge counts
- **Topbar**: Search bar, breadcrumb support (placeholder)
- **MovieCard/SeriesCard**: Poster placeholders, status bar, technical badges
- **StatCard**: Library statistics
- **FilterChip/FilterModal**: Multi-select filters with "Clear" and "Apply" buttons
- **ViewToggle**: Grid/list switcher with localStorage persistence
- **ScanStatusCard**: Real-time scan progress via WebSocket
- **ThemeToggle**: Light/dark mode

**✅ Features**:
- Infinite scroll pagination (custom `useInfiniteList` hook)
- Multi-criteria filtering with comma-separated OR logic (e.g., `?resolution=3840,1920`)
- Real-time WebSocket updates for scan progress
- Dark mode with CSS variables throughout
- Type-safe API client and interfaces

**⚠️ Missing/Incomplete**:
- Filter persistence — not saved in URL or localStorage (resets on page change)
- Error boundaries — no React error boundaries
- WebSocket auto-reconnect — requires manual page refresh
- Accessibility — no ARIA labels, limited keyboard navigation

### DevOps — ~90% Complete

**✅ Docker**:
- Multi-stage build: Node (frontend) + Go (backend) + Alpine runtime
- Nginx reverse proxy: Frontend static files + backend API proxy on port 80
- Health checks, volume mounts, environment variables, non-root user
- CI/CD: GitHub Actions workflow builds multi-arch images → GitHub Container Registry

**⚠️ Missing**:
- Monitoring/metrics (no Prometheus, Grafana, etc.)
- Structured logging (console logs only, no JSON structured logs)

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
SERIES_LIBRARY_PATHS=/data/tv-shows

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
│   │   ├── media.go             # Media Catalog Service
│   │   ├── scanner.go           # File Scanner
│   │   ├── mediainfo.go         # Mediainfo Parser
│   │   ├── search.go            # Search & Filter Engine
│   │   └── stats.go             # Statistics Service
│   ├── models/
│   │   ├── movie.go
│   │   ├── series.go
│   │   ├── episode.go
│   │   └── mediainfo.go
│   ├── api/
│   │   ├── handlers.go          # HTTP handlers
│   │   └── routes.go            # Route definitions
│   ├── repository/              # Database/persistence layer
│   ├── config/                  # Configuration management
│   └── utils/                   # Helper functions
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
- Status (complete, ongoing, partial)
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
├── LICENSE                                    # GPL v3
├── AGENTS.md                                  # This file
├── README.md                                  # Installation & features
├── plan.md                                    # Phase-by-phase implementation status
├── DOCKER.md                                  # Docker management guide
├── docker-compose.yml                         # Production Docker orchestration
├── Dockerfile                                 # Multi-stage build (frontend + backend)
├── nginx.conf                                 # Nginx reverse proxy config
│
├── backend/go/                                # Go backend
│   ├── cmd/server/main.go                     # Entry point
│   ├── internal/
│   │   ├── api/                               # HTTP handlers, routes, WebSocket
│   │   ├── config/                            # Environment configuration
│   │   ├── models/                            # Data models (Movie, Series, Episode, etc.)
│   │   ├── repository/                        # Database layer (SQLite)
│   │   └── services/                          # Business logic (scanner, importer, TMDB/TVDB)
│   ├── go.mod                                 # Module definition
│   └── indexarr.db                            # SQLite database (dev)
│
├── frontend/react/                            # React frontend
│   ├── src/
│   │   ├── components/                        # UI components
│   │   ├── pages/                             # Page components
│   │   ├── api/client.ts                      # API client
│   │   ├── hooks/                             # Custom hooks (useInfiniteList, useAppContext)
│   │   ├── styles/                            # CSS modules + variables
│   │   └── types/                             # TypeScript interfaces
│   ├── package.json                           # Dependencies
│   └── vite.config.ts                         # Vite configuration
│
└── ux-ui/
    ├── medialib_v4_detail_pages.html          # Complete HTML/CSS mockup
    └── prompt.md                              # Implementation specification
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

1. **Filter State Persistence**: When user applies filters, badge counter updates and filtered results display. When returning to list from detail page, filters should be preserved.

2. **Fixed Sidebar Layout**: Sidebar is 210px fixed; main content area uses `flex: 1` with overflow handling. Ensure responsive behavior on smaller screens.

3. **Dynamic Breadcrumbs**: Breadcrumb updates based on current page and detail item. On list pages, breadcrumb is empty; on detail pages, shows "Context / Item Name".

4. **Search Hotkey**: "/" triggers focus on search input (Material Design pattern). Prevent form submission on Enter in search input.

5. **Dark Mode via CSS Variables**: All colors use CSS variables. Test both light and dark themes during development. No hardcoded color values.

6. **View Toggle State**: Grid/list toggle affects card layout significantly. Persist view preference in localStorage.

7. **Modal Outside Click**: Filter chips open modals that should close on outside click or "Cancel" button.

8. **Image Placeholders**: Movie/series posters use placeholder with initials (first letter(s) of title) on secondary background. Posters load from TMDB later.

### Backend (Go)

1. **SQLite Locking**: See [/memories/repo/database-locking-analysis.md](/memories/repo/database-locking-analysis.md) for detailed analysis. Key points:
   - WAL mode enabled with 5s busy timeout
   - Connection pool limited to avoid contention
   - Long transactions can cause "database is locked" errors
   - Batch scan status updates instead of per-file updates

2. **Path Mapping**: Radarr/Sonarr paths may differ from local filesystem (Docker mounts). Use `RADARR_PATH_MAPPING` and `SONARR_PATH_MAPPING` to translate paths (e.g., `/downloads:/mnt/media`).

3. **Per-Scan Caching**: Scanner caches TVDB lookups per scan to avoid redundant API calls. Cache is cleared at start of each scan.

4. **Memory Management**: Extractor calls `unix.Fadvise(FADV_DONTNEED)` after reading files to clear Linux page cache and avoid memory bloat during large scans.

5. **Nil Service Checks**: Importers can be `nil` if not configured (e.g., no Radarr URL). Always check `if movieImporter != nil` before calling methods.

6. **Mediainfo Timeouts**: Mediainfo extraction has 30s timeout per file. Large files or network mounts may timeout.

7. **Status Calculation**: Status (available, missing, problem) derived from file existence. Check happens during scan, not on-demand.

8. **Filter Combinations**: Backend supports comma-separated OR logic: `?resolution=3840,1920` means "3840 OR 1920".

9. **WebSocket Broadcast**: All connected clients receive scan progress updates. No per-client tracking.

---

## Implementation Specifications

For detailed UI/UX specifications, design mockups, and full implementation guidance, see:

- **HTML/CSS Mockup** (complete design with all pages): [ux-ui/medialib_v4_detail_pages.html](ux-ui/medialib_v4_detail_pages.html)
- **Implementation Prompt** (detailed feature specs): [ux-ui/prompt.md](ux-ui/prompt.md)

---

## Known Issues & Future Work

### Frontend
- [ ] Filter persistence (URL or localStorage)
- [ ] Error boundaries for graceful failure handling
- [ ] WebSocket auto-reconnect on disconnect
- [ ] Accessibility (ARIA labels, keyboard navigation)
- [ ] Loading skeleton screens

### Backend
- [ ] TMDB/TVDB rate limit handling with retry logic
- [ ] Automatic cleanup of deleted files (filesystem scanner)
- [ ] Structured logging (JSON format)
- [ ] Unit and integration tests
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Metrics/monitoring (Prometheus endpoints)

### DevOps
- [ ] Health check on actual database connectivity (not just stats endpoint)
- [ ] Database backup/restore procedures
- [ ] Migration rollback documentation

## Related Resources

- **Repository Memory**: [/memories/repo/](/memories/repo/) — Critical gotchas and lessons learned
- **Implementation Plan**: [plan.md](plan.md) — Phase status and completion tracking
- **Design Mockups**: [ux-ui/medialib_v4_detail_pages.html](ux-ui/medialib_v4_detail_pages.html) — Complete UI reference

---

**Created**: 5 mai 2026  
**License**: GPL v3 ([LICENSE](LICENSE))
