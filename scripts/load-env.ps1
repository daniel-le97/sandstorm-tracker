Get-Content .env | ForEach-Object { 
    if ($_ -and -not $_.StartsWith('#')) {
        $parts = $_ -split '=', 2
        [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1], 'Process')
    }
}