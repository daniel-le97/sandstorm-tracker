# run-with-update.ps1
# Wrapper script to check for updates, then start the app
# Supports execution from ./scripts/run-with-update.ps1 or ./run-with-update.ps1
# All app logging goes to app.log and the database

param(
    [string]$AppName = "sandstorm-tracker"
)

# Detect the project root directory
# If script is in 'scripts' subdirectory, go up one level
# If script is in project root, use current directory
$scriptDir = Split-Path -Parent $PSCommandPath
$scriptName = Split-Path -Leaf $PSCommandPath

if ($scriptDir -like "*\scripts") {
    # Script is in ./scripts directory, go up one level to project root
    $appDir = Split-Path -Parent $scriptDir
} else {
    # Script is in project root directory
    $appDir = $scriptDir
}

$appPath = Join-Path $appDir "$AppName.exe"

# Check if app exists
if (-not (Test-Path $appPath)) {
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $errorMsg = "$timestamp - ERROR: $AppName.exe not found at $appPath"
    Write-Error $errorMsg
    exit 1
}

Write-Host "App directory: $appDir"
Write-Host "Checking for updates..."
& $appPath update

if ($LASTEXITCODE -ne 0) {
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "$timestamp - Update check exited with code $LASTEXITCODE"
}

Write-Host "Starting server..."
& $appPath serve

exit $LASTEXITCODE
