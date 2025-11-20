# setup-service.ps1
# Sets up sandstorm-tracker as a Windows Task Scheduler service
# Run as Administrator from the sandstorm-tracker directory
#
# Usage:
#   cd C:\opt\sandstorm-tracker
#   .\scripts\setup-service.ps1
#
# Or with custom task name:
#   .\scripts\setup-service.ps1 -TaskName "SandstormTrackerMain"

param(
    [string]$TaskName = "SandstormTracker"
)

# Get the app directory (parent of scripts directory)
$scriptDir = Split-Path -Parent $PSCommandPath
$AppPath = Split-Path -Parent $scriptDir

# Check if running as administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
if (-not $isAdmin) {
    Write-Error "This script must be run as Administrator"
    exit 1
}

# Validate app path
if (-not (Test-Path "$AppPath\sandstorm-tracker.exe")) {
    Write-Error "sandstorm-tracker.exe not found at $AppPath"
    exit 1
}

$scriptPath = Join-Path $AppPath "scripts\run-with-update.ps1"
$exePath = Join-Path $AppPath "sandstorm-tracker.exe"

Write-Host "Setting up $TaskName service..." -ForegroundColor Cyan
Write-Host "App Path: $AppPath"
Write-Host "Script: $scriptPath"

# Verify run-with-update.ps1 exists
if (-not (Test-Path $scriptPath)) {
    Write-Error "run-with-update.ps1 not found at $scriptPath"
    exit 1
}

# Remove existing task if it exists
$existingTask = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
if ($existingTask) {
    Write-Host "Removing existing task '$TaskName'..." -ForegroundColor Yellow
    Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
    Start-Sleep -Seconds 1
}

# Create task action
$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`"" `
    -WorkingDirectory "$AppPath"

# Create task trigger - At startup
$triggerAtStartup = New-ScheduledTaskTrigger -AtStartup

# Create task settings
$settings = New-ScheduledTaskSettingsSet `
    -MultipleInstances Queue `
    -StartWhenAvailable `
    -RestartCount 5 `
    -RestartInterval (New-TimeSpan -Minutes 5) `
    -RunOnlyIfNetworkAvailable

# Create the task
$principal = New-ScheduledTaskPrincipal `
    -UserId "NT AUTHORITY\SYSTEM" `
    -LogonType ServiceAccount `
    -RunLevel Highest

Write-Host "Creating scheduled task '$TaskName'..." -ForegroundColor Yellow
Register-ScheduledTask `
    -TaskName $TaskName `
    -Description "Runs sandstorm-tracker server with automatic restart on exit" `
    -Action $action `
    -Trigger $triggerAtStartup `
    -Settings $settings `
    -Principal $principal

Write-Host "Task created successfully!" -ForegroundColor Green

# Start the task
Write-Host "Starting task..." -ForegroundColor Yellow
Start-ScheduledTask -TaskName $TaskName

# Verify it's running
Start-Sleep -Seconds 2
$taskInfo = Get-ScheduledTaskInfo -TaskName $TaskName
if ($taskInfo.LastTaskResult -eq 0 -or $taskInfo.LastRunTime) {
    Write-Host "Task is running!" -ForegroundColor Green
}
else {
    Write-Host "Task may not have started yet. Check Event Viewer." -ForegroundColor Yellow
}

Write-Host "`nSetup complete!" -ForegroundColor Green
Write-Host "The service will:"
Write-Host "  - Start automatically on Windows startup"
Write-Host "  - Restart automatically if it crashes"
Write-Host "  - Retry up to 5 times with 5-minute intervals between restarts"
Write-Host "`nManagement commands:"
Write-Host "  Start:   Start-ScheduledTask -TaskName $TaskName"
Write-Host "  Stop:    Stop-ScheduledTask -TaskName $TaskName"
Write-Host "  Status:  Get-ScheduledTask -TaskName $TaskName"
Write-Host "  Remove:  Unregister-ScheduledTask -TaskName $TaskName -Confirm:`$false"
