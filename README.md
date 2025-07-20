# Anti-Bot TLS Fingerprint Scraper

This project implements an advanced web scraper that uses TLS fingerprinting techniques to bypass anti-bot detection systems.

## Features

- **TLS Fingerprinting**: Uses uTLS library to mimic real browser signatures (Chrome, Firefox, Safari, Edge)
- **HTTP/2 Support**: Custom implementation to handle modern web protocols
- **Anti-Detection**: Custom headers, user agent rotation, rate limiting
- **Cookie Management**: Automatic session handling and persistence
- **CLI Interface**: Easy-to-use command-line tool
- **Advanced Features**: Retry logic, proxy support, custom headers

## Quick Start

### Build the CLI tool
```bash
go build -o bin/scraper.exe ./cmd/scraper
```

### Basic usage
```bash
# Simple GET request
./bin/scraper.exe -url https://httpbin.org/headers

# Different browser fingerprints
./bin/scraper.exe -url https://httpbin.org/headers -browser firefox -verbose

# POST request with data
./bin/scraper.exe -url https://httpbin.org/post -method POST -data "@test_data.json"
```

## CLI Reference

### Command Line Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-url` | string | required | URL to scrape |
| `-browser` | string | chrome | Browser fingerprint (chrome, firefox, safari, edge) |
| `-method` | string | GET | HTTP method (GET, POST) |
| `-headers` | string | | Custom headers in JSON format or @filename |
| `-data` | string | | POST data in JSON format or @filename |
| `-output` | string | text | Output format (text, json) |
| `-retries` | int | 3 | Number of retries |
| `-rate-limit` | duration | 1s | Rate limit between requests |
| `-user-agent` | string | | Custom User-Agent (overrides browser default) |
| `-timeout` | duration | 30s | Request timeout |
| `-verbose` | bool | false | Verbose output |
| `-show-headers` | bool | false | Show response headers |
| `-version` | bool | false | Show version information |

### Examples

```bash
# Chrome fingerprint with verbose output
./bin/scraper.exe -url https://httpbin.org/headers -browser chrome -verbose

# Firefox with custom headers
./bin/scraper.exe -url https://httpbin.org/headers -browser firefox -headers "@headers.json"

# POST request with form data
./bin/scraper.exe -url https://httpbin.org/post -method POST -data "@login.json"

# JSON output with rate limiting
./bin/scraper.exe -url https://example.com -output json -rate-limit 3s -retries 5
```

## Dependencies

- [uTLS](https://github.com/refraction-networking/utls) - TLS fingerprinting library
- Go 1.21+ - Modern Go features

## Usage as Library

```go
package main

import (
    "fmt"
    "anti-bot-scraper/internal/scraper"
)

func main() {
    s := scraper.NewScraper()
    response, err := s.Get("https://example.com")
    if err != nil {
        panic(err)
    }
    fmt.Println(response)
}
```
