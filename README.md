# Indexarr (Mediarr)

Indexarr is a media library management application inspired by Sonarr and Radarr. It provides centralized catalog management for movies and TV series with detailed tracking of media file properties, library statistics, and advanced filtering.

## Features
- Centralized movie and TV series catalog
- Advanced multi-criteria filtering (status, resolution, codec, audio, HDR)
- Real-time statistics (total count, disk space, 4K %, problems)
- Detailed media info (video/audio/subtitle tracks)
- Responsive UI with grid/list views
- RESTful API backend

## Project Structure

```
indexarr/
├── AGENTS.md                # Chat customization guide
├── LICENSE                  # GPL v3
├── backend/                 # Go backend
│   └── go/
│       ├── cmd/server/      # Entry point
│       ├── internal/
│       │   ├── api/         # HTTP handlers
│       │   ├── config/      # Configuration
│       │   ├── models/      # Data models
│       │   ├── repository/  # Database layer
│       │   └── services/    # Business logic
│       ├── go.mod           # Go module
│       └── README.md        # Backend docs
├── frontend/                # React frontend
│   └── react/
│       ├── src/
│       │   ├── components/  # UI components
│       │   ├── pages/       # Page components
│       │   ├── api/         # API client
│       │   ├── hooks/       # Custom hooks
│       │   ├── styles/      # CSS modules
│       │   ├── types/       # TypeScript types
│       │   └── App.tsx      # Root component
│       ├── package.json     # Dependencies
│       └── README.md        # Frontend docs
├── ux-ui/                   # UI/UX design
│   ├── medialib_v4_detail_pages.html  # Full HTML/CSS mockup
│   └── prompt.md            # Implementation specs
```

## Getting Started

### Prerequisites
- Node.js (>=18)
- Go (>=1.20)

### Backend Setup
1. Navigate to backend:
   ```bash
   cd backend/go
   ```
2. Install dependencies:
   ```bash
   go mod download && go mod tidy
   ```
3. Run the server:
   ```bash
   go run ./cmd/server
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

### Frontend Setup
1. Navigate to frontend:
   ```bash
   cd frontend/react
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start development server:
   ```bash
   npm run dev
   ```
4. Run tests:
   ```bash
   npm test
   ```

## Design & Implementation
- **Design system:** See `ux-ui/medialib_v4_detail_pages.html` for full mockups and CSS variables.
- **Implementation guide:** See `ux-ui/prompt.md` for detailed frontend specs.
- **Chat agent customization:** See `AGENTS.md` for agent and workflow details.

## License
GPL v3 — see [LICENSE](LICENSE)

---

For more details, see the [backend README](backend/go/README.md) and [frontend README](frontend/react/README.md).
