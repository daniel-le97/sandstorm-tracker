# Kill sandstorm-tracker.exe processes
# This script terminates any running sandstorm-tracker instances

$processName = "sandstorm-tracker"
$processes = Get-Process -Name $processName -ErrorAction SilentlyContinue

if ($processes) {
    Write-Host "Found $($processes.Count) sandstorm-tracker process(es). Terminating..." -ForegroundColor Yellow

    foreach ($process in $processes) {
        Write-Host "  Killing process ID: $($process.Id)"
        Stop-Process -Id $process.Id -Force
    }

    Write-Host "All sandstorm-tracker processes have been terminated" -ForegroundColor Green
} else {
    Write-Host "No sandstorm-tracker processes found running" -ForegroundColor Gray
}
