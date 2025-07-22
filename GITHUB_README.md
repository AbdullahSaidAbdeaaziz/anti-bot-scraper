# 🕵️ Anti-Bot TLS Fingerprint Scraper

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![TLS](https://img.shields.io/badge/TLS-Fingerprinting-red.svg)](https://github.com/refraction-networking/utls)

> **Advanced web scraper using TLS fingerprinting to bypass anti-bot detection systems**

## 🎯 What Makes This Special?

Unlike traditional scrapers that only modify HTTP headers, this tool operates at the **TLS protocol level**, creating authentic browser fingerprints that are virtually indistinguishable from real user traffic.

## ⚡ Quick Start

```bash
# Build the scraper
go build -o bin/scraper.exe ./cmd/scraper

# Test with Chrome fingerprint
./bin/scraper.exe -url https://httpbin.org/headers -browser chrome -verbose

# Test with Firefox fingerprint  
./bin/scraper.exe -url https://httpbin.org/headers -browser firefox -verbose
```

## 🚀 Features

- **🔒 TLS Fingerprinting**: Mimics Chrome, Firefox, Safari, and Edge TLS handshakes
- **🎭 Anti-Detection**: Browser-specific headers and behaviors
- **⚙️ CLI Interface**: Professional command-line tool with 15+ options
- **🍪 Session Management**: Automatic cookie handling and persistence
- **🔄 Retry Logic**: Exponential backoff for robust error handling
- **📊 Multiple Formats**: Text and JSON output options

## 📖 Usage Examples

```bash
# Different browser fingerprints
./bin/scraper.exe -url https://example.com -browser chrome
./bin/scraper.exe -url https://example.com -browser firefox
./bin/scraper.exe -url https://example.com -browser safari
./bin/scraper.exe -url https://example.com -browser edge

# POST request with data
./bin/scraper.exe -url https://httpbin.org/post -method POST -data "@data.json"

# Advanced options
./bin/scraper.exe -url https://example.com -browser chrome -rate-limit 3s -retries 5 -output json
```

## 🔧 CLI Options

| Flag | Description | Default |
|------|-------------|---------|
| `-url` | Target URL to scrape | *required* |
| `-browser` | Browser fingerprint (chrome/firefox/safari/edge) | chrome |
| `-method` | HTTP method (GET/POST) | GET |
| `-headers` | Custom headers (JSON or @file) | - |
| `-data` | POST data (JSON or @file) | - |
| `-output` | Output format (text/json) | text |
| `-retries` | Number of retries | 3 |
| `-rate-limit` | Rate limit between requests | 1s |
| `-verbose` | Verbose output | false |

## 🛠️ Installation

### Prerequisites
- Go 1.21 or later
- Internet connection for dependencies

### Build from Source
```bash
git clone [your-repo-url]
cd anti-bot-scraper
go mod download
go build -o bin/scraper.exe ./cmd/scraper
```

## 🎮 Demo

Check out the `demo/` folder for working examples:
- API header inspection
- Login form submission  
- Stealth browsing patterns

```bash
# Run the quick demo
cd demo
./quick_demo.bat
```

## 🏗️ Architecture

```
anti-bot-scraper/
├── cmd/scraper/              # CLI application
├── internal/scraper/         # Core scraping engine
│   ├── scraper.go           # Basic scraper with TLS fingerprinting
│   ├── fingerprints.go      # Browser-specific configurations
│   └── advanced.go          # Advanced features (retry, cookies)
├── demo/                    # Working examples and demos
└── docs/                    # Documentation
```

## 🔬 How It Works

1. **TLS Fingerprinting**: Uses uTLS library to replicate browser TLS handshakes
2. **Custom ClientHello**: Avoids HTTP/2 compatibility issues
3. **Browser Headers**: Each fingerprint includes unique headers
4. **Session Management**: Handles cookies and persistence
5. **Anti-Detection**: Rate limiting and human-like patterns

## 📊 Browser Fingerprints

| Browser | Key Features |
|---------|--------------|
| **Chrome** | Sec-Ch-Ua headers, modern cipher suites |
| **Firefox** | DNT header, Mozilla-specific patterns |
| **Safari** | WebKit User-Agent, Safari accept headers |
| **Edge** | Chromium-Edge hybrid fingerprint |

## 🎯 Use Cases

- ✅ Competitive intelligence and market research
- ✅ SEO monitoring and SERP analysis  
- ✅ Security testing and penetration testing
- ✅ Academic research on web technologies
- ✅ Public data collection and business intelligence

## ⚠️ Ethical Usage

This tool is designed for legitimate purposes such as:
- Testing your own applications
- Academic research
- Competitive analysis of public data
- Security testing with proper authorization

**Please respect robots.txt, rate limits, and terms of service of target websites.**

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## ⭐ Star This Project

If you find this project useful, please consider giving it a star on GitHub!

---

**Built with ❤️ using Go and uTLS**
