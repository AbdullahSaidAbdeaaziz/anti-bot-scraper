# 🚀 Anti-Bot TLS Fingerprint Scraper

**Enterprise-grade web scraper with advanced anti-detection capabilities**

A sophisticated web scraping tool that uses cutting-edge TLS fingerprinting, human behavior simulation, and intelligent evasion techniques to bypass modern anti-bot systems. Built with Go for high performance and reliability.

## 📚 Table of Contents

### 🎯 **Getting Started**
- [✨ Features Overview](#-features-overview)
- [🚀 Quick Start](#-quick-start)
  - [Installation](#installation)
  - [Basic Usage Examples](#basic-usage-examples)

### 🛠️ **Configuration & Usage**
- [🚀 Enhanced Configuration Features](#-enhanced-configuration-features)
  - [📝 Configurable Input Sources](#-configurable-input-sources)
  - [🎭 Enhanced TLS Profile Management](#-enhanced-tls-profile-management)
  - [🛡️ Advanced Header Mimicry](#️-advanced-header-mimicry)
  - [🍪 Enhanced Cookie & Redirect Handling](#-enhanced-cookie--redirect-handling)
  - [🔄 File-Based Proxy Management](#-file-based-proxy-management)
  - [🎯 Comprehensive Evasion Example](#-comprehensive-evasion-example)

### 🎮 **Feature Demonstrations**
- [🎯 Core Feature Demonstrations](#-core-feature-demonstrations)
  - [🛡️ TLS Fingerprinting & HTTP/2](#️-tls-fingerprinting--http2)
  - [🧠 JavaScript Engine & Dynamic Content](#-javascript-engine--dynamic-content)
  - [🎭 Human Behavior Simulation](#-human-behavior-simulation)
  - [🔄 Intelligent Proxy Management](#-intelligent-proxy-management)
  - [🧩 CAPTCHA Solving](#-captcha-solving)
  - [⚡ High-Performance Concurrent Processing](#-high-performance-concurrent-processing)

### 📖 **Reference & Documentation**
- [📋 Complete Command Reference](#-complete-command-reference)
- [🏗️ Architecture Overview](#️-architecture-overview)
- [🎯 Feature Implementation Status](#-feature-implementation-status)
- [📚 Enhanced Configuration Reference](#-enhanced-configuration-reference)
  - [🎯 Core Input Configuration Flags](#-core-input-configuration-flags)
  - [🛡️ TLS & Fingerprinting Flags](#️-tls--fingerprinting-flags)
  - [⏱️ Timing & Delay Configuration](#️-timing--delay-configuration)
  - [🎭 Enhanced Header Mimicry Flags](#-enhanced-header-mimicry-flags)
  - [🍪 Cookie & Session Management Flags](#-cookie--session-management-flags)
  - [🔄 Redirect Handling Flags](#-redirect-handling-flags)
  - [📁 File Format Examples](#-file-format-examples)
  - [🚀 Quick Start Examples](#-quick-start-examples)

### 🚀 **Advanced Topics**
- [🚀 Advanced Use Cases](#-advanced-use-cases)
- [🔧 Advanced Configuration](#-advanced-configuration)
- [📈 Performance Benchmarks](#-performance-benchmarks)

### 🛠️ **Development & Support**
- [🛠️ Development & Testing](#️-development--testing)
- [🔍 Troubleshooting](#-troubleshooting)
- [🛡️ Security & Ethics](#️-security--ethics)
- [📚 Dependencies & Requirements](#-dependencies--requirements)

### 📄 **Legal & Credits**
- [📄 License](#-license)
- [🙏 Acknowledgments](#-acknowledgments)

---

## ✨ Features Overview

### 🔐 **Core Anti-Detection**
- **🛡️ Advanced TLS Fingerprinting**: uTLS library with Chrome, Firefox, Safari, Edge signatures
- **🌐 HTTP/2 Support**: Full HTTP/2 transport with enhanced authenticity
- **🧠 JavaScript Engine**: Headless browser automation with chromedp
- **🎭 Behavioral Simulation**: Human-like interaction patterns with 4 behavior types
- **🔄 Intelligent Proxy Management**: Health monitoring and smart rotation
- **🧩 CAPTCHA Integration**: Multi-service solving with 4 major providers

### ⚡ **Performance & Scalability**
- **🚄 Concurrent Processing**: Worker pools with intelligent request management
- **💾 Memory Optimization**: Automatic garbage collection and resource management
- **📋 Priority Queueing**: Smart request prioritization and load balancing
- **🔗 Connection Pooling**: HTTP connection reuse for optimal performance

### 🛠️ **Professional Tools**
- **💻 Comprehensive CLI**: 45+ configuration flags for complete control
- **📊 Real-time Monitoring**: Performance, memory, and queue statistics
- **🔧 Flexible Configuration**: JSON headers, file-based data, proxy rotation
- **📈 Detailed Analytics**: Success rates, latency tracking, error analysis

## 🚀 Quick Start

### Installation
```bash
# Clone the repository
git clone https://github.com/AbdullahSaidAbdeaaziz/anti-bot-scraper.git
cd anti-bot-scraper

# Build the scraper
go build -o bin/scraper.exe ./cmd/scraper

# Test basic functionality
./bin/scraper.exe -url https://httpbin.org/headers -verbose
```

### Basic Usage Examples

```bash
# Simple GET request with Chrome fingerprint
./bin/scraper.exe -url https://httpbin.org/headers

# POST request with custom data
./bin/scraper.exe -url https://httpbin.org/post -method POST -data '{"test":"data"}'

# Advanced browser simulation with JavaScript
./bin/scraper.exe -url https://example.com -browser firefox -enable-js -js-mode behavior

# High-performance concurrent scraping
./bin/scraper.exe -url https://httpbin.org/headers -enable-concurrent -max-concurrent 20 -show-performance-stats
```

## 🚀 Enhanced Configuration Features

### 📝 **Configurable Input Sources**
```bash
# Single URL processing
./scraper -url https://httpbin.org/headers -verbose

# Multiple URLs from file
./scraper -urls-file examples/urls.txt -num-requests 3 -verbose

# Multiple requests per URL with delays
./scraper -url https://httpbin.org/ip -num-requests 5 \
  -delay-min 1s -delay-max 3s -delay-randomize
```

### 🎭 **Enhanced TLS Profile Management**
```bash
# Fixed TLS profile
./scraper -url https://httpbin.org/headers -tls-profile chrome -verbose

# Randomized TLS profiles across requests
./scraper -urls-file examples/urls.txt -tls-randomize -num-requests 3 -verbose

# TLS randomization with specific timing
./scraper -url https://httpbin.org/headers -tls-randomize \
  -num-requests 10 -delay-min 500ms -delay-max 2s
```

### 🛡️ **Advanced Header Mimicry**
```bash
# Automatic header profile matching TLS fingerprint
./scraper -url https://httpbin.org/headers -header-mimicry \
  -header-profile auto -enable-sec-headers -verbose

# Custom header configuration
./scraper -url https://httpbin.org/headers -header-mimicry \
  -header-profile firefox -accept-language "en-US,en;q=0.5" \
  -accept-encoding "gzip, deflate, br" -enable-sec-headers=false

# Override User-Agent with custom value
./scraper -url https://httpbin.org/user-agent \
  -custom-user-agent "Mozilla/5.0 (Custom Browser)" -verbose
```

### 🍪 **Enhanced Cookie & Redirect Handling**
```bash
# Session-based cookie persistence
./scraper -url https://httpbin.org/cookies/set/test/value \
  -cookie-jar -cookie-persistence session -verbose

# Proxy-based cookie isolation
./scraper -url https://httpbin.org/cookies \
  -cookie-jar -cookie-persistence proxy -proxy-file examples/proxies.txt

# Advanced redirect handling
./scraper -url https://httpbin.org/redirect/5 \
  -follow-redirects -max-redirects 10 -redirect-timeout 30s -verbose

# Cookie file persistence
./scraper -url https://httpbin.org/cookies/set/persistent/data \
  -cookie-file session.cookies -cookie-jar -verbose
```

### 🔄 **File-Based Proxy Management**
```bash
# Load proxies from file with round-robin rotation
./scraper -urls-file examples/urls.txt -proxy-file examples/proxies.txt \
  -num-requests 5 -verbose

# Combined with enhanced evasion features
./scraper -url https://httpbin.org/ip -proxy-file examples/proxies.txt \
  -tls-randomize -header-mimicry -num-requests 3 \
  -delay-min 1s -delay-max 3s -verbose
```

### 🎯 **Comprehensive Evasion Example**
```bash
# All enhanced features combined
./scraper -urls-file examples/urls.txt \
  -num-requests 2 \
  -tls-randomize \
  -header-mimicry \
  -header-profile auto \
  -enable-sec-headers \
  -cookie-jar \
  -cookie-persistence session \
  -follow-redirects \
  -max-redirects 5 \
  -delay-min 800ms \
  -delay-max 2500ms \
  -delay-randomize \
  -proxy-file examples/proxies.txt \
  -accept-language "en-US,en;q=0.9" \
  -output json \
  -verbose
```

## 🎯 Core Feature Demonstrations

### 🛡️ **TLS Fingerprinting & HTTP/2**
```bash
# HTTP/2 with Chrome fingerprint
./bin/scraper.exe -url https://httpbin.org/headers -http-version 2 -browser chrome

# Automatic protocol selection
./bin/scraper.exe -url https://httpbin.org/headers -http-version auto -verbose
```

### 🧠 **JavaScript Engine & Dynamic Content**
```bash
# Standard JavaScript execution
./bin/scraper.exe -url https://example.com -enable-js -js-mode standard

# Human behavior simulation
./bin/scraper.exe -url https://example.com -enable-js -js-mode behavior -viewport 1920x1080

# Wait for specific elements
./bin/scraper.exe -url https://example.com -enable-js -js-mode wait-element -js-wait-selector ".content"
```

### 🎭 **Human Behavior Simulation**
```bash
# Normal human behavior
./bin/scraper.exe -url https://example.com -enable-behavior -behavior-type normal

# Cautious browsing pattern
./bin/scraper.exe -url https://example.com -enable-behavior -behavior-type cautious -show-behavior-info

# Random behavior with custom timing
./bin/scraper.exe -url https://example.com -enable-behavior -behavior-type random \
  -behavior-min-delay 800ms -behavior-max-delay 3s -enable-random-activity
```

### 🔄 **Intelligent Proxy Management**
```bash
# Health-aware proxy rotation
./bin/scraper.exe -url https://httpbin.org/ip \
  -proxies "proxy1:8080,proxy2:8080,proxy3:8080" \
  -proxy-rotation health-aware -enable-proxy-health -show-proxy-metrics

# Custom health monitoring
./bin/scraper.exe -url https://httpbin.org/ip -proxies "proxy1:8080,proxy2:8080" \
  -enable-proxy-health -proxy-health-interval 2m -proxy-health-timeout 5s
```

### 🧩 **CAPTCHA Solving**
```bash
# 2captcha integration
./bin/scraper.exe -url https://example.com/captcha \
  -enable-captcha -captcha-service 2captcha -captcha-api-key "YOUR_API_KEY" -show-captcha-info

# Multiple service support
./bin/scraper.exe -url https://example.com/recaptcha \
  -enable-captcha -captcha-service anticaptcha -captcha-api-key "YOUR_KEY" \
  -captcha-min-score 0.7 -captcha-timeout 3m
```

### ⚡ **High-Performance Concurrent Processing**
```bash
# Enterprise-grade concurrent scraping
./bin/scraper.exe -url https://httpbin.org/headers \
  -enable-concurrent -max-concurrent 50 -worker-pool-size 10 \
  -requests-per-second 15 -show-performance-stats

# Memory-optimized processing
./bin/scraper.exe -url https://httpbin.org/json \
  -enable-concurrent -enable-memory-optimization -max-memory-mb 256 \
  -enable-intelligent-queue -show-memory-stats

# Full performance suite
./bin/scraper.exe -url https://httpbin.org/headers \
  -enable-concurrent -enable-memory-optimization -enable-intelligent-queue \
  -max-concurrent 30 -worker-pool-size 8 -max-memory-mb 512 \
  -connection-pool-size 50 -show-performance-stats -show-memory-stats
```

## 📋 Complete Command Reference

### 🎯 **Core Options**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-url` | Target URL (required) | - | `https://example.com` |
| `-browser` | Browser fingerprint | `chrome` | `chrome`, `firefox`, `safari`, `edge` |
| `-method` | HTTP method | `GET` | `GET`, `POST` |
| `-data` | POST data (JSON or @file) | - | `'{"key":"value"}'` or `@data.json` |
| `-headers` | Custom headers (JSON or @file) | - | `'{"Auth":"Bearer token"}'` |
| `-output` | Output format | `text` | `text`, `json` |

### 🌐 **HTTP & Protocol Options**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-http-version` | HTTP protocol version | `1.1` | `1.1`, `2`, `auto` |
| `-user-agent` | Custom User-Agent | - | `"Custom Bot 1.0"` |
| `-timeout` | Request timeout | `30s` | `60s`, `2m` |
| `-retries` | Number of retries | `3` | `5` |
| `-rate-limit` | Rate limit between requests | `1s` | `500ms`, `2s` |
| `-follow-redirect` | Follow redirects | `true` | `true`, `false` |

### 🔄 **Proxy Management**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-proxy` | Single proxy URL | - | `http://proxy:8080` |
| `-proxies` | Multiple proxies (comma-separated) | - | `proxy1:8080,proxy2:8080` |
| `-proxy-rotation` | Rotation mode | `per-request` | `per-request`, `on-error`, `health-aware` |
| `-enable-proxy-health` | Enable health monitoring | `false` | `true` |
| `-proxy-health-interval` | Health check interval | `5m` | `2m`, `10m` |
| `-proxy-health-timeout` | Health check timeout | `10s` | `5s`, `30s` |
| `-proxy-health-test-url` | Health test URL | `https://httpbin.org/ip` | Custom URL |
| `-proxy-max-failures` | Max failures before disable | `3` | `5` |
| `-show-proxy-metrics` | Show proxy statistics | `false` | `true` |

### 🧠 **JavaScript Engine**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-enable-js` | Enable JavaScript engine | `false` | `true` |
| `-js-mode` | JavaScript execution mode | `standard` | `standard`, `behavior`, `wait-element` |
| `-js-code` | Custom JavaScript code | - | `"console.log('test')"` |
| `-js-timeout` | JavaScript timeout | `30s` | `60s` |
| `-js-wait-selector` | CSS selector to wait for | - | `".content"`, `"#result"` |
| `-headless` | Run browser headless | `true` | `true`, `false` |
| `-no-images` | Disable image loading | `false` | `true` |
| `-viewport` | Browser viewport size | `1920x1080` | `1366x768` |

### 🎭 **Human Behavior Simulation**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-enable-behavior` | Enable behavior simulation | `false` | `true` |
| `-behavior-type` | Behavior pattern | `normal` | `normal`, `cautious`, `aggressive`, `random` |
| `-behavior-min-delay` | Minimum action delay | `500ms` | `300ms` |
| `-behavior-max-delay` | Maximum action delay | `2s` | `5s` |
| `-enable-mouse-movement` | Enable mouse simulation | `true` | `true`, `false` |
| `-enable-scroll-simulation` | Enable scroll simulation | `true` | `true`, `false` |
| `-enable-typing-delay` | Enable typing delays | `true` | `true`, `false` |
| `-enable-random-activity` | Enable random interactions | `false` | `true` |
| `-show-behavior-info` | Show behavior details | `false` | `true` |

### 🧩 **CAPTCHA Integration**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-enable-captcha` | Enable CAPTCHA solving | `false` | `true` |
| `-captcha-service` | CAPTCHA service provider | `2captcha` | `2captcha`, `deathbycaptcha`, `anticaptcha`, `capmonster` |
| `-captcha-api-key` | Service API key | - | `"your-api-key-here"` |
| `-captcha-timeout` | Solving timeout | `5m` | `3m`, `10m` |
| `-captcha-poll-interval` | Solution polling interval | `5s` | `3s`, `10s` |
| `-captcha-max-retries` | Max solving retries | `3` | `5` |
| `-captcha-min-score` | Min reCAPTCHA v3 score | `0.3` | `0.7` |
| `-show-captcha-info` | Show CAPTCHA details | `false` | `true` |

### ⚡ **Performance Optimization**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-enable-concurrent` | Enable concurrent processing | `false` | `true` |
| `-max-concurrent` | Maximum concurrent requests | `10` | `50` |
| `-worker-pool-size` | Number of worker goroutines | `5` | `20` |
| `-requests-per-second` | Rate limit (RPS) | `5.0` | `15.0` |
| `-connection-pool-size` | HTTP connection pool size | `20` | `100` |
| `-max-idle-conns` | Maximum idle connections | `10` | `50` |
| `-idle-conn-timeout` | Idle connection timeout | `90s` | `2m` |
| `-queue-size` | Request queue buffer size | `1000` | `5000` |
| `-show-performance-stats` | Show performance metrics | `false` | `true` |

### 💾 **Memory Optimization**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-enable-memory-optimization` | Enable memory management | `false` | `true` |
| `-max-memory-mb` | Maximum memory usage (MB) | `512` | `1024` |
| `-enable-intelligent-queue` | Enable priority queueing | `false` | `true` |
| `-show-memory-stats` | Show memory statistics | `false` | `true` |

### 🛠️ **Output & Debug**
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-verbose` | Verbose output | `false` | `true` |
| `-show-headers` | Show response headers | `false` | `true` |
| `-version` | Show version information | `false` | `true` |

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Anti-Bot Scraper Architecture                │
├─────────────────────────────────────────────────────────────────┤
│  CLI Interface (45+ Configuration Flags)                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ TLS Fingerprint │  │ JavaScript      │  │ Behavior        │  │
│  │ • uTLS Library  │  │ • chromedp      │  │ • Human Actions │  │
│  │ • HTTP/2        │  │ • 3 Exec Modes  │  │ • 4 Patterns    │  │
│  │ • 4 Browsers    │  │ • Element Wait  │  │ • Timing Sim    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Proxy Management│  │ CAPTCHA Solver  │  │ Performance     │  │
│  │ • Health Monitor│  │ • 4 Services    │  │ • Worker Pools  │  │
│  │ • Smart Rotation│  │ • 7 Types       │  │ • Memory Opt    │  │
│  │ • Failover      │  │ • Auto Detect   │  │ • Conn Pooling  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  Core HTTP Engine (uTLS + HTTP/2 + Connection Pooling)          │
└─────────────────────────────────────────────────────────────────┘
```

## 🎯 Feature Implementation Status

### ✅ **Fully Implemented Features**
- **🛡️ TLS Fingerprinting**: Advanced uTLS with 4 browser signatures
- **🌐 HTTP/2 Support**: Full HTTP/2 transport with automatic fallback
- **🧠 JavaScript Engine**: Complete chromedp integration with 3 execution modes
- **🎭 Behavioral Simulation**: 4 behavior types with realistic human patterns
- **🔄 Proxy Management**: Health monitoring with intelligent rotation
- **🧩 CAPTCHA Integration**: 4 services supporting 7 CAPTCHA types
- **⚡ Concurrent Processing**: Worker pools with intelligent request management
- **💾 Memory Optimization**: Automatic garbage collection and resource management
- **📋 Priority Queueing**: Smart request prioritization and load balancing
- **🔗 Connection Pooling**: HTTP connection reuse for optimal performance
- **💻 Professional CLI**: 45+ configuration flags for complete control
- **📊 Real-time Monitoring**: Performance, memory, and queue statistics

### 🔮 **Future Enhancements** 
- **🎨 Dynamic Fingerprints**: Browser version rotation
- **🌍 Geolocation Awareness**: IP-based country detection  
- **📊 Web Dashboard**: Real-time monitoring interface
- **💾 Database Integration**: Result persistence and querying
- **🔧 Config Files**: YAML/TOML configuration support

## 🚀 Advanced Use Cases

### 🎯 **E-commerce Price Monitoring**
```bash
# Monitor product prices with human-like behavior
./bin/scraper.exe -url "https://shop.example.com/product/123" \
  -enable-js -js-mode behavior -enable-behavior -behavior-type cautious \
  -browser firefox -http-version 2 -verbose
```

### 🔍 **Social Media Content Extraction**
```bash
# Extract content with JavaScript execution and CAPTCHA solving
./bin/scraper.exe -url "https://social.example.com/posts" \
  -enable-js -js-mode wait-element -js-wait-selector ".post-content" \
  -enable-captcha -captcha-service 2captcha -captcha-api-key "YOUR_KEY" \
  -enable-behavior -behavior-type normal
```

### 📊 **Market Research at Scale**
```bash
# High-performance concurrent data collection
./bin/scraper.exe -url "https://api.example.com/data" \
  -enable-concurrent -max-concurrent 100 -worker-pool-size 25 \
  -requests-per-second 50 -enable-memory-optimization -max-memory-mb 1024 \
  -proxies "proxy1:8080,proxy2:8080,proxy3:8080" -proxy-rotation health-aware
```

### 🛒 **Inventory Tracking**
```bash
# Track product availability with intelligent queueing
./bin/scraper.exe -url "https://store.example.com/inventory" \
  -enable-concurrent -enable-intelligent-queue -max-concurrent 20 \
  -enable-js -js-mode standard -browser chrome -http-version auto \
  -show-performance-stats -show-memory-stats
```

## 🔧 Advanced Configuration

### 📁 **File-Based Configuration**
```bash
# Headers from file
echo '{"Authorization":"Bearer token","Custom-Header":"value"}' > headers.json
./bin/scraper.exe -url https://api.example.com -headers @headers.json

# POST data from file  
echo '{"user":"admin","action":"login"}' > data.json
./bin/scraper.exe -url https://api.example.com/login -method POST -data @data.json
```

### 🔄 **Complex Proxy Setups**
```bash
# Health-aware rotation with custom settings
./bin/scraper.exe -url https://httpbin.org/ip \
  -proxies "proxy1.example.com:8080,proxy2.example.com:8080,proxy3.example.com:8080" \
  -proxy-rotation health-aware -enable-proxy-health \
  -proxy-health-interval 30s -proxy-health-timeout 5s \
  -proxy-max-failures 2 -show-proxy-metrics
```

### 🎭 **Realistic Browser Simulation**
```bash
# Complete browser simulation with all features
./bin/scraper.exe -url "https://example.com" \
  -browser firefox -http-version 2 -enable-js -js-mode behavior \
  -enable-behavior -behavior-type normal -enable-mouse-movement \
  -enable-scroll-simulation -enable-typing-delay -viewport 1366x768 \
  -behavior-min-delay 800ms -behavior-max-delay 3s
```

## 📈 Performance Benchmarks

### ⚡ **Concurrent Performance**
```bash
# Benchmark concurrent vs sequential performance
# Sequential (baseline)
time ./bin/scraper.exe -url https://httpbin.org/delay/1

# Concurrent (10x faster)
time ./bin/scraper.exe -url https://httpbin.org/delay/1 \
  -enable-concurrent -max-concurrent 10 -show-performance-stats
```

### 💾 **Memory Efficiency**
```bash
# Monitor memory usage during high-volume operations
./bin/scraper.exe -url https://httpbin.org/json \
  -enable-concurrent -max-concurrent 50 -enable-memory-optimization \
  -max-memory-mb 256 -show-memory-stats -verbose
```

## 🛠️ Development & Testing

### 🧪 **Testing Different Scenarios**
```bash
# Test TLS fingerprinting detection
./bin/scraper.exe -url https://tls.browserleaks.com/json -browser chrome -verbose

# Test HTTP/2 support
./bin/scraper.exe -url https://http2.akamai.com/demo -http-version 2 -verbose

# Test proxy functionality  
./bin/scraper.exe -url https://httpbin.org/ip -proxy "proxy.example.com:8080" -verbose

# Test JavaScript capabilities
./bin/scraper.exe -url https://httpbin.org/html -enable-js -js-mode standard -verbose
```

### 📊 **Performance Analysis**
```bash
# Comprehensive performance testing
./bin/scraper.exe -url https://httpbin.org/headers \
  -enable-concurrent -enable-memory-optimization -enable-intelligent-queue \
  -show-performance-stats -show-memory-stats -verbose \
  -max-concurrent 25 -worker-pool-size 10 -requests-per-second 20
```

## 🔍 Troubleshooting

### ❌ **Common Issues**

1. **CAPTCHA Solving Failures**
   ```bash
   # Check API key and service status
   ./bin/scraper.exe -url https://example.com -enable-captcha \
     -captcha-service 2captcha -captcha-api-key "YOUR_KEY" \
     -show-captcha-info -verbose
   ```

2. **Proxy Connection Issues**
   ```bash
   # Test proxy health and rotation
   ./bin/scraper.exe -url https://httpbin.org/ip \
     -proxies "proxy1:8080,proxy2:8080" -enable-proxy-health \
     -show-proxy-metrics -verbose
   ```

3. **JavaScript Execution Problems**
   ```bash
   # Debug JavaScript issues
   ./bin/scraper.exe -url https://example.com -enable-js \
     -js-mode standard -headless=false -verbose
   ```

4. **Memory Usage Optimization**
   ```bash
   # Monitor and optimize memory usage
   ./bin/scraper.exe -url https://example.com -enable-concurrent \
     -enable-memory-optimization -max-memory-mb 256 \
     -show-memory-stats -verbose
   ```

## 🛡️ Security & Ethics

### ⚖️ **Responsible Usage**
- **Respect robots.txt**: Always check and follow website policies
- **Rate Limiting**: Use appropriate delays to avoid overwhelming servers
- **Terms of Service**: Ensure compliance with website terms
- **Legal Compliance**: Follow applicable laws and regulations

### 🔒 **Security Best Practices**
- **API Keys**: Store CAPTCHA service keys securely
- **Proxy Authentication**: Use secure proxy credentials
- **Data Handling**: Protect scraped data according to privacy laws
- **Network Security**: Use VPNs and secure connections when necessary

## 📚 Dependencies & Requirements

### 🔧 **Core Dependencies**
- **[uTLS](https://github.com/refraction-networking/utls)**: Advanced TLS fingerprinting library
- **[chromedp](https://github.com/chromedp/chromedp)**: Chrome DevTools Protocol for JavaScript execution
- **[golang.org/x/net](https://pkg.go.dev/golang.org/x/net)**: Extended networking libraries
- **Go 1.21+**: Modern Go runtime with HTTP/2 support

### 🖥️ **System Requirements**
- **Operating System**: Windows, macOS, Linux
- **Memory**: Minimum 512MB RAM (2GB+ recommended for concurrent processing)
- **Network**: Stable internet connection
- **Chrome/Chromium**: Required for JavaScript engine functionality

### 📦 **Installation Requirements**
```bash
# Verify Go version
go version  # Should be 1.21 or higher

# Install Chrome/Chromium (for JavaScript engine)
# Windows: Download from google.com/chrome
# macOS: brew install --cask google-chrome  
# Linux: apt-get install chromium-browser

# Build the scraper
go build -o bin/scraper.exe ./cmd/scraper
```

## � Enhanced Configuration Reference

### 🎯 **Core Input Configuration Flags**

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-url` | Single URL to scrape | - | `https://example.com` |
| `-urls-file` | File containing multiple URLs | - | `examples/urls.txt` |
| `-num-requests` | Number of requests per URL | `1` | `5` |
| `-proxy-file` | File containing proxy list | - | `examples/proxies.txt` |

### 🛡️ **TLS & Fingerprinting Flags**

| Flag | Description | Default | Options |
|------|-------------|---------|---------|
| `-tls-profile` | TLS profile for fingerprinting | `chrome` | `chrome`, `firefox`, `safari`, `edge` |
| `-tls-randomize` | Randomize TLS profiles across requests | `false` | `true`, `false` |
| `-browser` | Browser fingerprint (legacy) | `chrome` | `chrome`, `firefox`, `safari`, `edge` |

### ⏱️ **Timing & Delay Configuration**

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-delay-min` | Minimum delay between requests | `1s` | `500ms`, `2s` |
| `-delay-max` | Maximum delay between requests | `3s` | `1s`, `5s` |
| `-delay-randomize` | Randomize delays within range | `true` | `true`, `false` |
| `-rate-limit` | Rate limit between requests | `1s` | `500ms`, `2s` |

### 🎭 **Enhanced Header Mimicry Flags**

| Flag | Description | Default | Options |
|------|-------------|---------|---------|
| `-header-mimicry` | Enable browser-consistent headers | `true` | `true`, `false` |
| `-header-profile` | Header profile to use | `auto` | `auto`, `chrome`, `firefox`, `safari`, `edge` |
| `-custom-user-agent` | Custom User-Agent override | - | Custom string |
| `-enable-sec-headers` | Include security headers | `true` | `true`, `false` |
| `-accept-language` | Accept-Language header value | `auto` | `en-US,en;q=0.9` |
| `-accept-encoding` | Accept-Encoding header value | `auto` | `gzip, deflate, br` |

### 🍪 **Cookie & Session Management Flags**

| Flag | Description | Default | Options |
|------|-------------|---------|---------|
| `-cookie-jar` | Enable in-memory cookie jar | `true` | `true`, `false` |
| `-cookie-persistence` | Cookie persistence mode | `session` | `session`, `proxy`, `none` |
| `-cookie-file` | File to save/load cookies | - | `session.cookies` |
| `-clear-cookies` | Clear cookies before each request | `false` | `true`, `false` |

### 🔄 **Redirect Handling Flags**

| Flag | Description | Default | Options |
|------|-------------|---------|---------|
| `-follow-redirects` | Follow HTTP redirects | `true` | `true`, `false` |
| `-max-redirects` | Maximum redirects to follow | `10` | `5`, `20` |
| `-redirect-timeout` | Timeout for redirect chains | `30s` | `10s`, `60s` |

### 📁 **File Format Examples**

#### URLs File (`urls.txt`)
```text
https://httpbin.org/headers
https://httpbin.org/user-agent
https://httpbin.org/ip
# Comments start with #
https://httpbin.org/get
```

#### Proxy File (`proxies.txt`)
```text
http://proxy1.example.com:8080
http://user:pass@proxy2.example.com:3128
socks5://proxy3.example.com:1080
# HTTP and SOCKS5 proxies supported
socks5://user:pass@proxy4.example.com:1080
```

### 🚀 **Quick Start Examples**

#### Basic Enhanced Usage
```bash
# Single URL with enhanced features
./scraper -url https://httpbin.org/headers -header-mimicry -verbose

# Multiple URLs with randomization
./scraper -urls-file examples/urls.txt -tls-randomize -num-requests 2
```

#### Advanced Configuration
```bash
# Comprehensive evasion setup
./scraper -urls-file examples/urls.txt \
  -num-requests 3 \
  -tls-randomize \
  -header-mimicry \
  -header-profile auto \
  -delay-min 1s \
  -delay-max 3s \
  -delay-randomize \
  -cookie-jar \
  -follow-redirects \
  -proxy-file examples/proxies.txt \
  -verbose
```

## �📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **uTLS Team**: For the excellent TLS fingerprinting library
- **chromedp Team**: For browser automation capabilities  
- **Go Community**: For robust networking and HTTP libraries
- **Security Researchers**: For insights into anti-bot detection methods

---

**Built with ❤️ for ethical web scraping and security research**

*Last Updated: July 26, 2025*
