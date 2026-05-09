#!/bin/sh
set -e

# Set default UID/GID (can be overridden by environment variables)
UID=${UID:-1000}
GID=${GID:-1000}

echo "========================================="
echo "Starting Indexarr Container"
echo "========================================="

# Create group if it doesn't exist
if ! getent group appuser >/dev/null 2>&1; then
    addgroup -g ${GID} appuser
fi

# Create user if it doesn't exist
if ! getent passwd appuser >/dev/null 2>&1; then
    adduser -D -u ${UID} -G appuser appuser
fi

# Fix ownership of app directories
chown -R ${UID}:${GID} /app /var/log/nginx /var/lib/nginx /tmp/nginx

echo "User Configuration:"
echo "  - UID: ${UID}"
echo "  - GID: ${GID}"

# Print environment info
echo ""
echo "Environment:"
echo "  - Server Port: ${SERVER_PORT}"
echo "  - Database: ${DB_PATH}"
echo "  - Mediainfo: ${MEDIAINFO_PATH}"
echo "  - Media Libraries: ${MEDIA_LIBRARY_PATHS}"

# Verify mediainfo is available
if ! command -v "${MEDIAINFO_PATH}" >/dev/null 2>&1; then
    echo "ERROR: mediainfo not found at ${MEDIAINFO_PATH}"
    exit 1
fi
echo "  - Mediainfo version: $(${MEDIAINFO_PATH} --Version | head -n 1)"

# Create data directory if it doesn't exist
mkdir -p "$(dirname ${DB_PATH})"

# Verify backend binary exists
if [ ! -x /app/indexarr ]; then
    echo "ERROR: Backend binary not found or not executable at /app/indexarr"
    exit 1
fi

# Verify frontend files exist
if [ ! -f /app/frontend/index.html ]; then
    echo "ERROR: Frontend files not found at /app/frontend/"
    exit 1
fi

echo "========================================="
echo "Starting Nginx..."
echo "========================================="

# Start Nginx in background
nginx -t && nginx -g 'daemon off;' &
NGINX_PID=$!

# Wait a moment for Nginx to start
sleep 2

# Check if Nginx started successfully
if ! kill -0 $NGINX_PID 2>/dev/null; then
    echo "ERROR: Nginx failed to start"
    exit 1
fi

echo "Nginx started (PID: $NGINX_PID)"

echo "========================================="
echo "Starting Backend..."
echo "========================================="

# Start backend in foreground as app user
exec su-exec ${UID}:${GID} /app/indexarr
