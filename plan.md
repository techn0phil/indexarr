# Implementation Phases
## PHASE 1: Foundation & MVP Layout (2-3 weeks)

*Start both in parallel — no blocking dependencies*

### Backend (Go):

1. Project setup (module, HTTP server on port 8080)
2. Data models: Movie, Series, Episode, MediaInfo, Cast, FilterCriteria
3. SQLite schema with tables for movies, series, episodes, mediainfo, cast
4. Repository layer for CRUD + filtering
5. API handlers: `GET /api/movies`, `GET /api/series`, `GET /api/series/:id/seasons/:season/episodes`, `GET /api/stats`
6. Router with CORS middleware
7. Seed 20 sample movies + 10 series with complete mediainfo specs

### Frontend (React):

1. Vite + TypeScript setup
2. Type definitions matching Go models
3. API client with fetch functions
4. CSS variables design system (from AGENTS.md) + layout (210px sidebar fixed)
5. Layout components: Sidebar, Topbar (back/breadcrumb/search), Main
6. List pages: FilterBar (non-functional chips), StatCards, MovieCard/SeriesCard in grid
7. Detail pages: Hero section, CastGrid, MediainfoTable, EpisodeList (season tabs)
8. Context for theme, filters, pagination state
9. Mock data (same 20 movies/10 series as backend seed)
10. Simple routing (no React Router yet) with back button history

**Result**: ✅ Working UI with mock data, all pages display

## PHASE 2: Core Interactivity (1-2 weeks)
### Backend:

1. Filter service: Build SQL WHERE from FilterCriteria (status, resolution, codec, audio, HDR, sort)
2. Search service: Full-text search (SQLite FTS5) on title, synopsis
3. Stats service: Aggregate queries (count, disk space, 4K %, problems)
4. Enhance endpoints with filter + search + pagination support

### Frontend:

1. Filter chips interactive: Click → modal opens → multi-select → Apply/Clear
2. Search integration: "/" hotkey, Enter to search, global results
3. Pagination: prev/next buttons or infinite scroll
4. ✅ Remove mock data, fetch from API with filters
5. ✅ Detail pages: Fetch from API, populate all fields
6. ✅ View toggle: Grid ↔ List (persist in localStorage)

**Result**: ✅ Functional filters, search, pagination, real API integration

## PHASE 3: File Scanning & Metadata (2-3 weeks)
### Backend:

1. Install mediainfo CLI
2. Mediainfo parser: Run `mediainfo --Output=JSON`, parse codec/resolution/HDR/bitrate/fps/color space
3. File scanner: Walk library folders, extract mediainfo, query TMDB/TVDB, determine status
4. Cron scheduler: Run scanner every X hours (configurable)
5. Configuration: Load library paths, API keys from .env
6. Endpoints: `POST /api/scan` (trigger), `GET /api/scan/status`

### Frontend:

1. Library path configuration (optional UI in settings)
2. Stats page: Show scan status, last scanned time, library health (missing %, problems %)

**Result**: ✅ Real file discovery, technical metadata, scheduled scanning

## PHASE 4: Polish & Performance (1-2 weeks)
### Backend:

1. TMDB/TVDB caching: Store in SQLite with TTL (refresh after 30 days)
2. Input validation & error handling
3. API key auth: `X-API-Key` header validation
4. Docker: Multi-stage build, mediainfo CLI installed, non-root user

### Frontend:

1. Virtual scrolling: `react-window` for 5000+ items (< 100ms render)
2. Loading skeletons + error boundaries
3. Responsive design: Mobile-first (320px, 768px, 1024px)
4. Docker: Node build → Vite → Nginx
5. Unit tests: Jest for components, Go tests for services

**Result**: ✅ Production-ready, handles large datasets, containerized
