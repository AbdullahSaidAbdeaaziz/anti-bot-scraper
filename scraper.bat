@echo off
REM Anti-Bot Scraper CLI Helper Script

if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="--help" goto help
if "%1"=="-h" goto help

REM Run the scraper with all passed arguments
bin\scraper.exe %*
goto end

:help
echo Anti-Bot Scraper CLI Helper
echo.
echo Usage: scraper.bat [options]
echo.
echo Examples:
echo   scraper.bat -url https://httpbin.org/headers
echo   scraper.bat -url https://httpbin.org/headers -browser firefox -verbose
echo   scraper.bat -url https://httpbin.org/post -method POST -data "@test_data.json"
echo.
echo For full help, run: scraper.bat -help
echo.

:end
