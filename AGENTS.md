# Indexarr (Mediarr) — Chat Customization Guide

**Indexarr** is a media library management application inspired by Sonarr and Radarr. It provides centralized catalog management for movies and TV series with detailed tracking of media file properties, library statistics, and advanced filtering.

**Status**: Pre-implementation. Architecture and UI/UX fully designed; codebase ready for development.

**Stack**: React (frontend) + Go (backend)

---

## Quick Navigation

- **Frontend** (React): [frontend/react/](#frontend--react-conventions)
- **Backend** (Go): [backend/go/](#backend--go-conventions)
- **Design System**: [Design System](#design-system) — Colors, CSS variables, badges, typography
- **Design Specs**: [ux-ui/medialib_v4_detail_pages.html](ux-ui/medialib_v4_detail_pages.html) — Complete HTML/CSS mockup with all styling
- **Implementation Guide**: [ux-ui/prompt.md](ux-ui/prompt.md) — Detailed specifications for frontend implementation

---

## Architecture Overview

### Frontend (React)

**Purpose**: Single-page application for browsing and managing media library.

**Key Pages**:
1. **List Films** — Grid/list view of movies with filters, stat cards, and search
2. **List Series** — Grid/list view of TV series with filters and stats
3. **Movie Detail** — Hero section, cast grid, mediainfo table (video/audio/subtitle tracks), sidebar with metadata and external IDs
4. **Series Detail** — Hero section, season tabs, episode list with status and technical details

**Navigation**:
- Fixed sidebar (210px width) with logo, main sections (Films, Séries, Récents, Statistiques, Problèmes), with active item highlight and item count badges
- Top bar with back button (on detail pages), breadcrumb, and Material Design search pill (/ hotkey for focus)
- Dynamic breadcrumbs on detail pages

**Core Features**:
- Multi-criteria filtering: Status, Resolution, Codec, Audio, HDR, Sort
- Grid/list view toggle
- Status badges on cards (green=ok, orange=warning, red=missing)
- Technical badges: 4K, 1080p, Dolby Vision, HDR10+, Missing, Codec
- Real-time stat cards (total count, disk space, 4K %, problems)
- Search across all media

### Backend (Go)

**Purpose**: RESTful API and services for media management.

**Key Services**:
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

**To be decided during implementation** — Consider these approaches:
- **Redux**: For complex state shared across many components (filters, sidebar, search)
- **Zustand**: Lighter alternative to Redux with simpler API
- **Context API**: For simpler state needs (theme, user settings)

**Recommendation**: Start with Context API for theme/settings; use Zustand or Redux when filter/search state becomes complex.

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
/home/philippe/sources/indexarr/
├── LICENSE                                    # GPL v3
├── AGENTS.md                                  # This file
│
├── backend/
│   └── go/                                    # Go backend (empty, ready)
│       ├── cmd/server/main.go                 # Entry point (to create)
│       ├── internal/services/                 # Service layer
│       ├── internal/models/                   # Data structures
│       ├── internal/api/                      # HTTP handlers
│       ├── go.mod                             # Module definition
│       └── README.md                          # Backend docs
│
├── frontend/
│   └── react/                                 # React frontend (empty, ready)
│       ├── src/
│       │   ├── components/                    # Reusable components
│       │   ├── pages/                         # Page components
│       │   ├── api/                           # API client functions
│       │   ├── utils/                         # Utilities
│       │   ├── hooks/                         # Custom hooks
│       │   ├── types/                         # TypeScript interfaces
│       │   └── App.tsx                        # Root component
│       ├── package.json                       # Dependencies
│       └── README.md                          # Frontend docs
│
└── ux-ui/
    ├── medialib_v4_detail_pages.html          # Complete HTML/CSS mockup
    ├── prompt.md                              # Implementation specification
    └── (*.png mockup images)
```

---

## Build & Development Commands

### Backend (Go)

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

1. **Mediainfo Parsing**: Most complex logic. Video/audio/subtitle tracks are nested structures. Parse FFprobe or mediainfo CLI output carefully.

2. **Filter Combinations**: Support multi-criteria filtering (e.g., "4K + Dolby Vision + H.265"). Use query parameters: `?resolution=4K&hdr=DV&codec=H.265`.

3. **Status Calculation**: Status (available, missing, problem) is derived from file existence check. Missing if file path doesn't exist.

4. **Pagination**: List endpoints should support pagination. Default page size = 50.

5. **External API Rate Limiting**: TMDB and TVDB have rate limits. Implement caching of metadata.

6. **Error Responses**: Return consistent error format: `{ "success": false, "error": "error message" }`.

7. **File Path Security**: Validate file paths to prevent directory traversal attacks.

---

## Implementation Specifications

For detailed UI/UX specifications, design mockups, and full implementation guidance, see:

- **HTML/CSS Mockup** (complete design with all pages): [ux-ui/medialib_v4_detail_pages.html](ux-ui/medialib_v4_detail_pages.html)
- **Implementation Prompt** (detailed feature specs): [ux-ui/prompt.md](ux-ui/prompt.md)

---

## Related Chat Customizations

Future skills to create during implementation:

- **`indexarr-backend-go`** (skill) — Deep dive into Go service architecture, API contracts, mediainfo parsing patterns, testing strategies
- **`indexarr-frontend-react`** (skill) — React component patterns, state management decisions, filter UX implementation, styling with CSS modules

---

**Created**: 5 mai 2026  
**License**: GPL v3 ([LICENSE](LICENSE))
