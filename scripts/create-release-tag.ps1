# create-release-tag.ps1
# Automatically creates the next semantic version tag and pushes it to GitHub

param(
    [ValidateSet("major", "minor", "patch")]
    [string]$BumpType = "patch"
)

# Get the latest tag
$latestTag = git describe --tags --abbrev=0 2>$null
if (-not $latestTag) {
    Write-Host "No tags found. Starting with v0.0.1" -ForegroundColor Yellow
    $newTag = "v0.0.1"
} else {
    Write-Host "Latest tag: $latestTag" -ForegroundColor Cyan
    
    # Parse version (remove 'v' prefix)
    $version = $latestTag -replace '^v'
    $parts = $version -split '\.'
    
    if ($parts.Count -ne 3) {
        Write-Error "Invalid version format: $latestTag (expected v#.#.#)"
        exit 1
    }
    
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2]
    
    # Bump version
    switch ($BumpType) {
        "major" {
            $major++
            $minor = 0
            $patch = 0
        }
        "minor" {
            $minor++
            $patch = 0
        }
        "patch" {
            $patch++
        }
    }
    
    $newTag = "v$major.$minor.$patch"
}

Write-Host "Creating new tag: $newTag" -ForegroundColor Green

# Create tag
git tag $newTag
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to create tag"
    exit 1
}

# Push tag
git push origin $newTag
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to push tag. Deleting local tag..."
    git tag -d $newTag
    exit 1
}

Write-Host "`nTag created and pushed successfully!" -ForegroundColor Green
Write-Host "Next step: Run goreleaser"
Write-Host "  `$env:GITHUB_TOKEN='your-token'"
Write-Host "  goreleaser release --clean --skip=docker"
