# run-with-update.ps1
# Wrapper script to check for updates, then start the app
# All app logging goes to app.log and the database

param(
    [string]$AppName = "sandstorm-tracker"
)

# Get app directory (parent of scripts folder)
$scriptDir = Split-Path -Parent $PSCommandPath
$appDir = Split-Path -Parent $scriptDir
$appPath = Join-Path $appDir "$AppName.exe"

# Check if app exists
if (-not (Test-Path $appPath)) {
    Write-Error "Error: $AppName.exe not found at $appPath"
    exit 1
}

Write-Host "Checking for updates..."
& $appPath update

Write-Host "Starting server..."
& $appPath serve

exit $LASTEXITCODE
