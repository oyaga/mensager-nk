# =============================================================================
# Chatwoot-Go Unified Dockerfile
# Builds frontend + backend into a single image
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build Frontend (React/Vite)
# -----------------------------------------------------------------------------
FROM node:18-alpine AS frontend-builder

WORKDIR /frontend

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY frontend/ .

# Build frontend (outputs to /frontend/dist)
RUN npm run build

# -----------------------------------------------------------------------------
# Stage 2: Build Backend (Go)
# -----------------------------------------------------------------------------
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# -----------------------------------------------------------------------------
# Stage 3: Production Image
# -----------------------------------------------------------------------------
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy Go binary from backend builder
COPY --from=backend-builder /app/main .

# Copy frontend dist from frontend builder
COPY --from=frontend-builder /frontend/dist ./dist

# Expose port
EXPOSE 8080

# Environment variables (can be overridden)
ENV PORT=8080
ENV GO_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
