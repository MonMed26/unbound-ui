# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./

# Copy frontend build into the embed directory
COPY --from=frontend-builder /app/frontend/dist ./cmd/server/frontend/dist

RUN CGO_ENABLED=0 GOOS=linux go build -o /unbound-ui ./cmd/server

# Stage 3: Final image
FROM alpine:3.19

RUN apk add --no-cache unbound unbound-libs ca-certificates tzdata

# Create directories
RUN mkdir -p /etc/unbound/unbound.conf.d /var/lib/unbound-ui/blocklist /etc/unbound-ui

# Copy binary
COPY --from=backend-builder /unbound-ui /usr/local/bin/unbound-ui

# Copy default unbound config
COPY docker/unbound.conf /etc/unbound/unbound.conf

# Expose ports
EXPOSE 53/tcp 53/udp 8080/tcp

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/auth/status || exit 1

# Start script
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
