# Anti-Bot Scraper - Comprehensive Demo Script (PowerShell)
# This script demonstrates all the key features of the scraper

$SCRAPER = ".\bin\scraper.exe"
$DEMO_DIR = ".\demo"

Write-Host "🚀 Anti-Bot TLS Fingerprint Scraper - Final Demo" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Green
Write-Host ""

# Check if scraper is built
if (-not (Test-Path $SCRAPER)) {
    Write-Host "❌ Scraper not found. Building..." -ForegroundColor Red
    go build -o bin/scraper.exe ./cmd/scraper
    Write-Host "✅ Scraper built successfully" -ForegroundColor Green
    Write-Host ""
}

Write-Host "📋 Demo 1: Basic Browser Fingerprint Testing" -ForegroundColor Cyan
Write-Host "--------------------------------------------" -ForegroundColor Cyan
$browsers = @("chrome", "firefox", "safari", "edge")

foreach ($browser in $browsers) {
    Write-Host "🌐 Testing $browser fingerprint..." -ForegroundColor Yellow
    $result = & $SCRAPER -url "https://httpbin.org/headers" -browser $browser -output json | ConvertFrom-Json
    $headers = $result.body | ConvertFrom-Json
    Write-Host "   User-Agent: $($headers.headers.'User-Agent')" -ForegroundColor White
    Write-Host ""
}

Write-Host "📋 Demo 2: Custom Headers and API Testing" -ForegroundColor Cyan
Write-Host "-----------------------------------------" -ForegroundColor Cyan
Write-Host "🔑 Testing with API headers..." -ForegroundColor Yellow
$apiResult = & $SCRAPER -url "https://httpbin.org/headers" -browser chrome -headers "@$DEMO_DIR/api_headers.json" -output json | ConvertFrom-Json
$apiHeaders = ($apiResult.body | ConvertFrom-Json).headers
Write-Host "   Authorization: $($apiHeaders.Authorization)" -ForegroundColor White
Write-Host "   X-API-Key: $($apiHeaders.'X-Api-Key')" -ForegroundColor White
Write-Host ""

Write-Host "📋 Demo 3: POST Request with Form Data" -ForegroundColor Cyan
Write-Host "--------------------------------------" -ForegroundColor Cyan
Write-Host "📝 Sending login form data..." -ForegroundColor Yellow
& $SCRAPER -url "https://httpbin.org/post" -method POST -browser firefox -data "@$DEMO_DIR/login_data.json" -verbose | Select-Object -First 15
Write-Host ""

Write-Host "📋 Demo 4: Stealth Mode with Custom Headers" -ForegroundColor Cyan
Write-Host "-------------------------------------------" -ForegroundColor Cyan
Write-Host "🥷 Testing stealth headers..." -ForegroundColor Yellow
$stealthResult = & $SCRAPER -url "https://httpbin.org/headers" -browser safari -headers "@$DEMO_DIR/stealth_headers.json" -user-agent "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)" -output json | ConvertFrom-Json
$stealthHeaders = ($stealthResult.body | ConvertFrom-Json).headers
Write-Host "   Referer: $($stealthHeaders.Referer)" -ForegroundColor White
Write-Host "   X-Forwarded-For: $($stealthHeaders.'X-Forwarded-For')" -ForegroundColor White
Write-Host ""

Write-Host "📋 Demo 5: Rate Limiting and Retry Logic" -ForegroundColor Cyan
Write-Host "----------------------------------------" -ForegroundColor Cyan
Write-Host "⏱️  Testing rate limiting with multiple requests..." -ForegroundColor Yellow
for ($i = 1; $i -le 3; $i++) {
    Write-Host "Request $i:" -ForegroundColor White
    $start = Get-Date
    & $SCRAPER -url "https://httpbin.org/delay/1" -browser chrome -rate-limit 2s -timeout 10s -verbose 2>&1 | Where-Object { $_ -match "(Making GET|Status:|took)" }
    $elapsed = (Get-Date) - $start
    Write-Host "   Total time: $($elapsed.TotalSeconds) seconds" -ForegroundColor Gray
}
Write-Host ""

Write-Host "📋 Demo 6: Error Handling and Retries" -ForegroundColor Cyan
Write-Host "-------------------------------------" -ForegroundColor Cyan
Write-Host "🔄 Testing retry logic with failing endpoint..." -ForegroundColor Yellow
& $SCRAPER -url "https://httpbin.org/status/503" -browser edge -retries 3 -rate-limit 1s -verbose 2>&1 | Where-Object { $_ -match "(Making GET|Error:|failed after)" }
Write-Host ""

Write-Host "📋 Demo 7: Cookie Session Management" -ForegroundColor Cyan
Write-Host "------------------------------------" -ForegroundColor Cyan
Write-Host "🍪 Testing cookie persistence..." -ForegroundColor Yellow
Write-Host "Setting a cookie:" -ForegroundColor White
$cookieSetResult = & $SCRAPER -url "https://httpbin.org/cookies/set?demo=session123&user=testuser" -browser chrome -output json | ConvertFrom-Json
Write-Host "   Response: Cookie set successfully" -ForegroundColor Gray

Write-Host ""
Write-Host "Reading cookies (should show the set cookies):" -ForegroundColor White
$cookieGetResult = & $SCRAPER -url "https://httpbin.org/cookies" -browser chrome -output json | ConvertFrom-Json
$cookies = ($cookieGetResult.body | ConvertFrom-Json).cookies
if ($cookies) {
    $cookies.PSObject.Properties | ForEach-Object {
        Write-Host "   $($_.Name): $($_.Value)" -ForegroundColor Gray
    }
} else {
    Write-Host "   No cookies found" -ForegroundColor Gray
}
Write-Host ""

Write-Host "🎉 Demo Complete!" -ForegroundColor Green
Write-Host "=================" -ForegroundColor Green
Write-Host ""
Write-Host "💡 Pro Tips:" -ForegroundColor Yellow
Write-Host "- Use different browser fingerprints for different targets" -ForegroundColor White
Write-Host "- Combine with custom headers for maximum stealth" -ForegroundColor White
Write-Host "- Adjust rate limiting based on target's rate limits" -ForegroundColor White
Write-Host "- Use POST requests for form submissions and API calls" -ForegroundColor White
Write-Host "- Leverage cookie management for session-based scraping" -ForegroundColor White
Write-Host ""
Write-Host "📚 For more examples, see: CLI_EXAMPLES.md" -ForegroundColor Cyan
Write-Host "📖 For technical details, see: IMPLEMENTATION.md" -ForegroundColor Cyan
