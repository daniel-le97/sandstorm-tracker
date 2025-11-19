# run-app.ps1
# Runs sandstorm-tracker and restarts it if it exits

$appPath = Split-Path -Parent $PSCommandPath | Split-Path -Parent
$exePath = Join-Path $appPath "sandstorm-tracker.exe"
$logPath = Join-Path $appPath "logs\service.log"

# Ensure logs directory exists
$logsDir = Split-Path -Parent $logPath
if (-not (Test-Path $logsDir)) {
    New-Item -ItemType Directory -Path $logsDir -Force | Out-Null
}

# Log startup
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Add-Content -Path $logPath -Value "$timestamp - Starting sandstorm-tracker from $exePath"

# Run the app in serve mode
& $exePath serve 2>&1 | Tee-Object -FilePath $logPath -Append

# Log exit
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Add-Content -Path $logPath -Value "$timestamp - sandstorm-tracker exited with code $LASTEXITCODE"

# Wait a moment before exiting so Task Scheduler can restart
Start-Sleep -Seconds 2
