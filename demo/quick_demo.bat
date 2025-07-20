@echo off
echo 🚀 Anti-Bot TLS Fingerprint Scraper - Quick Demo
echo =================================================
echo.

REM Check if scraper exists
if not exist "bin\scraper.exe" (
    echo ❌ Building scraper...
    go build -o bin\scraper.exe .\cmd\scraper
    echo ✅ Scraper built successfully
    echo.
)

echo 📋 Demo 1: Different Browser Fingerprints
echo ------------------------------------------
echo 🌐 Chrome fingerprint:
bin\scraper.exe -url https://httpbin.org/headers -browser chrome -output json | findstr "User-Agent"
echo.

echo 🌐 Firefox fingerprint:
bin\scraper.exe -url https://httpbin.org/headers -browser firefox -output json | findstr "User-Agent"
echo.

echo 🌐 Safari fingerprint:
bin\scraper.exe -url https://httpbin.org/headers -browser safari -output json | findstr "User-Agent"
echo.

echo 📋 Demo 2: API Headers Test
echo ----------------------------
echo 🔑 Testing with custom API headers...
bin\scraper.exe -url https://httpbin.org/headers -headers "@demo\api_headers.json" -browser chrome
echo.

echo 📋 Demo 3: POST Request Test
echo -----------------------------
echo 📝 Sending form data...
bin\scraper.exe -url https://httpbin.org/post -method POST -data "@demo\login_data.json" -browser firefox -show-headers | findstr /C:"Status:" /C:"form"
echo.

echo 📋 Demo 4: Rate Limiting Test
echo ------------------------------
echo ⏱️ Testing with 3-second delays...
bin\scraper.exe -url https://httpbin.org/delay/1 -rate-limit 3s -verbose
echo.

echo 🎉 Demo Complete!
echo =================
echo.
echo 💡 Try these commands:
echo   bin\scraper.exe -help
echo   bin\scraper.exe -url https://httpbin.org/ip -browser edge -verbose
echo   bin\scraper.exe -url https://httpbin.org/json -output json
echo.
pause
