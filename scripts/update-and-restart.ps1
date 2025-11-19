# Update and Restart Script for Sandstorm Tracker
# Gracefully shuts down the app, runs the update command, and restarts it
# Usage: .\.\update-and-restart.ps1 -PID <process-id> [-ShutdownTimeout <seconds>] [-AppBinary <binary-name>]

param(
    [Parameter(Mandatory = $true)]
    [int]$PID,
    [int]$ShutdownTimeout = 30,
    [string]$AppBinary = "sandstorm-tracker"
)

# Enable strict mode
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Logging function
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message"
}

try {
    Write-Log "Starting update and restart procedure"
    Write-Log "Target PID: $PID"
    
    # Get the process by PID
    Write-Log "Looking for process with PID $PID..."
    $process = Get-Process -ID $PID -ErrorAction SilentlyContinue
    
    if ($null -eq $process) {
        Write-Log "Process with PID $PID not found. Skipping graceful shutdown." "WARN"
    }
    else {
        Write-Log "Found process $($process.Id). Sending graceful shutdown signal..."
        
        # Send SIGTERM (Ctrl+C) to allow graceful shutdown
        # PocketBase respects shutdown signals and closes databases/connections
        $process.CloseMainWindow()
        
        # Wait for process to exit
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        while (!$process.HasExited -and $stopwatch.Elapsed.TotalSeconds -lt $ShutdownTimeout) {
            Start-Sleep -Milliseconds 100
        }
        
        if (!$process.HasExited) {
            Write-Log "Process did not exit gracefully after $ShutdownTimeout seconds. Forcing termination." "WARN"
            Stop-Process -Id $process.Id -Force
            Start-Sleep -Milliseconds 500
        }
        else {
            Write-Log "Process exited gracefully (exit code: $($process.ExitCode))"
        }
    }
    
    Write-Log "Running update command..."
    # Run the app's update command (ghupdate plugin will handle the update)
    & ".\$AppBinary" update
    if ($LASTEXITCODE -ne 0) {
        Write-Log "Update command failed with exit code $LASTEXITCODE" "ERROR"
        exit $LASTEXITCODE
    }
    Write-Log "Update completed successfully"
    
    # Small delay to ensure file system is ready
    Start-Sleep -Seconds 1
    
    Write-Log "Starting application..."
    # Start the app in the background
    $newProcess = Start-Process -FilePath ".\$AppBinary" `
        -ArgumentList "serve", "--http=0.0.0.0:8090" `
        -PassThru `
        -NoNewWindow `
        -RedirectStandardOutput "logs/app_output.log" `
        -RedirectStandardError "logs/app_error.log"
    
    Write-Log "Application started with process ID $($newProcess.Id)"
    
    # Wait a moment and check if process is still running
    Start-Sleep -Seconds 2
    if ($newProcess.HasExited) {
        Write-Log "Application exited immediately after startup. Check logs for errors:" "ERROR"
        if (Test-Path "logs/app_error.log") {
            Get-Content "logs/app_error.log" | Write-Host
        }
        exit 1
    }
    
    Write-Log "Application is running successfully"
    
}
catch {
    Write-Log "Error during update and restart: $($_.Exception.Message)" "ERROR"
    Write-Log "Stack trace: $($_.ScriptStackTrace)" "ERROR"
    exit 1
}

Write-Log "Update and restart procedure completed successfully"
