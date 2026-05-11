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
  -v /mnt/tv-shows:/data/tv-shows \
  -e TMDB_API_KEY=your_key \
  -e TVDB_API_KEY=your_key \
  -e RADARR_URL=http://radarr:7878 \
  -e SONARR_URL=http://sonarr:8989 \
  indexarr:latest

# Pull from GitHub Container Registry
docker pull ghcr.io/pschmucker/indexarr:latest
```

### Management

```bash
# View logs
docker compose logs -f

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

# Test backend API
curl http://localhost/api/stats

# Test frontend
curl http://localhost/
```

## File Structure

```
indexarr/
├── Dockerfile              # Multi-stage build (frontend + backend + runtime)
├── docker compose.yml      # Orchestration config with volumes
├── .dockerignore           # Optimize build context
├── nginx.conf              # Nginx proxy config (frontend + API)
├── entrypoint.sh           # Container startup script
└── .github/workflows/
    └── docker-build.yml    # CI/CD pipeline to ghcr.io
```

## Environment Variables

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
| `SONARR_URL` | http://sonarr:8989 | Sonarr URL (optional) |
| `MEDIA_LIBRARY_PATHS` | - | Comma-separated media folder paths |
| `SCAN_INTERVAL` | `24` | Hours between automatic scans |
| `SCAN_TIMEOUT` | `30` | Scan timeout in minutes |
| `TZ` | `UTC` | Container timezone |
| `GIN_MODE` | `release` | Go server mode |

## Volume Mounts

### Database Persistence (Required)

```yaml
volumes:
  - indexarr-data:/app/data
```

### Media Library Access (Optional)

```yaml
volumes:
  - /host/path/to/movies:/data/movies:ro
  - /host/path/to/tv-shows:/data/tv-shows:ro
```

## Ports

| Port | Service | Description |
|------|---------|-------------|
| `80` | Nginx | Frontend + API proxy (exposed) |
| `8080` | Backend | Go API server (internal only) |

## Health Check

The container includes a health check that runs every 30 seconds:

```bash
wget --no-verbose --tries=1 -O /dev/null http://localhost:8787/health
```

Check health status:

```bash
docker inspect --format='{{.State.Health.Status}}' indexarr
```

## CI/CD Pipeline

Push to `main` branch triggers automatic build and push to GitHub Container Registry:

1. Multi-architecture build (linux/amd64, linux/arm64)
2. Push to `ghcr.io/pschmucker/indexarr:latest`
3. Also tagged with commit SHA

## Troubleshooting

### Container won't start

```bash
# Check logs
docker compose logs indexarr

# Verify environment variables
docker compose config

# Check if ports are available
sudo netstat -tlnp | grep -E '(80|8080)'
```

### Database issues

```bash
# Check volume exists
docker volume ls | grep indexarr

# Inspect volume
docker volume inspect indexarr_indexarr_data

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
```

### API not responding

```bash
# Check if backend is running
docker compose exec indexarr ps aux | grep indexarr

# Test backend directly (bypass Nginx)
curl http://localhost:8080/api/stats

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
