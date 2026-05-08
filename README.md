# Indexarr (Mediarr)

Indexarr is a media library management application inspired by Sonarr and Radarr. It provides centralized catalog management for movies and TV series with detailed tracking of media file properties, library statistics, and advanced filtering.

> [!WARNING]
> **Disclaimer:** This software is provided as-is, without any warranty. Use at your own risk. The authors and contributors are not responsible for any data loss, damage, or issues resulting from the use of this application.
>
> **Note:** This application has been developed with intensive help from AI coding agents (including GitHub Copilot and similar tools).


![Main movie page screenshot](ux-ui/movies.png)

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

### Quick Start installation

The easiest and recommended way to run Indexarr is with Docker Compose. The provided `docker-compose.yml` is production-ready with automatic restarts, data persistence, and proper networking.

#### Prerequisites
- Docker and Docker Compose installed
- TMDB and TVDB API keys (optional, but recommended for full metadata)

#### Installation Steps

1. **Create docker-compose file:**

   Download or copy content from https://github.com/pschmucker/indexarr/blob/main/docker-compose.yml
   
2. **Configure environment variables:**
   
   Download or copy content from https://github.com/pschmucker/indexarr/blob/main/.env.example

   Create a `.env` file with your configuration:
   ```bash
   cp .env.example .env
   # Edit .env with your TMDB/TVDB API keys and media library paths
   ```

3. **Start the application:**
   ```bash
   docker compose up -d
   ```
   
   This will:
   - Pull the latest image from GitHub Container Registry
   - Create a persistent volume for application data
   - Mount your media libraries (read-only)
   - Start the service with automatic restart on failure

4. **Verify it's running:**
   ```bash
   docker compose ps
   docker compose logs -f
   ```

5. **Access the application:**
   - **Frontend:** http://localhost:8787
   - **API:** http://localhost:8787/api
   - **Health check:** http://localhost:8787/health

#### Configuration Reference

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `TMDB_API_KEY` | - | No | TMDB API key for movie metadata ([get here](https://www.themoviedb.org/settings/api)) |
| `TVDB_API_KEY` | - | No | TVDB API key for tv-shows metadata ([get here](https://www.thetvdb.com/api-information)) |
| `MOVIES_PATH` | - | Yes | Comma-separated paths to movies folder (e.g., `/movies` or `/mnt/nas/movies,/external/movies`) |
| `TV_SHOWS_PATH` | - | Yes | Comma-separated paths to tv-shows folder (e.g., `/tv-shows` or `/mnt/nas/tv,/external/tv`) |
| `SCAN_INTERVAL` | 24 | No | Library scan interval in hours |
| `SCAN_TIMEOUT` | 30 | No | Scan timeout in minutes |
| `TZ` | UTC | No | Timezone (e.g., `Europe/Paris`, `America/New_York`) |

#### Common Operations

**View logs in real-time:**
```bash
docker compose logs -f
```

**Stop the application:**
```bash
docker compose down
```

**Stop and remove all data:**
```bash
docker compose down -v
```

**Restart the application:**
```bash
docker compose restart
```

**Update to latest version:**
```bash
docker compose pull
docker compose up -d
```

#### Using Pre-built Image from GitHub Container Registry

```bash
docker pull ghcr.io/pschmucker/indexarr:latest
docker run -d -p 8787:8787 \
      -e TMDB_API_KEY=fffffffffffffffff \
      -e TVDB_API_KEY=fffffffffffffffff \
      -e MOVIES_PATH=/movies \
      -e TV_SHOWS_PATH=/tv-shows \
      -v indexarr_data:/app/data ghcr.io/pschmucker/indexarr:latest
```

### Manual Development Setup

#### Prerequisites
- Node.js (>=18)
- Go (>=1.21)
- mediainfo CLI

#### Backend Setup
1. Navigate to backend:
   ```bash
   cd backend/go
   ```
2. Install dependencies:
   ```bash
   go mod download && go mod tidy
   ```
3. Create `.env` file from example:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```
4. Run the server:
   ```bash
   go run ./cmd/server
   ```
5. Run tests:
   ```bash
   go test ./...
   ```

#### Frontend Setup
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

### Building Docker Image Locally

```bash
# Build the image
docker build -t indexarr:latest .

# Run the container
docker run -d -p 80:80 -v indexarr_data:/app/data indexarr:latest
```

## Design & Implementation
- **Design system:** See `ux-ui/medialib_v4_detail_pages.html` for full mockups and CSS variables.
- **Implementation guide:** See `ux-ui/prompt.md` for detailed frontend specs.
- **Chat agent customization:** See `AGENTS.md` for agent and workflow details.

## License
GPL v3 — see [LICENSE](LICENSE)

---

For more details, see the [backend README](backend/go/README.md) and [frontend README](frontend/react/README.md).
