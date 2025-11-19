# Setup Windows Task Scheduler for Sandstorm Tracker Updates
# This script creates a scheduled task that checks for updates every hour
# Usage: .\setup-update-task.ps1

param(
    [string]$UpdaterPath = "C:\path\to\update-restart.exe",
    [string]$AppName = "Sandstorm Tracker",
    [string]$TaskName = "SandstormTrackerUpdate",
    [string]$IntervalMinutes = 60,
    [string]$WorkingDirectory = "C:\path\to\app"
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
    Write-Log "Setting up Task Scheduler for $AppName"
    
    # Check if updater exists
    if (-not (Test-Path $UpdaterPath)) {
        Write-Log "ERROR: Updater not found at $UpdaterPath" "ERROR"
        exit 1
    }
    
    # Check if working directory exists
    if (-not (Test-Path $WorkingDirectory)) {
        Write-Log "ERROR: Working directory not found at $WorkingDirectory" "ERROR"
        exit 1
    }
    
    Write-Log "Updater path: $UpdaterPath"
    Write-Log "Working directory: $WorkingDirectory"
    Write-Log "Task will run every $IntervalMinutes minutes"
    
    # Create task action (what the task will run)
    $action = New-ScheduledTaskAction `
        -Execute $UpdaterPath `
        -WorkingDirectory $WorkingDirectory `
        -ErrorAction Stop
    
    Write-Log "Created task action"
    
    # Create trigger (when the task will run)
    # Runs at system startup and repeats every X minutes
    $trigger = @(
        (New-ScheduledTaskTrigger -AtStartup -ErrorAction Stop),
        (New-ScheduledTaskTrigger -RepetitionInterval (New-TimeSpan -Minutes $IntervalMinutes) -RepetitionDuration (New-TimeSpan -Days 1000) -ErrorAction Stop)
    )
    
    Write-Log "Created task triggers (at startup + every $IntervalMinutes minutes)"
    
    # Create settings
    $settings = New-ScheduledTaskSettingsSet `
        -AllowStartIfOnBatteries `
        -DontStopIfGoingOnBatteries `
        -StartWhenAvailable `
        -RunOnlyIfNetworkAvailable `
        -ErrorAction Stop
    
    Write-Log "Created task settings"
    
    # Check if task already exists
    $existingTask = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
    if ($existingTask) {
        Write-Log "Task '$TaskName' already exists. Removing old task..." "WARN"
        Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction Stop
        Start-Sleep -Seconds 1
    }
    
    # Register the task
    $task = Register-ScheduledTask `
        -TaskName $TaskName `
        -Action $action `
        -Trigger $trigger `
        -Settings $settings `
        -Description "Automatically checks for updates and restarts $AppName if available" `
        -RunLevel Highest `
        -ErrorAction Stop
    
    Write-Log "Successfully registered Task Scheduler job: $TaskName" "SUCCESS"
    Write-Log "Task Details:"
    Write-Log "  Name: $($task.TaskName)"
    Write-Log "  Path: $($task.TaskPath)"
    Write-Log "  State: $($task.State)"
    Write-Log ""
    Write-Log "Next scheduled run: Check Task Scheduler for details"
    Write-Log ""
    Write-Log "To verify: Open Task Scheduler > Task Scheduler Library > find '$TaskName'"
    Write-Log "To run manually: schtasks /run /tn $TaskName"
    Write-Log "To view logs: Event Viewer > Windows Logs > System (search for 'SandstormTracker')"
    
}
catch {
    Write-Log "ERROR: Failed to create task scheduler job: $_" "ERROR"
    exit 1
}
