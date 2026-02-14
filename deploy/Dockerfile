# Production Dockerfile for CI/CD
# Multi-stage build: build stage + production stage

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o config-service ./cmd/main.go

# Stage 2: Production
FROM alpine:3.18

# Install ca-certificates and wget for health check
RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/config-service .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./config-service"]

