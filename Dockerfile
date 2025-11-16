FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sandstorm-tracker .

# Final stage - use minimal base image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 tracker && \
    adduser -D -u 1000 -G tracker tracker

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/sandstorm-tracker .

# Create volumes directory
RUN mkdir -p /app/pb_data /app/logs && \
    chown -R tracker:tracker /app

# Switch to non-root user
USER tracker

# Expose PocketBase admin UI and API port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8090/api/health || exit 1

# Run the application
CMD ["./sandstorm-tracker", "serve", "--http", "0.0.0.0:8090"]
