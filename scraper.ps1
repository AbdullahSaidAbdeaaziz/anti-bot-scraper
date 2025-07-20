# Anti-Bot Scraper PowerShell Wrapper
param(
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Arguments
)

$scraperPath = Join-Path $PSScriptRoot "bin\scraper.exe"

if (-not (Test-Path $scraperPath)) {
    Write-Error "Scraper executable not found at: $scraperPath"
    Write-Host "Please run: go build -o bin/scraper.exe ./cmd/scraper"
    exit 1
}

if ($Arguments.Count -eq 0 -or $Arguments -contains "-h" -or $Arguments -contains "--help" -or $Arguments -contains "help") {
    Write-Host "Anti-Bot Scraper PowerShell Wrapper" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage: .\scraper.ps1 [options]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Cyan
    Write-Host "  .\scraper.ps1 -url https://httpbin.org/headers"
    Write-Host "  .\scraper.ps1 -url https://httpbin.org/headers -browser firefox -verbose"
    Write-Host "  .\scraper.ps1 -url https://httpbin.org/post -method POST -data '@test_data.json'"
    Write-Host ""
    Write-Host "For full help:" -ForegroundColor Yellow
    & $scraperPath -help
    exit 0
}

# Execute the scraper with all arguments
try {
    & $scraperPath @Arguments
} catch {
    Write-Error "Failed to execute scraper: $_"
    exit 1
}
