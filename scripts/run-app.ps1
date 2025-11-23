# run-app.ps1
# Runs sandstorm-tracker and restarts it if it exits
# Supports execution from ./scripts/run-app.ps1 or ./run-app.ps1

# Detect the project root directory
# If script is in 'scripts' subdirectory, go up one level
# If script is in project root, use current directory
$scriptDir = Split-Path -Parent $PSCommandPath
$scriptName = Split-Path -Leaf $PSCommandPath

if ($scriptDir -like "*\scripts") {
    # Script is in ./scripts directory, go up one level to project root
    $appPath = Split-Path -Parent $scriptDir
} else {
    # Script is in project root directory
    $appPath = $scriptDir
}

$exePath = Join-Path $appPath "sandstorm-tracker.exe"
$logPath = Join-Path $appPath "logs\service.log"

# Ensure logs directory exists
$logsDir = Split-Path -Parent $logPath
if (-not (Test-Path $logsDir)) {
    New-Item -ItemType Directory -Path $logsDir -Force | Out-Null
}

# Verify the executable exists
if (-not (Test-Path $exePath)) {
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $errorMsg = "$timestamp - ERROR: sandstorm-tracker.exe not found at $exePath"
    Add-Content -Path $logPath -Value $errorMsg
    Write-Error $errorMsg
    exit 1
}

# Log startup
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Add-Content -Path $logPath -Value "$timestamp - Starting sandstorm-tracker from $exePath"
Write-Host "Starting sandstorm-tracker from $exePath"

# Run the app in serve mode
& $exePath serve 2>&1 | Tee-Object -FilePath $logPath -Append

# Log exit
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Add-Content -Path $logPath -Value "$timestamp - sandstorm-tracker exited with code $LASTEXITCODE"

# Wait a moment before exiting so Task Scheduler can restart
Start-Sleep -Seconds 2
