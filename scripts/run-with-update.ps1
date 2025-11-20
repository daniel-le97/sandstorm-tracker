# run-with-update.ps1
# Checks for updates, applies them, then starts the server with logging

param(
    [string]$AppName = "sandstorm-tracker",
    [string]$LogFile = "logs\update-serve.log"
)

# Get the directory where this script is located (scripts folder)
# Then go up one level to the app directory
$scriptDir = Split-Path -Parent $PSCommandPath
$appPath = Join-Path (Split-Path -Parent $scriptDir) "$AppName.exe"
$logDir = Join-Path (Split-Path -Parent $scriptDir) "logs"

# Ensure logs directory exists
if (-not (Test-Path $logDir)) {
    New-Item -ItemType Directory -Path $logDir -Force | Out-Null
}

$logPath = Join-Path $logDir $LogFile

# Helper function to log messages
function Write-Log {
    param(
        [string]$Message,
        [ValidateSet("INFO", "WARNING", "ERROR")]
        [string]$Level = "INFO"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "[$timestamp] [$Level] $Message"
    
    Write-Host $logEntry
    Add-Content -Path $logPath -Value $logEntry -Encoding UTF8
}

# Check if app exists
if (-not (Test-Path $appPath)) {
    Write-Log "Error: $AppName.exe not found at $appPath" "ERROR"
    exit 1
}

Write-Log "Starting $AppName with update check..." "INFO"

# Try to update
Write-Log "Checking for updates and applying if available..." "INFO"
& $appPath update 2>&1 | ForEach-Object {
    Write-Host $_
    Add-Content -Path $logPath -Value $_ -Encoding UTF8
}

if ($LASTEXITCODE -ne 0) {
    Write-Log "Update check completed (no updates or update skipped)" "INFO"
}
else {
    Write-Log "Update applied successfully" "INFO"
}

# Start the server
Write-Log "Starting server..." "INFO"
& $appPath serve 2>&1 | ForEach-Object {
    # Remove problematic Unicode characters
    $clean = $_ -replace '[^\x20-\x7E\n\r\t]', ''
    if ($clean -ne '') {
        Write-Host $_
        Add-Content -Path $logPath -Value $_ -Encoding UTF8
    }
}

$exitCode = $LASTEXITCODE
Write-Log "Server stopped with exit code: $exitCode" "WARNING"

# Wait before exiting so Task Scheduler can detect the exit
Start-Sleep -Seconds 2

exit $exitCode
