# 🚀 Quick Setup Guide

## Prerequisites
- Go 1.21 or later
- Windows/Linux/macOS

## 1. Clone or Download
```bash
# If using git
git clone <repository-url>
cd anti-bot-scraper

# Or download and extract the ZIP file
```

## 2. Build the CLI
```bash
go build -o bin/scraper.exe ./cmd/scraper
```

## 3. Quick Test
```bash
# Windows
.\bin\scraper.exe -url https://httpbin.org/headers -verbose

# Linux/macOS
./bin/scraper -url https://httpbin.org/headers -verbose
```

## 4. Run Demo (Windows)
```bash
.\demo\quick_demo.bat
```

## 5. Explore Examples
- Check `CLI_EXAMPLES.md` for usage examples
- Try different browser fingerprints: chrome, firefox, safari, edge
- Test with custom headers using files in `demo/` folder

## Need Help?
```bash
.\bin\scraper.exe -help
```

That's it! You're ready to scrape with TLS fingerprinting! 🎯
