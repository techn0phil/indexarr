# Docker Quick Reference — Indexarr

## Common Commands

### Build & Run

```bash
# Build the image locally
docker build -t indexarr:latest .

# Run with docker compose (recommended)
docker compose up -d

# Run standalone container
docker run -d \
  -p 8787:8787 \
  -v indexarr_data:/app/data \
  -v /mnt/movies:/data/movies \
  -v /mnt/series:/data/series \
  -e TMDB_API_KEY=your_key \
  -e TVDB_API_KEY=your_key \
  -e RADARR_URL=http://radarr:7878 \
  -e SONARR_URL=http://sonarr:8989 \
  indexarr:latest

# Pull from GitHub Container Registry
docker pull ghcr.io/techn0phil/indexarr:latest

# Use development docker-compose (builds locally)
cd /path/to/indexarr
docker compose -f docker-compose.dev.yml up -d
```

### Docker Compose Files

Indexarr provides two docker-compose configurations:

**docker-compose.yml** (Production)
- Pulls pre-built image from GitHub Container Registry (`ghcr.io/techn0phil/indexarr:latest`)
- Use for production deployments and faster container startup
- Requires: `.env` file with API keys and paths

**docker-compose.dev.yml** (Development)
- Builds image locally from source code
- Use when developing or testing local changes
- Rebuilds frontend and backend on every `docker compose up`
- Useful for testing without pushing to registry

### Management

```bash
# View logs (production)
docker compose logs -f

# View logs (development)
docker compose -f docker-compose.dev.yml logs -f

# Restart services
docker compose restart

# Stop services
docker compose down

# Stop and remove volumes (CAUTION: deletes database)
docker compose down -v

# Execute command in container
docker compose exec indexarr sh

# Check mediainfo version
docker compose exec indexarr mediainfo --Version
```

### Debugging

```bash
# Check container status
docker compose ps

# View container logs (last 100 lines)
docker compose logs --tail=100

# Follow logs in real-time
docker compose logs -f indexarr

# Inspect container
docker inspect indexarr

# Check if mediainfo is available
docker compose exec indexarr which mediainfo

# Test backend API (via Nginx proxy)
curl http://localhost:8787/api/stats

# Test frontend
curl http://localhost:8787/

# Test backend directly (bypass Nginx, if needed)
curl http://localhost:8080/api/stats
```

## File Structure

```
indexarr/
├── Dockerfile                   # Multi-stage build (Node 26.7.0, Go 1.27.0, Alpine 3.24.1)
├── docker-compose.yml           # Production: pulls pre-built image from ghcr.io
├── docker-compose.dev.yml       # Development: builds image locally from source
├── .dockerignore                # Optimize build context
├── nginx.conf                   # Nginx reverse proxy (listens on 8787)
├── entrypoint.sh                # Container startup script (creates user, manages permissions)
└── .github/workflows/
    └── docker-build.yml         # CI/CD pipeline to ghcr.io
```

### Dockerfile Stages

The Dockerfile uses a **3-stage build** for optimized image size and security:

**Stage 1: Frontend Builder** (Node.js 26.7.0)
- Builds React/Vite frontend to static files
- Output: `/app/frontend/` directory

**Stage 2: Backend Builder** (Go 1.27.0)
- Compiles Go backend with CGO (SQLite support)
- Statically linked binary for Alpine
- Output: `/indexarr` binary

**Stage 3: Runtime** (Alpine Linux 3.24.1)
- Minimal base image with only runtime dependencies
- Includes: Nginx, mediainfo, SQLite libs, ca-certificates
- Copies compiled artifacts from stages 1 & 2

## Environment Variables

### Build-Time Arguments

Used during `docker build` (in `docker-compose.dev.yml`):

| Argument | Default | Description |
|----------|---------|-------------|
| `UID` | `1000` | User ID for the `appuser` account created at container runtime |
| `GID` | `1000` | Group ID for the `appuser` group created at container runtime |

**Example** (docker-compose.dev.yml):
```dockerfile
build:
  context: .
  dockerfile: Dockerfile
  args:
    UID: ${UID:-1000}
    GID: ${GID:-1000}
```

### Runtime Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Backend server port (internal) |
| `DB_PATH` | `/app/data/indexarr.db` | SQLite database file path |
| `MEDIAINFO_PATH` | `/usr/bin/mediainfo` | Path to mediainfo binary |
| `UID` | `1000` | User ID used to run indexarr |
| `GID` | `1000` | Group ID used to run indexarr |
| `TMDB_API_KEY` | - | TMDB API key (required) |
| `TVDB_API_KEY` | - | TVDB API key (required) |
| `RADARR_URL` | http://radarr:7878 | Radarr URL (optional) |
| `RADARR_API_KEY` | - | Radarr API key (optional) |
| `RADARR_PATH_MAPPING` | - | Path mapping between Radarr and Indexarr |
| `SONARR_URL` | http://sonarr:8989 | Sonarr URL (optional) |
| `SONARR_API_KEY` | - | Sonarr API key (optional) |
| `SONARR_PATH_MAPPING` | - | Path mapping between Sonarr and Indexarr |
| `MEDIA_LIBRARY_PATHS` | /data/movies,/data/series | **Deprecated** use `MOVIES_LIBRARY_PATHS` and `SERIES_LIBRARY_PATHS` instead |
| `MOVIES_LIBRARY_PATHS` | /data/movies | Comma-separated movies folder paths |
| `SERIES_LIBRARY_PATHS` | /data/series | Comma-separated series folder paths |
| `SKIP_FOLDERS` | - | Comma-separated list of folder names to skip during scanning |
| `IGNORE_FILE_PATTERN` | - | Regular expression pattern to ignore certain files during scanning |
| `SCAN_INTERVAL` | `24` | Hours between automatic scans |
| `SCAN_TIMEOUT` | `30` | Scan timeout in minutes |
| `DETECTION_LANGUAGE` | `en` | Language code for media detection (e.g., "en", "fr") |
| `METADATA_LANGUAGE` | `en` | Language code for metadata fetching (e.g., "en", "fr") |
| `LOG_LEVEL` | `INFO` | Logging level |
| `AUTH_MODE` | `none` | Authentication mode (none, simple, or oidc) |
| `AUTH_ADMIN_USERNAME` | - | Username of administrator account |
| `AUTH_ADMIN_PASSWORD` | - | Password of administrator account |
| `AUTH_SESSION_SECRET` | - | A random secret (auto-generated if not provided) |
| `AUTH_SESSION_MAX_AGE` | `168` | Maximum session validity (default to 168 hours, ie. 7 days) |
| `TZ` | `UTC` | Container timezone |
| `GIN_MODE` | `release` | Go server mode |

## Volume Mounts

### Database Persistence (Required)

```yaml
volumes:
  data:
    name: indexarr-data      # Explicit volume name for CLI commands
    driver: local
```

**In service**:
```yaml
services:
  indexarr:
    volumes:
      - data:/app/data       # Reference by service volume alias
```

When referencing volumes from CLI, use the explicit name `indexarr-data`:
```bash
docker volume ls | grep indexarr-data
docker volume inspect indexarr-data
docker volume rm indexarr-data  # CAUTION: Deletes all data
```

### Media Library Access (Optional)

```yaml
volumes:
  - /host/path/to/movies:/data/movies:ro
  - /host/path/to/series:/data/series:ro
```

## Ports

| Port | Service | Description |
|------|---------|-------------|
| `8787` | Nginx | Frontend + API proxy on host (Nginx reverse proxy) |
| `8080` | Backend | Go API server (internal only, accessed via Nginx) |

### Architecture

The container architecture uses Nginx as a reverse proxy to provide a single entry point:

1. **Nginx (port 8787)**: Reverse proxy that serves:
   - Static frontend files from `/app/frontend`
   - API requests are proxied to backend on `/api/*`
   - Health checks on `/health`

2. **Go Backend (port 8080)**: Only accessible internally:
   - All API endpoints (scan, stats, movies, series, etc.)
   - Connected via Nginx proxy

This design provides a clean single port experience while keeping frontend and backend properly separated. All external traffic goes through port 8787 (Nginx).

## Health Check

The container includes a health check that runs every 30 seconds:

```bash
wget --no-verbose --tries=1 -O /dev/null http://localhost:8787/health
```

The health endpoint is proxied through Nginx to the backend Go server.

### Entrypoint Script

The `entrypoint.sh` script runs at container startup and performs critical setup:

1. **User/Group Creation**: Creates `appuser` with configurable UID/GID
   - Allows proper file ownership matching your host media library
   - Example: Set `UID=1000 GID=1000` to match your user/group IDs
2. **Permission Management**: Fixes ownership of app directories
3. **Dependency Verification**: Checks that mediainfo is available
4. **Configuration Logging**: Prints environment variables for debugging

After setup, services (Nginx + backend) run as the non-root `appuser` account.

### Checking Health Status

```bash
# Check overall container health
docker inspect --format='{{.State.Health.Status}}' indexarr

# View health check details
docker inspect --format='{{json .State.Health}}' indexarr | jq '.'
```

## CI/CD Pipeline

Push to `main` branch triggers automatic build and push to GitHub Container Registry:

1. Multi-architecture build (linux/amd64, linux/arm64)
2. Push to `ghcr.io/techn0phil/indexarr:latest`
3. Also tagged with commit SHA

## Troubleshooting

### Container won't start

```bash
# Check logs
docker compose logs indexarr

# Verify environment variables
docker compose config

# Check if port 8787 is available
sudo netstat -tlnp | grep 8787

# Check if port 8080 (backend) is in use
sudo netstat -tlnp | grep 8080
```

### Database issues

```bash
# Check volume exists
docker volume ls | grep indexarr

# Inspect volume
docker volume inspect indexarr-data

# Backup database
docker cp indexarr:/app/data/indexarr.db ./indexarr.db.backup
```

### Mediainfo not found

```bash
# Verify mediainfo is installed
docker compose exec indexarr which mediainfo

# Check version
docker compose exec indexarr mediainfo --Version

# Test mediainfo on a file
docker compose exec indexarr mediainfo /path/to/file.mkv
```

### Frontend not loading

```bash
# Check if Nginx is running
docker compose exec indexarr ps aux | grep nginx

# Test Nginx config
docker compose exec indexarr nginx -t

# Check frontend files exist
docker compose exec indexarr ls -la /app/frontend/

# Test Nginx directly on port 8787
curl http://localhost:8787/
```

### API not responding

```bash
# Check if backend is running
docker compose exec indexarr ps aux | grep indexarr

# Test backend via Nginx proxy (preferred)
curl http://localhost:8787/api/stats

# Test backend directly (bypass Nginx, internal only)
docker compose exec indexarr curl http://127.0.0.1:8080/api/stats

# Check Nginx proxy logs
docker compose exec indexarr tail -f /var/log/nginx/error.log
```

## Security Notes

- Container runs as non-root user (`appuser`, UID 1000)
- Media libraries should be mounted read-only (`:ro`)
- API keys stored in environment variables (use `.env` file, never commit)
- Database stored in Docker volume (persistent across restarts)

## Performance

### Resource Limits

Add to `docker compose.yml`:

```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 1G
    reservations:
      cpus: '0.5'
      memory: 256M
```

### Build Cache

GitHub Actions uses GitHub Cache for faster builds:

```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```
