# =============================================================================
# Stage 1: Build Frontend (React + Vite)
# =============================================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# Copy package files for dependency caching
COPY frontend/react/package*.json ./
RUN npm ci

# Copy frontend source
COPY frontend/react/ ./

# Build production frontend
RUN npm run build

# =============================================================================
# Stage 2: Build Backend (Go)
# =============================================================================
FROM golang:1.21-alpine AS backend-builder

WORKDIR /build/backend

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go.mod and go.sum for dependency caching
COPY backend/go/go.mod backend/go/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/go/ ./

# Build static binary with proper CGO flags for Alpine/musl
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN CGO_ENABLED=1 GOOS=linux go build -a -tags musl -ldflags="-s -w -extldflags '-static'" -o /indexarr ./cmd/server

# =============================================================================
# Stage 3: Runtime (Alpine with Nginx + mediainfo + Go backend)
# =============================================================================
FROM alpine:latest

# Build arguments for dynamic user/group configuration
ARG UID=1000
ARG GID=1000

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    nginx \
    mediainfo \
    sqlite-libs \
    tzdata \
    wget \
    su-exec

# Create necessary directories (user will be created at runtime via entrypoint.sh)
RUN mkdir -p /app/data /app/frontend /var/log/nginx /var/lib/nginx/tmp \
    && mkdir -p /tmp/nginx/{client_body,proxy_temp,fastcgi_temp,uwsgi_temp,scgi_temp}

WORKDIR /app

# Copy backend binary from builder
COPY --from=backend-builder /indexarr /app/indexarr

# Copy frontend build from builder
COPY --from=frontend-builder /build/frontend/dist /app/frontend

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Copy entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Entrypoint will handle user creation and permission setup at runtime
# Container runs as root initially, entrypoint.sh creates user and switches to it

# Expose ports
EXPOSE 8787

# Set environment variables
ENV SERVER_PORT=8080 \
    DB_PATH=/app/data/indexarr.db \
    MEDIAINFO_PATH=/usr/bin/mediainfo \
    GIN_MODE=release \
    MEDIA_LIBRARY_PATHS=/data/movies,/data/tv-shows

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8787/health || exit 1

# Start services
ENTRYPOINT ["/app/entrypoint.sh"]
