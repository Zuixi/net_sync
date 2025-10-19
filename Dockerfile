# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s -extldflags "-static"' -o easy-sync ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p uploads data && \
    chown -R appuser:appgroup /app

# Copy binary from builder stage
COPY --from=builder /app/easy-sync .

# Copy web assets
COPY --from=builder /app/web ./web

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3280

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3280/health || exit 1

# Run the application
CMD ["./easy-sync"]