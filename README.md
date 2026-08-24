# Indexarr

**Indexarr** is a media library application inspired by Sonarr and Radarr. It provides a centralized catalog for movies and series with detailed tracking of media file properties, library statistics, and advanced filtering capabilities.

![Main movie page screenshot](ux-ui/movies.png)


## Table of contents

- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation Steps](#installation-steps)
  - [Configuration Reference](#configuration-reference)
  - [Common Operations](#common-operations)
  - [Using Pre-built Image from GitHub Container Registry](#using-pre-built-image-from-github-container-registry)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites-1)
  - [Backend Setup](#backend-setup)
  - [Frontend Setup](#frontend-setup)
  - [Building Docker Image Locally](#building-docker-image-locally)
  - [Project Structure](#project-structure)
  - [Design & Implementation](#design--implementation)
- [Common Issues](#common-issues)
  - [Incorrect Matching](#incorrect-matching)
  - [Media Permissions](#media-permissions)
  - [Extra Files](#extra-files)
- [License](#license)


## Features

- **Centralized movies and series catalog**
  - **Radarr / Sonarr integrations** — Import from existing Radarr / Sonarr libraries
  - **Filesystem scanning** — Discover media from local directories
  - **Blu-ray formats support** — Uncompressed folder and ISO files
- **Accurate media Intelligence**
  - **Detailed media info detection** — video, audio, and subtitle tracks
  - **Multi-criteria filtering** — title, year, status, resolution, codec, audio, HDR
  - **Real-time statistics** — total count, disk space, 4K %, problem counts
- **Easy user experience**
  - **Multi-language support** — English, Français, Español, Italiano, Deutsch
  - **Responsive UI** — Supports desktop, tablet, and mobile devices
  - **User authentication** — Builtin authentication with local users


## Getting started

The easiest and recommended way to run Indexarr is with Docker Compose. The provided `docker-compose.yml` is production-ready with automatic restarts, data persistence, and proper networking.

### Dual Import Modes

Indexarr supports flexible media import with two modes that can be mixed and matched:

1. **Radarr/Sonarr Integration**: Import existing libraries from Radarr and/or Sonarr with automatic metadata enrichment and stale item cleanup
2. **Filesystem Scanning**: Discover media directly from local directories with automatic TMDB/TVDB metadata enrichment

**Example configurations:**
- Movies via Radarr + Series via Sonarr
- Movies via filesystem + Series via Sonarr
- All media via filesystem scanning
- Movies via Radarr + movies via filesystem (both enabled)

### Prerequisites
- Docker and Docker Compose installed
- TMDB and TVDB API keys (optional, but recommended for full metadata)


### Installation steps

1. **Create docker-compose file:**

   Download or copy content from https://github.com/techn0phil/indexarr/blob/main/docker-compose.yml
   
2. **Configure environment variables:**
   
   Download or copy content from https://github.com/techn0phil/indexarr/blob/main/.env.example

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

### Configuration reference

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `TMDB_API_KEY` | - | No | TMDB API key for movie metadata ([get here](https://www.themoviedb.org/settings/api)) |
| `TVDB_API_KEY` | - | No | TVDB API key for series metadata ([get here](https://www.thetvdb.com/api-information)) |
| `MOVIES_HOST_PATH` | - | No | Comma-separated paths to movies folder on the host (e.g., `/movies` or `/mnt/nas/movies,/external/movies`) |
| `SERIES_HOST_PATH` | - | No | Comma-separated paths to series folder on the host (e.g., `/series` or `/mnt/nas/tv,/external/tv`) |
| `MEDIA_LIBRARY_PATHS` | /data/movies,/data/series | No | Comma-separated paths to media on the guest [**Deprecated**: prefer `MOVIES_LIBRARY_PATHS` and `SERIES_LIBRARY_PATHS`] |
| `MOVIES_LIBRARY_PATHS` | /data/movies | No | Comma-separated paths to movies on the guest |
| `SERIES_LIBRARY_PATHS` | /data/series | No | Comma-separated paths to series on the guest |
| `SKIP_FOLDERS` | - | No | Comma-separated list of folder names to skip during scanning |
| `IGNORE_FILE_PATTERN` | - | No | Regular expression pattern to ignore certain files during scanning |
| `RADARR_URL` | http://radarr:7878 | No | Radarr URL |
| `RADARR_API_KEY` | - | No | Radarr API key for importing movies from Radarr |
| `RADARR_PATH_MAPPING` | - | No | Used to map Radarr paths to local paths (e.g. `/movies:/data/movies`) |
| `SONARR_URL` | http://sonarr:8989 | No | Sonarr URL |
| `SONARR_API_KEY` | - | No | Sonarr API key for importing series from Sonarr |
| `SONARR_PATH_MAPPING` | - | No | Used to map Sonarr paths to local paths (e.g. `/series:/data/series`) |
| `SCAN_INTERVAL` | 24 | No | Library scan interval in hours |
| `SCAN_TIMEOUT` | 30 | No | Scan timeout in minutes |
| `DETECTION_LANGUAGE` | en | No | Language code for media detection (e.g., "en", "fr") |
| `METADATA_LANGUAGE` | en | No | Language code for metadata fetching (e.g., "en", "fr") |
| `LOG_LEVEL` | `INFO` | No | Logging level |
| `AUTH_MODE` | `none` | No | Authentication mode (none, simple, or oidc) |
| `AUTH_ADMIN_USERNAME` | - | Yes if `AUTH_MODE` is `simple` | Username of administrator account |
| `AUTH_ADMIN_PASSWORD` | - | Yes if `AUTH_MODE` is `simple` | Password of administrator account |
| `AUTH_SESSION_SECRET` | - | No | A random secret (auto-generated if not provided) |
| `AUTH_SESSION_MAX_AGE` | 168 | No | Maximum session validity (default to 168 hours, ie. 7 days) |
| `TZ` | UTC | No | Timezone (e.g., `Europe/Paris`, `America/New_York`) |
| `UID` | 1000 | No | User ID inside container (match your media library owner) |
| `GID` | 1000 | No | Group ID inside container (match your media library owner) |

### Common operations

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

### Using pre-built image from GitHub Container Registry

```bash
docker pull ghcr.io/techn0phil/indexarr:latest
docker run -d -p 8787:8787 \
      -v indexarr_data:/app/data \
      -v /mnt/movies:/data/movies \
      -v /mnt/series:/data/series \
      -e TMDB_API_KEY=fffffffffffffffff \
      -e TVDB_API_KEY=fffffffffffffffff \
      -e RADARR_URL=http://radarr:7878 \
      -e SONARR_URL=http://sonarr:8989 \
      ghcr.io/techn0phil/indexarr:latest
```

## Development setup

### Prerequisites
- Node.js (>=24)
- Go (>=1.25)
- mediainfo CLI (for video file analysis)

### Backend setup
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

### Frontend setup
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

### Building Docker image locally

```bash
# Build the image
docker build -t indexarr:latest .

# Run the container
docker run -d -p 80:80 -v indexarr_data:/app/data indexarr:latest
```

### Project Structure

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
│   ├── medialib_v5.html         # Complete HTML/CSS design mockup
│   └── prompt.md                # Detailed implementation specifications
├── AGENTS.md                    # AI agent and skill customization guide
├── docker-compose.yml           # Production Docker Compose configuration
├── docker-compose.dev.yml       # Development Docker Compose configuration
├── Dockerfile                   # Multi-stage build (Node + Go + Alpine runtime)
├── nginx.conf                   # Nginx reverse proxy configuration
└── LICENSE                      # GPL v3
```


### Design & implementation
- **Design system:** See `ux-ui/medialib_v5.html` for full mockups and CSS variables.
- **Implementation guide:** See `ux-ui/prompt.md` for detailed frontend specs.
- **Chat agent customization:** See `AGENTS.md` for agent and workflow details.


## Common issues

### Incorrect matching

If your media files are not correctly detected (no poster, duration, wrong links, etc...), review the files and folders naming with the following recommendations:
- **Movies**: file must be in a folder named `{movie title} ({year})`
- **Series**: each series must have a folder named `{series name} ({year})`
- **Episodes**: file must contain the season and episode number with a standard pattern like `S01E05` or `1x05`

If the files and folders naming are correct but still have incorrect matches, check whether the title and year are correct on [TheMovieDB](https://www.themoviedb.org/).

### Media permissions

Indexarr runs as a non-root user inside the container for security. By default, it uses UID 1000 and GID 1000. **If your media library is owned by a different user** (e.g., Radarr, Sonarr, or another service), you must configure `UID` and `GID` to match the owner, or the container won't be able to read your files.

**Why this matters:**
- Indexarr reads media files from mounted volumes (read-only)
- If the container user doesn't have read permission on these files, scans will fail

**How to fix it:**

1. **Find your media library owner:**
   ```bash
   # Check media library ownership
   ls -ld /mnt/media/movies
   # Example output: drwxr-x--- 1220 radarr media-center 77824 May  6 movies
   
   # Get UID and GID of the owner
   id radarr
   # Example output: uid=1041(radarr) gid=100(users) groups=100(users),65541(media-center)
   ```

2. **Override `UID` and `GID` environment variables:**
   ```yaml
   environment:
     UID: 1041
     GID: 65541
   ```

3. **Restart application:**
   ```bash
   # With docker-compose.yml (pre-built image)
   docker compose up -d
   ```

4. **Verify permissions are working:**
   ```bash
   # Check if app is running as correct user
   docker exec indexarr id
   # Should show: uid=1041(appuser) gid=65541(media-center)
   
   # Check if media files are readable
   docker exec indexarr ls -la /data/movies/
   # Should show files, not permission denied errors
   ```

**Note on local builds:**
- When building locally with `docker-compose.dev.yml`, build args set the initial file ownership at build time
- At runtime, the container adjusts file ownership to match `UID` and `GID`
- With pre-built images from ghcr.io, only the runtime environment variables matter


### Extra files

Extra files such as **Behind the scenes**, **interviews**, **trailers**, etc... will likely be incorrectly matched. To avoid that, you can exclude some folders from scanning by setting the `SKIP_FOLDERS` environment variable:

```yaml
environment:
  # Skip the folders based on Plex extras list:
  # https://support.plex.tv/articles/local-files-for-trailers-and-extras/
  SKIP_FOLDERS: Behind The Scenes, Deleted Scenes, Featurettes, Interviews, Scenes, Shorts, Trailers, Other
```

If you need a more advanced way to exclude only some files, you can set the `IGNORE_FILE_PATTERN` environment variable instead and define
the regular expression you want:

```yaml
environment:
  # Ignore all AVI files
  IGNORE_FILE_PATTERN: "\.avi$"
```


## License
GPL v3 — see [LICENSE](LICENSE)

---

For more details, see the [backend README](backend/go/README.md) and [frontend README](frontend/react/README.md).
