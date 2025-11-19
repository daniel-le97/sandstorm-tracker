# Update and Restart Scripts

This directory contains scripts to gracefully update and restart the sandstorm-tracker application.

## Overview

Both scripts perform the same three-step process:

1. **Graceful Shutdown**: Send SIGTERM to the running app (by PID), allowing it to close databases and connections properly
2. **Update**: Run the app's `update` command (powered by the `ghupdate` PocketBase plugin)
3. **Restart**: Start the app with the updated binary

The scripts are designed to be called **from your main app**, passing the app's PID so it can gracefully shut itself down.

## PowerShell Script

### File

`update-and-restart.ps1`

### Usage

```powershell
# Pass the PID of your app
.\update-and-restart.ps1 -PID 12345

# With custom timeout
.\update-and-restart.ps1 -PID 12345 -ShutdownTimeout 60
```

### Parameters

- **`-PID`** (required)
  The process ID of the app to shutdown and restart. Typically passed as `[System.Diagnostics.Process]::GetCurrentProcess().Id` from your app.

- **`-ShutdownTimeout`** (default: `30`)
  Grace period in seconds before force-killing the process if it doesn't exit gracefully

- **`-AppBinary`** (default: `sandstorm-tracker`)
  The name of the app binary to restart

### Features

- ✅ Graceful process shutdown using `CloseMainWindow()`
- ✅ Force kill if graceful shutdown takes too long
- ✅ Logs output to `logs/app_output.log` and `logs/app_error.log`
- ✅ Detailed timestamped logging to console
- ✅ Error handling with exit codes
- ✅ Windows-native PowerShell (no external dependencies)
- ✅ **Accepts PID parameter** - called from your app

### Requirements

- PowerShell 5.0+
- The app binary must be in the current directory or in PATH
- PocketBase `update` command must be available

### Example Output

```
[2025-11-18 15:30:45] [INFO] Starting update and restart procedure
[2025-11-18 15:30:45] [INFO] Looking for running sandstorm-tracker process...
[2025-11-18 15:30:45] [INFO] Found process 12345. Sending graceful shutdown signal...
[2025-11-18 15:30:46] [INFO] Process exited gracefully (exit code: 0)
[2025-11-18 15:30:46] [INFO] Running update command...
[2025-11-18 15:30:48] [INFO] Update completed successfully
[2025-11-18 15:30:49] [INFO] Starting application...
[2025-11-18 15:30:49] [INFO] Application started with process ID 12356
[2025-11-18 15:30:51] [INFO] Application is running successfully
[2025-11-18 15:30:51] [INFO] Update and restart procedure completed successfully
```

## Go Binary

### Files

- `tools/update-restart/main.go` - Source code
- `update-restart` - Compiled binary (after building)

### Build

```bash
go build -o update-restart ./tools/update-restart
```

Or from the root directory:

```bash
go build -o ./tools/update-restart/update-restart ./tools/update-restart
```

### Usage

```bash
# Pass the PID of your app
./update-restart -pid 12345

# With custom timeout
./update-restart -pid 12345 -timeout 60s

# Full parameters
./update-restart -pid 12345 -timeout 30s -app sandstorm-tracker -http 0.0.0.0:8090 -logs logs
```

### Parameters

- **`-pid`** (required)
  The process ID of the app to shutdown and restart. Typically `os.Getpid()` from your app.

- **`-timeout`** (default: `30s`)
  Grace period before force kill (Go duration format: `30s`, `1m`, etc.)

- **`-app`** (default: `sandstorm-tracker`)
  Name of the app binary to restart

- **`-logs`** (default: `logs`)
  Directory for output logs

- **`-http`** (default: `0.0.0.0:8090`)
  HTTP address for the app to listen on

### Features

- ✅ Cross-platform (Windows, Linux, macOS)
- ✅ Graceful shutdown with SIGTERM
- ✅ Configurable timeouts
- ✅ Uses `os.FindProcess()` - reliable PID-based shutdown
- ✅ No process lookup needed - **PID is passed as parameter**

### Limitations

None! By passing the PID as a parameter, we avoid any process lookup complexity.

## Recommended Setup: Calling from Your App

```powershell
# PowerShell (as Administrator)
$action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-ExecutionPolicy Bypass -File C:\path\to\update-and-restart.ps1"
$trigger = New-ScheduledTaskTrigger -Daily -At 3:00AM
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "SandstormTrackerUpdate" -Description "Update and restart sandstorm-tracker daily"
```

### Calling from Your Go App

The ideal way to use these scripts is from your PocketBase app. Here's an example:

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func TriggerUpdate() error {
	// On Windows, use PowerShell script
	cmd := exec.Command(
		"powershell.exe",
		"-ExecutionPolicy", "Bypass",
		"-File", "./scripts/update-and-restart.ps1",
		"-PID", fmt.Sprintf("%d", os.Getpid()),
	)

	// Or on Linux/macOS, use Go binary
	// cmd := exec.Command(
	//   "./tools/update-restart/update-restart",
	//   "-pid", fmt.Sprintf("%d", os.Getpid()),
	// )

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
```

You can call this from an HTTP endpoint, scheduled job, or anywhere in your app when an update is available.

### Trigger from ghupdate Hook

If the `ghupdate` plugin has an "update available" hook, you can trigger the script automatically:

```go
app.OnRecordBeforeCreateRequest("updates").BindFunc(func(e *core.RecordRequestEvent) error {
	// Update available - trigger restart
	return TriggerUpdate()
})
```

## Integration with ghupdate Plugin

The scripts assume your PocketBase app has the `ghupdate` plugin registered (which it does). The plugin provides the `update` command:

```bash
./sandstorm-tracker update
```

This command:

1. Checks for new releases on GitHub
2. Downloads the new binary
3. Replaces the current binary
4. Exits (the script then restarts it)

## Troubleshooting

### Process doesn't exit gracefully

- Increase `ShutdownTimeout` (default 30 seconds)
- Check if the app has long-running operations that should complete first
- Review `logs/app_error.log` for any issues

### Update command fails

- Verify the app can reach GitHub (firewall/proxy issues?)
- Check release assets are available on the GitHub repository
- Review `logs/app_error.log` for specific error messages

### App crashes immediately after restart

- Check `logs/app_output.log` and `logs/app_error.log`
- Verify the binary was actually updated
- Ensure all required configuration files are present

### PocketBase error: "database locked"

This can happen if the update process doesn't wait long enough after shutdown. Try increasing the delay between shutdown and update in the script.

## Security Considerations

- The scripts assume the app binary is trusted
- `ghupdate` plugin downloads from GitHub - ensure your repository is secure
- Consider verifying binary signatures if deploying to production
- Run scripts with appropriate privileges (service account, not root)
