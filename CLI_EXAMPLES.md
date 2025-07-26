# Anti-Bot Scraper CLI Examples

## Basic Usage

### Simple GET request
```bash
./bin/scraper.exe -url https://httpbin.org/headers
```

### Different browser fingerprints
```bash
./bin/scraper.exe -url https://httpbin.org/headers -browser chrome
./bin/scraper.exe -url https://httpbin.org/headers -browser firefox
./bin/scraper.exe -url https://httpbin.org/headers -browser safari
./bin/scraper.exe -url https://httpbin.org/headers -browser edge
```

### Verbose output with headers
```bash
./bin/scraper.exe -url https://httpbin.org/headers -verbose -show-headers
```

### JSON output
```bash
./bin/scraper.exe -url https://httpbin.org/headers -output json
```

## Advanced Usage

### Proxy Support
```bash
# HTTP proxy
./bin/scraper.exe -url https://httpbin.org/ip -proxy http://proxy.example.com:8080

# SOCKS5 proxy
./bin/scraper.exe -url https://httpbin.org/ip -proxy socks5://proxy.example.com:1080

# Proxy with authentication (HTTP Basic)
./bin/scraper.exe -url https://httpbin.org/ip -proxy http://username:password@proxy.example.com:8080
```

### Custom User-Agent
```bash
./bin/scraper.exe -url https://httpbin.org/headers -user-agent "Custom Bot 1.0"
```

### Custom headers from file
```bash
./bin/scraper.exe -url https://httpbin.org/headers -headers "@test_headers.json"
```

### POST request with data from file
```bash
./bin/scraper.exe -url https://httpbin.org/post -method POST -data "@test_data.json"
```

### Rate limiting and retries
```bash
./bin/scraper.exe -url https://httpbin.org/status/503 -retries 5 -rate-limit 3s -verbose
```

### Complex example with all options
```bash
./bin/scraper.exe \
  -url https://httpbin.org/post \
  -method POST \
  -browser firefox \
  -data "@test_data.json" \
  -headers "@test_headers.json" \
  -retries 3 \
  -rate-limit 2s \
  -timeout 60s \
  -output json \
  -show-headers \
  -verbose
```

## Test Files

### test_data.json (POST data)
```json
{
  "username": "testuser",
  "password": "secret123",
  "action": "login"
}
```

### test_headers.json (Custom headers)
```json
{
  "Authorization": "Bearer token123",
  "X-API-Key": "abc123",
  "Custom-Header": "test-value"
}
```

## Real-world Examples

### Scraping with session management
```bash
# Set cookies
./bin/scraper.exe -url "https://httpbin.org/cookies/set?session=abc123" -browser chrome

# Use cookies (automatic with AdvancedScraper)
./bin/scraper.exe -url "https://httpbin.org/cookies" -browser chrome
```

### API testing with different browsers
```bash
./bin/scraper.exe -url "https://api.example.com/endpoint" -browser chrome -headers "@api_headers.json"
```

### Load testing with rate limiting
```bash
./bin/scraper.exe -url "https://example.com" -retries 10 -rate-limit 5s -verbose
```
