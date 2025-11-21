# GoReleaser optimized Dockerfile for multi-platform builds
# Uses pre-built binaries from GoReleaser build context
# Supports linux/amd64 and linux/arm64

FROM cgr.dev/chainguard/static:latest

WORKDIR /app

# Copy pre-built binary from GoReleaser context
# The $TARGETPLATFORM variable is set by buildx automatically
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/sandstorm-tracker /app/sandstorm-tracker

# Create required directories with proper permissions
RUN mkdir -p /app/pb_data /app/logs && \
    chmod 755 /app /app/pb_data /app/logs && \
    chmod +x /app/sandstorm-tracker

# Use built-in nonroot user (uid 65532)
USER 65532:65532

# Expose PocketBase admin UI and API port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
    CMD ["/app/sandstorm-tracker", "version"]

# Run the application
ENTRYPOINT ["/app/sandstorm-tracker", "serve", "--http", "0.0.0.0:8090"]
