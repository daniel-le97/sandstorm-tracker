#!/bin/bash
# run-with-update.sh
# Wrapper script to check for updates, then start the app
# All app logging goes to app.log and the database

APP_NAME="${1:-sandstorm-tracker}"

# Get app directory (parent of scripts folder)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(dirname "$SCRIPT_DIR")"
APP_PATH="$APP_DIR/$APP_NAME"

# Check if app exists
if [ ! -f "$APP_PATH" ]; then
    echo "Error: $APP_NAME not found at $APP_PATH" >&2
    exit 1
fi

# Make sure binary is executable
chmod +x "$APP_PATH"

echo "Checking for updates..."
"$APP_PATH" update

echo "Starting server..."
"$APP_PATH" serve

exit $?
