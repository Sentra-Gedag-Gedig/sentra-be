# Build stage
FROM golang:1.23-alpine AS builder

# Set build arguments
ARG GO_ENV=production
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

# Set environment variables for build
ENV GO111MODULE=on \
    CGO_ENABLED=${CGO_ENABLED} \
    GOOS=${GOOS} \
    GOARCH=${GOARCH} \
    GOTOOLCHAIN=auto

# Install build dependencies
RUN apk --no-cache add \
    git \
    ca-certificates \
    tzdata \
    upx

# Set working directory
WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 go build \
    -a \
    -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o main ./cmd/app/main.go && \
    # Compress binary to reduce size
    upx --best --lzma main

# Development stage
FROM alpine:3.18 AS development

# Install development runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    wget \
    bash \
    git \
    air \
    delve && \
    update-ca-certificates

# Create app user
RUN adduser -D -g '' -h /home/appuser -s /bin/bash appuser

# Set working directory
WORKDIR /home/appuser/app

# Copy from builder (for initial setup)
COPY --from=builder /app/main /home/appuser/
COPY --from=builder /app/.env.example /home/appuser/.env

# Create necessary directories
RUN mkdir -p storage/logs && \
    mkdir -p config && \
    chown -R appuser:appuser /home/appuser

# Switch to non-root user
USER appuser

# Expose application port and debugger port
EXPOSE 8080 2345

# Development health check (more lenient)
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/ || wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Development command with hot reload
CMD ["sh", "-c", "if [ -f /home/appuser/app/main.go ]; then air; else ./main; fi"]

# Production base stage
FROM alpine:3.18 AS production-base

# Install minimal runtime dependencies for production
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl \
    wget && \
    update-ca-certificates && \
    # Remove package cache to reduce image size
    rm -rf /var/cache/apk/*

# Create app user with minimal permissions
RUN adduser -D -g '' -h /home/appuser -s /sbin/nologin appuser

# Production stage
FROM production-base AS production

# Set working directory
WORKDIR /home/appuser

# Copy only the necessary files from builder
COPY --from=builder --chown=appuser:appuser /app/main .

# Create necessary directories and files
RUN mkdir -p storage/logs && \
    mkdir -p config && \
    touch .env && \
    chown -R appuser:appuser . && \
    chmod +x main && \
    # Set proper permissions for security
    chmod 750 . && \
    chmod 640 .env && \
    chmod 750 storage && \
    chmod 755 storage/logs

# Switch to non-root user
USER appuser

# Expose only the application port
EXPOSE 8080

# Production health check (stricter)
HEALTHCHECK --interval=30s --timeout=10s --start-period=90s --retries=5 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Set production environment
ENV ENV=production \
    GIN_MODE=release \
    GO_ENV=production

# Production command
CMD ["./main"]

# Slim production stage (even smaller image)
FROM scratch AS production-slim

# Copy CA certificates for HTTPS requests
COPY --from=alpine:3.18 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=alpine:3.18 /usr/share/zoneinfo /usr/share/zoneinfo

# Copy passwd file for user information
COPY --from=production-base /etc/passwd /etc/passwd
COPY --from=production-base /etc/group /etc/group

# Copy the binary
COPY --from=builder --chown=1000:1000 /app/main /main

# Set user
USER 1000:1000

# Expose port
EXPOSE 8080

# Set production environment
ENV ENV=production \
    GIN_MODE=release \
    GO_ENV=production

# Health check (simplified for scratch image)
HEALTHCHECK --interval=30s --timeout=10s --start-period=90s --retries=5 \
    CMD ["/main", "--healthcheck"] || exit 1

# Command to run
ENTRYPOINT ["/main"]