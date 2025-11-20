#!/bin/bash
# run-with-update.sh
# Checks for updates, applies them, then starts the server with logging

APP_NAME="${1:-sandstorm-tracker}"
LOG_FILE="${2:-logs/update-serve.log}"

# Get the directory where this script is located (scripts folder)
# Then go up one level to the app directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(dirname "$SCRIPT_DIR")"
APP_PATH="$APP_DIR/$APP_NAME"
LOG_DIR="$APP_DIR/logs"

# Ensure logs directory exists
mkdir -p "$LOG_DIR"

LOG_PATH="$LOG_DIR/$LOG_FILE"

# Helper function to log messages
write_log() {
    local message="$1"
    local level="${2:-INFO}"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local log_entry="[$timestamp] [$level] $message"
    
    echo "$log_entry" | tee -a "$LOG_PATH"
}

# Check if app exists
if [ ! -f "$APP_PATH" ]; then
    write_log "Error: $APP_NAME not found at $APP_PATH" "ERROR"
    exit 1
fi

# Make sure binary is executable
chmod +x "$APP_PATH"

write_log "Starting $APP_NAME with update check..." "INFO"

# Try to update
write_log "Checking for updates and applying if available..." "INFO"
"$APP_PATH" update 2>&1 | tee -a "$LOG_PATH"

if [ $? -ne 0 ]; then
    write_log "Update check completed (no updates or update skipped)" "INFO"
else
    write_log "Update applied successfully" "INFO"
fi

# Start the server
write_log "Starting server..." "INFO"
"$APP_PATH" serve 2>&1 | tee -a "$LOG_PATH"

exit_code=$?
write_log "Server stopped with exit code: $exit_code" "WARNING"

exit $exit_code
