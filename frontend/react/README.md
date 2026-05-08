# Indexarr Frontend (React)

This is the frontend for Indexarr, a media library management application inspired by Sonarr and Radarr. It provides a modern, responsive UI for browsing and managing movies and TV series.

## Features
- Grid/list views for movies and tv-shows
- Multi-criteria filtering (status, resolution, codec, audio, HDR)
- Real-time statistics and search
- Detail pages with technical info and cast
- Responsive sidebar and topbar navigation
- Dark mode via CSS variables

## Project Structure
```
frontend/react/
├── src/
│   ├── components/   # UI components (MovieCard, Sidebar, StatCard, etc.)
│   ├── pages/        # Page components (ListFilms, MovieDetail, etc.)
│   ├── api/          # API client functions
│   ├── hooks/        # Custom React hooks
│   ├── styles/       # CSS modules and variables
│   ├── types/        # TypeScript interfaces
│   └── App.tsx       # Root component
├── package.json      # Dependencies
├── tsconfig.json     # TypeScript config
├── vite.config.ts    # Vite config
└── README.md         # This file
```

## Getting Started
1. Install dependencies:
   ```bash
   npm install
   ```
2. Start development server:
   ```bash
   npm run dev
   ```
3. Build for production:
   ```bash
   npm run build
   ```
4. Run tests:
   ```bash
   npm test
   ```

## Design System
- All colors and spacing use CSS variables (see `src/styles/variables.css`).
- For UI/UX specs, see `../../ux-ui/medialib_v4_detail_pages.html` and `../../ux-ui/prompt.md`.

## Conventions
- Components: PascalCase
- CSS classes: kebab-case, scoped via CSS modules
- State management: Context API (expandable to Zustand/Redux)

## License
GPL v3 — see root LICENSE
