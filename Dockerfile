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

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    nginx \
    mediainfo \
    sqlite-libs \
    tzdata \
    wget

# Create app user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Create necessary directories
RUN mkdir -p /app/data /app/frontend /var/log/nginx /var/lib/nginx/tmp \
    && chown -R appuser:appuser /app /var/log/nginx /var/lib/nginx

WORKDIR /app

# Copy backend binary from builder
COPY --from=backend-builder --chown=appuser:appuser /indexarr /app/indexarr

# Copy frontend build from builder
COPY --from=frontend-builder --chown=appuser:appuser /build/frontend/dist /app/frontend

# Copy nginx configuration
COPY --chown=appuser:appuser nginx.conf /etc/nginx/nginx.conf

# Copy entrypoint script
COPY --chown=appuser:appuser entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8787

# Set environment variables
ENV SERVER_PORT=8080 \
    DB_PATH=/app/data/indexarr.db \
    MEDIAINFO_PATH=/usr/bin/mediainfo \
    GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8787/api/stats || exit 1

# Start services
ENTRYPOINT ["/app/entrypoint.sh"]
