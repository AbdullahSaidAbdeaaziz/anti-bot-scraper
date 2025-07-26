# Anti-Bot TLS Fingerprint Scraper

Advanced web scraper using TLS fingerprinting to bypass anti-bot detection systems by mimicking real browser traffic.

## Features

- **TLS Fingerprinting**: Uses uTLS library to mimic Chrome, Firefox, Safari, Edge signatures
- **Anti-Detection**: Browser-specific headers, rate limiting, retry logic
- **Cookie Management**: Automatic session handling and persistence  
- **CLI Interface**: Easy-to-use command-line tool with comprehensive options

## Quick Start

```bash
# Build
go build -o bin/scraper.exe ./cmd/scraper

# Basic usage
./bin/scraper.exe -url https://httpbin.org/headers

# Different browsers
./bin/scraper.exe -url https://httpbin.org/headers -browser firefox

# POST request  
./bin/scraper.exe -url https://httpbin.org/post -method POST -data '{"key":"value"}'
```

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-url` | Target URL (required) | - |
| `-browser` | Browser fingerprint (chrome, firefox, safari, edge) | chrome |
| `-method` | HTTP method (GET, POST) | GET |
| `-data` | POST data (JSON or @filename) | - |
| `-headers` | Custom headers (JSON or @filename) | - |
| `-output` | Output format (text, json) | text |
| `-proxy` | Single proxy URL (http://proxy:port, socks5://proxy:port) | - |
| `-proxies` | Multiple proxies for rotation (comma-separated) | - |
| `-proxy-rotation` | Rotation mode: 'per-request' or 'on-error' | per-request |
| `-retries` | Number of retries | 3 |
| `-rate-limit` | Rate limit between requests | 1s |
| `-timeout` | Request timeout | 30s |
| `-verbose` | Verbose output | false |

## Examples

See [CLI Examples](./CLI_EXAMPLES.md) for comprehensive usage examples.

## Dependencies

- [uTLS](https://github.com/refraction-networking/utls) - TLS fingerprinting library
- Go 1.21+ required

## License

MIT License
