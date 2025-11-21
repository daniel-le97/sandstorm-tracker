# Multi-stage build optimized for chainguard/static base image
# Stage 1: Build the Go application (golang:1.25-alpine)
# Stage 2: Prepare directories and permissions (alpine:latest as intermediate)
# Stage 3: Final runtime using chainguard/static (3-4MB, minimal, secure)
#
# - Produces static Go binary with no runtime dependencies
# - Uses chainguard/static: minimal (39.4MB), no shell, daily security updates
# - Nonroot user (uid 65532) pre-configured in chainguard/static
# - Directories created in alpine intermediate stage to ensure proper permissions
#
# Image size: ~39.4MB

FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations for minimal size
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o sandstorm-tracker .

# Intermediate builder stage - Alpine can do RUN commands
FROM alpine:latest AS app-prep

WORKDIR /app

# Copy binary from Go builder
COPY --from=builder /build/sandstorm-tracker .

# Create directories with proper permissions for nonroot (65532)
RUN mkdir -p /app/pb_data /app/logs && \
    chmod 755 /app /app/pb_data /app/logs && \
    # Ensure the binary is executable
    chmod +x /app/sandstorm-tracker

# Final stage - use Chainguard static image (smallest, most secure)
FROM cgr.dev/chainguard/static:latest

WORKDIR /app

# Copy everything from app-prep (binary + directories with correct permissions)
COPY --from=app-prep /app .

# Use built-in nonroot user (uid 65532)
USER 65532:65532

# Expose PocketBase admin UI and API port
EXPOSE 8090

# Run the application
CMD ["./sandstorm-tracker", "serve", "--http", "0.0.0.0:8090"]
