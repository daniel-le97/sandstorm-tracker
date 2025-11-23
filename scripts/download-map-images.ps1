# Download all map images from maps.json to assets/static

$projectRoot = Split-Path -Parent $PSScriptRoot
$mapsJsonPath = "$projectRoot\docs\maps.json"
$outputDir = "$projectRoot\assets\static"

# Ensure output directory exists
if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    Write-Host "Created directory: $outputDir"
}

# Read maps.json
$mapsData = Get-Content $mapsJsonPath | ConvertFrom-Json

# Download each image
$maps = $mapsData.maps
$count = 0

foreach ($map in $maps) {
    $imageUrl = $map.imageUrl
    $fileName = $map.fileName
    $outputPath = Join-Path $outputDir $fileName
    
    try {
        Write-Host "Downloading $fileName..." -ForegroundColor Cyan
        Invoke-WebRequest -Uri $imageUrl -OutFile $outputPath
        $count++
        Write-Host "Downloaded: $fileName" -ForegroundColor Green
    }
    catch {
        Write-Host "Failed to download $fileName - $_" -ForegroundColor Red
    }
}

Write-Host "`nDownload complete! Downloaded $count/$($maps.Count) images to $outputDir" -ForegroundColor Cyan
