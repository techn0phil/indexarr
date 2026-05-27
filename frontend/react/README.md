
# Indexarr Frontend (React)

This is the React frontend for Indexarr, a media library management application inspired by Sonarr and Radarr. It provides a modern, responsive UI for browsing and managing movies and TV series, with advanced filtering, technical metadata, and real-time scan status.

## Features

- Grid and list views for movies and TV series
- Multi-criteria filtering (status, resolution, codec, audio, HDR)
- Infinite scroll pagination
- Real-time statistics and search
- Detail pages with technical info, cast, and mediainfo tracks
- Responsive sidebar and topbar navigation
- Real-time scan status with WebSocket updates
- Dark mode via CSS variables

## Project Structure

```
frontend/react/
├── src/
│   ├── components/   # UI components (MovieCard, Sidebar, StatCard, etc.)
│   ├── pages/        # Page components (ListFilms, ListSeries, MovieDetail, SeriesDetail)
│   ├── api/          # API client functions (REST endpoints)
│   ├── hooks/        # Custom React hooks (useAppContext, useInfiniteList)
│   ├── styles/       # CSS modules and variables (dark mode, layout)
│   ├── types/        # TypeScript interfaces (Movie, Series, MediaInfo, etc.)
│   └── App.tsx       # Root component
├── package.json      # Dependencies and scripts
├── tsconfig.json     # TypeScript config
├── vite.config.ts    # Vite config (API proxy)
└── README.md         # This file
```

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```
2. Start development server (http://localhost:5173):
   ```bash
   npm run dev
   ```
3. Build for production:
   ```bash
   npm run build
   ```
4. Lint code:
   ```bash
   npm run lint
   ```

## Main Components & Pages

- **MovieCard, SeriesCard**: Display movie/series poster, title, status, and technical badges
- **MovieCardList, SeriesCardList**: Grid/list layouts for media items
- **ListFilms, ListSeries**: Main pages for browsing movies/series with filters, stats, and infinite scroll
- **MovieDetail, SeriesDetail**: Detail pages with hero section, cast, mediainfo, and refresh actions
- **FilterChip, FilterModal**: Multi-select filter chips and modal dialogs
- **StatCard**: Library statistics (totals, disk usage, 4K %)
- **ScanStatusCard**: Real-time scan progress (WebSocket)
- **Sidebar, Topbar**: Navigation and search
- **ThemeToggle, ViewToggle**: Dark mode and grid/list toggle

## API & Data

- All API calls are defined in `src/api/client.ts` and use `/api` endpoints (proxied to backend)
- TypeScript interfaces for all entities in `src/types/index.ts`
- Infinite scroll and filtering handled by `useInfiniteList` hook
- App-wide state (theme, navigation, stats, scan status) managed by `useAppContext` hook

## Design System

- All colors, spacing, and typography use CSS variables (see `src/styles/variables.css`)
- Light/dark mode supported via CSS variables and `ThemeToggle`
- UI/UX specs and mockups: `../../ux-ui/medialib_v5.html`, `../../ux-ui/prompt.md`
- Consistent border radius, spacing, and badge colors for technical metadata

## Conventions

- **Components**: PascalCase (e.g., `MovieCard`, `FilterChip`)
- **CSS classes**: kebab-case, scoped via CSS modules
- **State management**: Context API (`useAppContext`), local state for filters
- **TypeScript**: Strict typing throughout (`strict: true` in `tsconfig.json`)
- **API**: All requests proxied to backend via Vite config

## Development Notes

- Uses React 19, Vite, and TypeScript
- Linting via ESLint (`npm run lint`)
- No test suite yet (`npm test` is a placeholder)
- For backend/API details, see main project README

## License

GPL v3 — see root LICENSE
