# Kill all Insurgency server processes (not the game client)

$procs = Get-Process -Name '*InsurgencyServer*' -ErrorAction SilentlyContinue

if ($procs) {
    Write-Host "Found $($procs.Count) Insurgency server process(es):" -ForegroundColor Yellow
    $procs | ForEach-Object {
        Write-Host "  - PID $($_.Id): $($_.ProcessName)" -ForegroundColor Cyan
    }
    
    Write-Host "Stopping all processes..." -ForegroundColor Yellow
    $procs | Stop-Process -Force
    Write-Host "All processes killed." -ForegroundColor Green
}
else {
    Write-Host "No Insurgency server processes found." -ForegroundColor Gray
}
