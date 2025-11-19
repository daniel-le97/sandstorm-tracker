# Windows Task Scheduler Setup Guide

## Overview

This guide sets up automatic updates for Sandstorm Tracker on Windows Server using Task Scheduler.

## How It Works

1. **Task Scheduler** runs the updater binary every hour (configurable)
2. **Updater** reads the PID from `sandstorm-tracker.pid`
3. **Updater** sends graceful shutdown signal to the app
4. **App** shuts down and deletes the PID file
5. **Updater** runs the update command
6. **Updater** restarts the app
7. **App** writes new PID file on startup

## Prerequisites

- Windows Server 2012 R2 or later
- PowerShell 5.0 or later (built-in on Windows Server 2016+)
- Administrator privileges to create Task Scheduler jobs
- `update-restart.exe` built and available (e.g., `C:\Program Files\sandstorm-tracker\update-restart.exe`)
- `sandstorm-tracker.exe` in PATH or specified via `-AppBinary` parameter

## Setup Steps

### Step 1: Verify Paths

Before running setup, ensure you have:

- ✅ Compiled updater binary: `update-restart.exe`
- ✅ App working directory (where `sandstorm-tracker.pid` will be created)
- ✅ App binary in PATH or in the working directory

Example structure:

```
C:\Program Files\sandstorm-tracker\
├── sandstorm-tracker.exe
├── update-restart.exe
├── sandstorm-tracker.yml
└── logs\
    ├── app_output.log
    └── app_error.log
```

### Step 2: Run Setup Script

Run the setup script with the correct paths:

```powershell
# Open PowerShell as Administrator
cd C:\Path\To\scripts
.\setup-update-task.ps1 `
    -UpdaterPath "C:\Program Files\sandstorm-tracker\update-restart.exe" `
    -WorkingDirectory "C:\Program Files\sandstorm-tracker" `
    -IntervalMinutes 60 `
    -TaskName "SandstormTrackerUpdate"
```

**Parameters:**

- `-UpdaterPath`: Full path to `update-restart.exe` binary
- `-WorkingDirectory`: Directory where the app runs (where `sandstorm-tracker.pid` is created)
- `-IntervalMinutes`: How often to check for updates (default: 60)
- `-TaskName`: Name of the Task Scheduler job (default: "SandstormTrackerUpdate")
- `-AppName`: Display name (default: "Sandstorm Tracker")

### Step 3: Verify Task Creation

```powershell
# List the task
Get-ScheduledTask -TaskName "SandstormTrackerUpdate"

# Run it manually (optional, to test)
Start-ScheduledTask -TaskName "SandstormTrackerUpdate"
```

### Step 4: Monitor Execution

Check Windows Event Viewer for task execution logs:

1. Open **Event Viewer** (`eventvwr.msc`)
2. Navigate to **Windows Logs** > **System**
3. Filter by Event ID: `129` (Task Scheduler task started) or `4698` (Task created)
4. Search for "SandstormTracker" in the task name

Check app logs:

```
C:\Program Files\sandstorm-tracker\logs\
├── app_output.log     # Normal operation logs
├── app_error.log      # Error logs
└── app.log            # Full app log
```

## How the PID File Works

### On App Startup

```go
// app.go - OnServe hook
pidData := []byte(fmt.Sprintf("%d", os.Getpid()))
os.WriteFile("sandstorm-tracker.pid", pidData, 0644)
```

Creates file with content: `12345` (just the process ID number)

### On App Shutdown

```go
// app.go - OnTerminate hook
os.Remove("sandstorm-tracker.pid")
```

Deletes the PID file after graceful shutdown

### Updater Reads It

```powershell
# Read PID from file
$pidContent = Get-Content "sandstorm-tracker.pid" -Raw
$pid = [int]$pidContent.Trim()
```

## Troubleshooting

### Task doesn't run automatically

- Check Task Scheduler: Right-click task > Properties > Triggers
- Verify "At Startup" and "Repeat every 60 minutes" are enabled
- Check if "Stop task if it runs longer than" is too short (should be at least 5 min)

### Updater can't find process

- Verify PID file exists: `Test-Path "C:\Program Files\sandstorm-tracker\sandstorm-tracker.pid"`
- Check if app is running: `Get-Process -Name sandstorm-tracker`
- Review updater logs in `logs\` directory

### Update hangs or fails

- Check `logs\app_output.log` and `logs\app_error.log` for errors
- Verify `sandstorm-tracker serve` works manually
- Check disk space and permissions in the working directory

### Task runs but app doesn't restart

- Verify `-AppBinary` parameter matches actual binary name
- Check that app binary is in PATH or specified with full path
- Review Event Viewer logs for exit codes

## Manual Execution

If needed, you can run the updater manually:

```powershell
# Automatic (reads from PID file)
C:\Program Files\sandstorm-tracker\update-restart.exe

# With explicit PID
C:\Program Files\sandstorm-tracker\update-restart.exe -pid 12345

# Custom parameters
C:\Program Files\sandstorm-tracker\update-restart.exe `
    -pidfile "C:\Program Files\sandstorm-tracker\sandstorm-tracker.pid" `
    -timeout 45s `
    -logs "C:\Program Files\sandstorm-tracker\logs" `
    -http "0.0.0.0:8090"
```

## Advanced: Modify Update Frequency

To change how often updates are checked:

```powershell
# Edit the task
$task = Get-ScheduledTask -TaskName "SandstormTrackerUpdate"
$trigger = $task.Triggers[1]  # Get the repeat trigger (index 1)
$trigger.Repetition.Interval = "PT30M"  # Change to 30 minutes (PT = Period Time)
$task | Set-ScheduledTask
```

Valid intervals: `PT15M`, `PT30M`, `PT1H`, `PT2H`, etc. (ISO 8601 format)

## Advanced: Add Additional Actions

To add a pre-update or post-update action:

```powershell
# Stop the task (edit required)
Unregister-ScheduledTask -TaskName "SandstormTrackerUpdate" -Confirm:$false

# Create new task with additional logic
# (Modify setup-update-task.ps1 and rerun)
```

## Related Files

- **Updater Binary**: `tools/update-restart/main.go`
- **Setup Script**: `scripts/setup-update-task.ps1`
- **App Hooks**: `internal/app/app.go` (OnServe, OnTerminate)
- **PID File Location**: `sandstorm-tracker.pid` (in working directory)

## Support

For issues or questions:

1. Check logs in `logs/` directory
2. Review Event Viewer for Task Scheduler errors
3. Run setup script with `-Verbose` flag (if available)
4. Check that all paths and permissions are correct
