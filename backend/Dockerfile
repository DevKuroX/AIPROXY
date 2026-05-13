# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking (no CGO)
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -ldflags="-s -w" -o /ai_proxy ./cmd/server

# Stage 2: Runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /ai_proxy /app/ai_proxy

# Copy frontend static files (Next.js export)
COPY frontend/out /app/frontend/out

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the default port
EXPOSE 20128

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:20128/health || exit 1

# Run the binary
ENTRYPOINT ["/app/ai_proxy"]
