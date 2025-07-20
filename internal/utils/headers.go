package utils

import (
	"math/rand"
	"time"
)

// RandomUserAgent returns a random user agent string
func RandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
	}

	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// RandomDelay adds a random delay between requests
func RandomDelay(minMs, maxMs int) {
	if minMs >= maxMs {
		return
	}

	delay := time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond
	time.Sleep(delay)
}

// GetCommonHeaders returns a map of common HTTP headers
func GetCommonHeaders() map[string]string {
	return map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"Accept-Encoding":           "gzip, deflate, br",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	}
}

// RotateProxies provides a simple proxy rotation mechanism
type ProxyRotator struct {
	proxies []string
	current int
}

// NewProxyRotator creates a new proxy rotator
func NewProxyRotator(proxies []string) *ProxyRotator {
	return &ProxyRotator{
		proxies: proxies,
		current: 0,
	}
}

// Next returns the next proxy in rotation
func (p *ProxyRotator) Next() string {
	if len(p.proxies) == 0 {
		return ""
	}

	proxy := p.proxies[p.current]
	p.current = (p.current + 1) % len(p.proxies)
	return proxy
}
