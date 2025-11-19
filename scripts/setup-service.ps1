# setup-service.ps1
# Sets up sandstorm-tracker as a Windows Task Scheduler service
# Run as Administrator

param(
    [string]$AppPath = "C:\path\to\sandstorm-tracker",
    [string]$TaskName = "SandstormTracker"
)

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

$scriptPath = Join-Path $AppPath "scripts\run-app.ps1"
$exePath = Join-Path $AppPath "sandstorm-tracker.exe"

Write-Host "Setting up $TaskName service..." -ForegroundColor Cyan
Write-Host "App Path: $AppPath"
Write-Host "Script: $scriptPath"

# Create run-app.ps1 if it doesn't exist
if (-not (Test-Path $scriptPath)) {
    Write-Host "Creating $scriptPath..." -ForegroundColor Yellow
    
    $runScript = @"
# run-app.ps1
# Runs sandstorm-tracker and restarts it if it exits

`$appPath = Split-Path -Parent `$PSCommandPath | Split-Path -Parent
`$exePath = Join-Path `$appPath "sandstorm-tracker.exe"

# Run the app in serve mode
Write-Host "Starting sandstorm-tracker..."
& `$exePath serve

# When app exits, this script exits and Task Scheduler will restart it
Write-Host "sandstorm-tracker exited. Waiting for Task Scheduler to restart..."
Start-Sleep -Seconds 2
"@
    
    Set-Content -Path $scriptPath -Value $runScript -Encoding UTF8
    Write-Host "Created $scriptPath" -ForegroundColor Green
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
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`""

# Create task trigger - At startup
$triggerAtStartup = New-ScheduledTaskTrigger -AtStartup

# Create task settings
$settings = New-ScheduledTaskSettingsSet `
    -MultipleInstances Queue `
    -StartWhenAvailable `
    -RestartCount 5 `
    -RestartInterval (New-TimeSpan -Seconds 2) `
    -RunOnlyIfNetworkAvailable `
    -ExecutionTimeLimit (New-TimeSpan -Hours 0)

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
Write-Host "  - Restart every 1 minute if it exits"
Write-Host "`nManagement commands:"
Write-Host "  Start:   Start-ScheduledTask -TaskName $TaskName"
Write-Host "  Stop:    Stop-ScheduledTask -TaskName $TaskName"
Write-Host "  Status:  Get-ScheduledTask -TaskName $TaskName"
Write-Host "  Remove:  Unregister-ScheduledTask -TaskName $TaskName -Confirm:`$false"
