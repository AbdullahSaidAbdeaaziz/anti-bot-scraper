#!/usr/bin/env bash
# Anti-Bot Scraper - Comprehensive Demo Script
# This script demonstrates all the key features of the scraper

SCRAPER="./bin/scraper.exe"
DEMO_DIR="./demo"

echo "🚀 Anti-Bot TLS Fingerprint Scraper - Final Demo"
echo "=================================================="
echo

# Check if scraper is built
if [ ! -f "$SCRAPER" ]; then
    echo "❌ Scraper not found. Building..."
    go build -o bin/scraper.exe ./cmd/scraper
    echo "✅ Scraper built successfully"
    echo
fi

echo "📋 Demo 1: Basic Browser Fingerprint Testing"
echo "--------------------------------------------"
browsers=("chrome" "firefox" "safari" "edge")

for browser in "${browsers[@]}"; do
    echo "🌐 Testing $browser fingerprint..."
    $SCRAPER -url "https://httpbin.org/headers" -browser "$browser" -output json | jq -r '.body | fromjson | .headers["User-Agent"]'
    echo
done

echo "📋 Demo 2: Custom Headers and API Testing"
echo "-----------------------------------------"
echo "🔑 Testing with API headers..."
$SCRAPER -url "https://httpbin.org/headers" \
         -browser chrome \
         -headers "@$DEMO_DIR/api_headers.json" \
         -output json | jq -r '.body | fromjson | .headers | to_entries[] | select(.key | startswith("X-")) | "\(.key): \(.value)"'
echo

echo "📋 Demo 3: POST Request with Form Data"
echo "--------------------------------------"
echo "📝 Sending login form data..."
$SCRAPER -url "https://httpbin.org/post" \
         -method POST \
         -browser firefox \
         -data "@$DEMO_DIR/login_data.json" \
         -verbose \
         -show-headers | head -20
echo

echo "📋 Demo 4: Stealth Mode with Custom Headers"
echo "-------------------------------------------"
echo "🥷 Testing stealth headers..."
$SCRAPER -url "https://httpbin.org/headers" \
         -browser safari \
         -headers "@$DEMO_DIR/stealth_headers.json" \
         -user-agent "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15" \
         -output text | grep -E "(Referer|X-Forwarded-For|User-Agent)" | head -5
echo

echo "📋 Demo 5: Rate Limiting and Retry Logic"
echo "----------------------------------------"
echo "⏱️  Testing rate limiting with multiple requests..."
for i in {1..3}; do
    echo "Request $i:"
    $SCRAPER -url "https://httpbin.org/delay/1" \
             -browser chrome \
             -rate-limit 2s \
             -timeout 10s \
             -verbose 2>&1 | grep -E "(Making GET|Status:|took)"
done
echo

echo "📋 Demo 6: Error Handling and Retries"
echo "-------------------------------------"
echo "🔄 Testing retry logic with failing endpoint..."
$SCRAPER -url "https://httpbin.org/status/503" \
         -browser edge \
         -retries 3 \
         -rate-limit 1s \
         -verbose 2>&1 | grep -E "(Making GET|Error:|failed after)"
echo

echo "📋 Demo 7: Cookie Session Management"
echo "------------------------------------"
echo "🍪 Testing cookie persistence..."
echo "Setting a cookie:"
$SCRAPER -url "https://httpbin.org/cookies/set?demo=session123&user=testuser" \
         -browser chrome \
         -output json | jq -r '.body | fromjson | .cookies // "No cookies in response"'

echo
echo "Reading cookies (should show the set cookies):"
$SCRAPER -url "https://httpbin.org/cookies" \
         -browser chrome \
         -output json | jq -r '.body | fromjson | .cookies // "No cookies found"'
echo

echo "🎉 Demo Complete!"
echo "=================="
echo
echo "💡 Pro Tips:"
echo "- Use different browser fingerprints for different targets"
echo "- Combine with custom headers for maximum stealth"
echo "- Adjust rate limiting based on target's rate limits"
echo "- Use POST requests for form submissions and API calls"
echo "- Leverage cookie management for session-based scraping"
echo
echo "📚 For more examples, see: CLI_EXAMPLES.md"
echo "📖 For technical details, see: IMPLEMENTATION.md"
