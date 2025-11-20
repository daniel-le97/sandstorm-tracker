#!/bin/bash
# setup-service.sh
# Sets up sandstorm-tracker as a Linux systemd service
# Run as root or with sudo

set -e

# Get the app directory (parent directory of this script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_PATH="$(dirname "$SCRIPT_DIR")/scripts"

# If called from scripts directory, use parent
if [ "$(basename "$SCRIPT_DIR")" = "scripts" ]; then
    APP_PATH="$SCRIPT_DIR/.."
else
    APP_PATH="$SCRIPT_DIR"
fi

TASK_NAME="${1:-sandstorm-tracker}"
WRAPPER_SCRIPT="$APP_PATH/run-with-update.sh"
APP_BINARY="$APP_PATH/sandstorm-tracker"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "This script must be run as root (sudo ./setup-service.sh)" >&2
    exit 1
fi

# Validate files exist
if [ ! -f "$APP_BINARY" ]; then
    echo "Error: sandstorm-tracker binary not found at $APP_BINARY" >&2
    exit 1
fi

if [ ! -f "$WRAPPER_SCRIPT" ]; then
    echo "Error: run-with-update.sh not found at $WRAPPER_SCRIPT" >&2
    exit 1
fi

# Make scripts executable
chmod +x "$APP_BINARY" "$WRAPPER_SCRIPT"

echo "Setting up $TASK_NAME service..."
echo "App Path: $APP_PATH"
echo "Wrapper Script: $WRAPPER_SCRIPT"

# Create systemd service file
SERVICE_FILE="/etc/systemd/system/${TASK_NAME}.service"

cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Sandstorm Tracker Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$APP_PATH
ExecStart=$WRAPPER_SCRIPT sandstorm-tracker logs/update-serve.log
Restart=always
RestartSec=60

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sandstorm-tracker

# Security
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

echo "Created systemd service at $SERVICE_FILE"

# Reload systemd daemon
systemctl daemon-reload
echo "Reloaded systemd daemon"

# Enable the service for auto-start
systemctl enable "$TASK_NAME.service"
echo "Enabled service for auto-start"

# Start the service
systemctl start "$TASK_NAME.service"
echo "Started service"

# Check status
sleep 2
if systemctl is-active --quiet "$TASK_NAME.service"; then
    echo "✓ Service is running!"
else
    echo "✗ Service failed to start. Check logs with: journalctl -u $TASK_NAME -n 50"
    exit 1
fi

echo ""
echo "Setup complete!"
echo "The service will:"
echo "  - Start automatically on system boot"
echo "  - Restart automatically if it crashes"
echo "  - Restart every 60 seconds if it exits"
echo ""
echo "Management commands:"
echo "  Status:  sudo systemctl status $TASK_NAME"
echo "  Start:   sudo systemctl start $TASK_NAME"
echo "  Stop:    sudo systemctl stop $TASK_NAME"
echo "  Logs:    sudo journalctl -u $TASK_NAME -f"
echo "  Remove:  sudo systemctl disable $TASK_NAME && sudo rm $SERVICE_FILE && sudo systemctl daemon-reload"
