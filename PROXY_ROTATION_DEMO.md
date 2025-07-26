# Proxy Rotation Demo

This demonstrates the new proxy rotation functionality in the Anti-Bot TLS Fingerprint Scraper.

## Features Added

### 1. **Proxy Rotation Modes**
- **Per-Request Rotation**: Switches to next proxy on every request
- **On-Error Rotation**: Only switches proxy when current one fails/times out

### 2. **Multiple Proxy Support**
- Support for multiple proxies in rotation
- Comma-separated proxy list
- Mix HTTP and SOCKS5 proxies in same rotation

### 3. **Failure Tracking**
- Tracks failure count per proxy
- Automatically rotates to next proxy on failure
- Marks problematic proxies for debugging

## Usage Examples

### Single Proxy (Original)
```bash
./bin/scraper.exe -url https://httpbin.org/ip -proxy http://proxy.example.com:8080
```

### Multiple Proxy Rotation - Per Request
```bash
./bin/scraper.exe -url https://httpbin.org/ip \
  -proxies "http://proxy1.com:8080,http://proxy2.com:8080,socks5://proxy3.com:1080" \
  -proxy-rotation per-request \
  -verbose
```

### Multiple Proxy Rotation - On Error Only
```bash
./bin/scraper.exe -url https://httpbin.org/ip \
  -proxies "http://proxy1.com:8080,http://proxy2.com:8080" \
  -proxy-rotation on-error \
  -retries 5 \
  -verbose
```

## How It Works

### Per-Request Rotation
1. Makes request using proxy #1
2. Makes next request using proxy #2  
3. Makes next request using proxy #3
4. Cycles back to proxy #1
5. Continues rotation for every request

### On-Error Rotation  
1. Makes request using proxy #1
2. If successful, continues using proxy #1
3. If proxy #1 fails/times out, switches to proxy #2
4. Only rotates when current proxy has issues

## CLI Flags Added

| Flag | Description | Default |
|------|-------------|---------|
| `-proxies` | Multiple proxies for rotation (comma-separated) | - |
| `-proxy-rotation` | Rotation mode: 'per-request' or 'on-error' | per-request |

## Benefits

- **Higher Success Rate**: If one proxy fails, automatically tries next
- **Load Distribution**: Spreads requests across multiple proxies
- **Anti-Detection**: Different IP addresses make scraping harder to detect
- **Fallback Support**: Automatic failover when proxies go down
- **Flexible Rotation**: Choose when to rotate based on needs

## Technical Implementation

- Thread-safe proxy rotation with mutex locks
- Failure counting per proxy for debugging
- Seamless integration with existing retry logic
- Support for mixed proxy types (HTTP/SOCKS5) in rotation
- Automatic proxy configuration on rotation
