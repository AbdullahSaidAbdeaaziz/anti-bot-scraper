package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"anti-bot-scraper/internal/utils"
	"golang.org/x/net/proxy"
)

// AdvancedScraper extends the basic scraper with additional features
type AdvancedScraper struct {
	*Scraper
	cookieJar   *CookieJar
	proxyURL    string
	retryCount  int
	rateLimiter *RateLimiter
}

// CookieJar manages cookies for the scraper
type CookieJar struct {
	cookies map[string][]*http.Cookie
}

// RateLimiter controls request frequency
type RateLimiter struct {
	requests    chan struct{}
	lastRequest time.Time
	minInterval time.Duration
}

// NewAdvancedScraper creates a new advanced scraper
func NewAdvancedScraper(fingerprint Fingerprint, options ...ScraperOption) (*AdvancedScraper, error) {
	baseScraper, err := NewScraper(fingerprint)
	if err != nil {
		return nil, err
	}

	advanced := &AdvancedScraper{
		Scraper:     baseScraper,
		cookieJar:   NewCookieJar(),
		retryCount:  3,
		rateLimiter: NewRateLimiter(1 * time.Second),
	}

	// Apply options
	for _, option := range options {
		option(advanced)
	}

	return advanced, nil
}

// ScraperOption configures the advanced scraper
type ScraperOption func(*AdvancedScraper)

// WithProxy sets a proxy for the scraper
func WithProxy(proxyURL string) ScraperOption {
	return func(s *AdvancedScraper) {
		s.proxyURL = proxyURL
		s.configureProxy()
	}
}

// configureProxy configures the HTTP transport to use a proxy
func (a *AdvancedScraper) configureProxy() error {
	if a.proxyURL == "" {
		return nil
	}

	proxyParsed, err := url.Parse(a.proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}

	// Get the current transport
	transport := a.client.Transport.(*http.Transport)

	switch proxyParsed.Scheme {
	case "http", "https":
		// HTTP proxy
		transport.Proxy = http.ProxyURL(proxyParsed)
	case "socks5":
		// SOCKS5 proxy
		dialer, err := proxy.SOCKS5("tcp", proxyParsed.Host, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS5 proxy: %v", err)
		}
		transport.DialContext = dialer.(proxy.ContextDialer).DialContext
	default:
		return fmt.Errorf("unsupported proxy scheme: %s", proxyParsed.Scheme)
	}

	return nil
}

// WithRetryCount sets the number of retries
func WithRetryCount(count int) ScraperOption {
	return func(s *AdvancedScraper) {
		s.retryCount = count
	}
}

// WithRateLimit sets the minimum interval between requests
func WithRateLimit(interval time.Duration) ScraperOption {
	return func(s *AdvancedScraper) {
		s.rateLimiter = NewRateLimiter(interval)
	}
}

// GetWithRetry performs a GET request with retry logic
func (a *AdvancedScraper) GetWithRetry(targetURL string) (*Response, error) {
	var lastErr error

	for i := 0; i <= a.retryCount; i++ {
		// Rate limiting
		a.rateLimiter.Wait()

		// Add random delay to appear more human
		utils.RandomDelay(100, 500)

		response, err := a.Get(targetURL)
		if err == nil && response.StatusCode < 500 {
			// Save cookies
			a.saveCookies(targetURL, response.Headers)
			return response, nil
		}

		lastErr = err
		if i < a.retryCount {
			backoff := time.Duration(i+1) * 2 * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", a.retryCount, lastErr)
}

// PostWithRetry performs a POST request with retry logic
func (a *AdvancedScraper) PostWithRetry(targetURL string, data map[string]string) (*Response, error) {
	formData := url.Values{}
	for key, value := range data {
		formData.Set(key, value)
	}

	var lastErr error

	for i := 0; i <= a.retryCount; i++ {
		a.rateLimiter.Wait()
		utils.RandomDelay(100, 500)

		req, err := http.NewRequest("POST", targetURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		for key, value := range a.headers {
			req.Header.Set(key, value)
		}
		req.Header.Set("User-Agent", a.userAgent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Add cookies
		a.addCookies(req, targetURL)

		resp, err := a.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}

			response := &Response{
				StatusCode: resp.StatusCode,
				Headers:    resp.Header,
				Body:       string(body),
				URL:        targetURL,
			}

			a.saveCookies(targetURL, response.Headers)
			return response, nil
		}

		lastErr = err
		if resp != nil {
			resp.Body.Close()
		}

		if i < a.retryCount {
			backoff := time.Duration(i+1) * 2 * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", a.retryCount, lastErr)
}

// NewCookieJar creates a new cookie jar
func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make(map[string][]*http.Cookie),
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make(chan struct{}, 1),
		minInterval: minInterval,
	}
}

// Wait blocks until it's safe to make another request
func (rl *RateLimiter) Wait() {
	now := time.Now()
	if elapsed := now.Sub(rl.lastRequest); elapsed < rl.minInterval {
		time.Sleep(rl.minInterval - elapsed)
	}
	rl.lastRequest = time.Now()
}

// saveCookies saves cookies from response headers
func (a *AdvancedScraper) saveCookies(targetURL string, headers map[string][]string) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return
	}

	domain := parsedURL.Hostname()

	if setCookies, exists := headers["Set-Cookie"]; exists {
		for _, cookieStr := range setCookies {
			// Parse each Set-Cookie header
			header := http.Header{}
			header.Add("Set-Cookie", cookieStr)
			cookies := (&http.Request{Header: header}).Cookies()

			for _, cookie := range cookies {
				a.cookieJar.cookies[domain] = append(a.cookieJar.cookies[domain], cookie)
			}
		}
	}
}

// SetUserAgent sets a custom user agent
func (a *AdvancedScraper) SetUserAgent(userAgent string) {
	a.userAgent = userAgent
}

// SetHeader sets a custom header
func (a *AdvancedScraper) SetHeader(key, value string) {
	if a.headers == nil {
		a.headers = make(map[string]string)
	}
	a.headers[key] = value
}

// addCookies adds saved cookies to the request
func (a *AdvancedScraper) addCookies(req *http.Request, targetURL string) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return
	}

	domain := parsedURL.Hostname()

	if cookies, exists := a.cookieJar.cookies[domain]; exists {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}
}
