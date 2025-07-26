package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"anti-bot-scraper/internal/utils"

	"golang.org/x/net/proxy"
)

// ProxyRotationMode defines how proxies should be rotated
type ProxyRotationMode int

const (
	RotatePerRequest ProxyRotationMode = iota // Rotate on every request
	RotateOnError                             // Rotate only on timeout/error/block
	HealthAware                               // Use health monitoring for smart rotation
)

// ProxyRotator manages proxy rotation with health monitoring
type ProxyRotator struct {
	proxies       []string
	current       int
	mode          ProxyRotationMode
	mutex         sync.RWMutex
	failedCount   map[string]int      // Track failures per proxy (legacy)
	healthChecker *ProxyHealthChecker // Health monitoring system
	enableHealth  bool                // Whether health monitoring is enabled
	maxRetries    int                 // Max retries for finding healthy proxy
}

// AdvancedScraper extends the basic scraper with additional features
type AdvancedScraper struct {
	*Scraper
	cookieJar         *CookieJar              // Cookie management
	proxyURL          string                  // Single proxy (deprecated, use proxyRotator)
	proxyRotator      *ProxyRotator           // Multi-proxy rotation
	healthChecker     *ProxyHealthChecker     // Proxy health monitoring
	captchaSolver     *CaptchaSolver          // CAPTCHA solving service
	captchaDetector   *CaptchaDetector        // CAPTCHA detection and handling
	behaviorSimulator *HumanBehaviorSimulator // Human behavior simulation
	retryCount        int
	rateLimiter       *RateLimiter
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

// NewAdvancedScraper creates a new advanced scraper (HTTP/1.1 for compatibility)
func NewAdvancedScraper(fingerprint Fingerprint, options ...ScraperOption) (*AdvancedScraper, error) {
	return NewAdvancedScraperWithProtocol(fingerprint, HTTP1_1, options...)
}

// NewAdvancedScraperWithProtocol creates a new advanced scraper with specified protocol
func NewAdvancedScraperWithProtocol(fingerprint Fingerprint, protocol ProtocolVersion, options ...ScraperOption) (*AdvancedScraper, error) {
	return NewAdvancedScraperWithJS(fingerprint, protocol, JSEngineConfig{Enabled: false}, options...)
}

// NewAdvancedScraperWithJS creates a new advanced scraper with JavaScript engine support
func NewAdvancedScraperWithJS(fingerprint Fingerprint, protocol ProtocolVersion, jsConfig JSEngineConfig, options ...ScraperOption) (*AdvancedScraper, error) {
	baseScraper, err := NewScraperWithJS(fingerprint, protocol, jsConfig)
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

// NewProxyRotator creates a new proxy rotator
func NewProxyRotator(proxies []string, mode ProxyRotationMode) *ProxyRotator {
	rotator := &ProxyRotator{
		proxies:     proxies,
		current:     0,
		mode:        mode,
		failedCount: make(map[string]int),
		maxRetries:  len(proxies) * 2, // Allow cycling through all proxies twice
	}

	// Enable health monitoring for HealthAware mode
	if mode == HealthAware {
		rotator.enableHealth = true
		rotator.healthChecker = NewProxyHealthChecker(ProxyHealthConfig{
			CheckInterval: 5 * time.Minute,
			Timeout:       10 * time.Second,
			TestURL:       "https://httpbin.org/ip",
			MaxFailures:   3,
		})

		// Add all proxies to health monitoring
		for _, proxy := range proxies {
			rotator.healthChecker.AddProxy(proxy)
		}

		// Start health monitoring
		rotator.healthChecker.Start()
	}

	return rotator
}

// NewProxyRotatorWithHealthChecker creates a proxy rotator with custom health checker
func NewProxyRotatorWithHealthChecker(proxies []string, mode ProxyRotationMode, healthChecker *ProxyHealthChecker) *ProxyRotator {
	rotator := &ProxyRotator{
		proxies:       proxies,
		current:       0,
		mode:          mode,
		failedCount:   make(map[string]int),
		healthChecker: healthChecker,
		enableHealth:  true,
		maxRetries:    len(proxies) * 2,
	}

	// Add proxies to health checker
	if healthChecker != nil {
		for _, proxy := range proxies {
			healthChecker.AddProxy(proxy)
		}
	}

	return rotator
}

// GetNext returns the next proxy based on rotation mode
func (pr *ProxyRotator) GetNext() string {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if len(pr.proxies) == 0 {
		return ""
	}

	switch pr.mode {
	case RotatePerRequest:
		// Always rotate to next proxy
		proxy := pr.proxies[pr.current]
		pr.current = (pr.current + 1) % len(pr.proxies)
		return proxy

	case HealthAware:
		// Use health monitoring to find best proxy
		return pr.getHealthyProxy()

	default: // RotateOnError
		// Return current proxy
		return pr.proxies[pr.current]
	}
}

// getHealthyProxy finds the best healthy proxy using health monitoring
func (pr *ProxyRotator) getHealthyProxy() string {
	if !pr.enableHealth || pr.healthChecker == nil {
		// Fallback to simple rotation
		proxy := pr.proxies[pr.current]
		pr.current = (pr.current + 1) % len(pr.proxies)
		return proxy
	}

	// Get healthy proxies
	healthyProxies := pr.healthChecker.GetHealthyProxies()

	if len(healthyProxies) == 0 {
		// No healthy proxies available, use any proxy
		proxy := pr.proxies[pr.current]
		pr.current = (pr.current + 1) % len(pr.proxies)
		return proxy
	}

	// Find the best healthy proxy (lowest latency)
	var bestProxy string
	var bestLatency time.Duration = time.Hour // Start with very high value

	for _, proxyURL := range healthyProxies {
		if health, exists := pr.healthChecker.GetProxyHealth(proxyURL); exists {
			if health.Latency < bestLatency {
				bestLatency = health.Latency
				bestProxy = proxyURL
			}
		}
	}

	if bestProxy == "" {
		// Fallback to first healthy proxy
		bestProxy = healthyProxies[0]
	}

	return bestProxy
}

// MarkFailed marks a proxy as failed and rotates to next if needed
func (pr *ProxyRotator) MarkFailed(proxyURL string) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	// Legacy failure tracking
	pr.failedCount[proxyURL]++

	// If health monitoring is enabled, let it handle the failure
	if pr.enableHealth && pr.healthChecker != nil {
		// Health checker will handle the failure internally
		// We just need to trigger a health check
		go pr.healthChecker.CheckProxyHealth(proxyURL)
	}

	// Rotate to next proxy on failure (for non-health-aware modes)
	if pr.mode != HealthAware {
		pr.current = (pr.current + 1) % len(pr.proxies)
	}
}

// MarkSuccess marks a proxy as successful (for health monitoring)
func (pr *ProxyRotator) MarkSuccess(proxyURL string, latency time.Duration) {
	if pr.enableHealth && pr.healthChecker != nil {
		// Record success in health monitoring
		// This will be handled internally by the health checker
	}
}

// GetFailureCount returns the failure count for a proxy
func (pr *ProxyRotator) GetFailureCount(proxyURL string) int {
	pr.mutex.RLock()
	defer pr.mutex.RUnlock()
	return pr.failedCount[proxyURL]
}

// GetHealthyProxies returns all currently healthy proxies
func (pr *ProxyRotator) GetHealthyProxies() []string {
	if pr.enableHealth && pr.healthChecker != nil {
		return pr.healthChecker.GetHealthyProxies()
	}

	// Fallback to all proxies if health monitoring is disabled
	return pr.proxies
}

// GetProxyHealth returns detailed health information for a proxy
func (pr *ProxyRotator) GetProxyHealth(proxyURL string) (*ProxyHealth, bool) {
	if pr.enableHealth && pr.healthChecker != nil {
		return pr.healthChecker.GetProxyHealth(proxyURL)
	}
	return nil, false
}

// GetAllProxiesHealth returns health status of all proxies
func (pr *ProxyRotator) GetAllProxiesHealth() map[string]*ProxyHealth {
	if pr.enableHealth && pr.healthChecker != nil {
		return pr.healthChecker.GetAllProxiesHealth()
	}
	return make(map[string]*ProxyHealth)
}

// GetProxyMetrics returns overall proxy pool metrics
func (pr *ProxyRotator) GetProxyMetrics() map[string]interface{} {
	if pr.enableHealth && pr.healthChecker != nil {
		return pr.healthChecker.GetProxyMetrics()
	}

	// Basic metrics without health monitoring
	return map[string]interface{}{
		"total_proxies":  len(pr.proxies),
		"current_proxy":  pr.current,
		"health_enabled": false,
	}
}

// EnableHealthMonitoring enables health monitoring for existing rotator
func (pr *ProxyRotator) EnableHealthMonitoring(config ProxyHealthConfig) {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if !pr.enableHealth {
		pr.enableHealth = true
		pr.healthChecker = NewProxyHealthChecker(config)

		// Add all proxies to health monitoring
		for _, proxy := range pr.proxies {
			pr.healthChecker.AddProxy(proxy)
		}

		// Start health monitoring
		pr.healthChecker.Start()
	}
}

// DisableHealthMonitoring disables health monitoring
func (pr *ProxyRotator) DisableHealthMonitoring() {
	pr.mutex.Lock()
	defer pr.mutex.Unlock()

	if pr.enableHealth && pr.healthChecker != nil {
		pr.healthChecker.Stop()
		pr.healthChecker = nil
		pr.enableHealth = false
	}
}

// Stop stops the proxy rotator and health monitoring
func (pr *ProxyRotator) Stop() {
	pr.DisableHealthMonitoring()
}

// WithProxyRotation sets up proxy rotation for the scraper
func WithProxyRotation(proxies []string, mode ProxyRotationMode) ScraperOption {
	return func(s *AdvancedScraper) {
		s.proxyRotator = NewProxyRotator(proxies, mode)
	}
}

// WithHealthAwareProxyRotation sets up health-aware proxy rotation
func WithHealthAwareProxyRotation(proxies []string, healthConfig ProxyHealthConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		s.proxyRotator = NewProxyRotator(proxies, HealthAware)
		s.healthChecker = NewProxyHealthChecker(healthConfig)

		// Configure the rotator to use our health checker
		s.proxyRotator.healthChecker = s.healthChecker
		s.proxyRotator.enableHealth = true

		// Add proxies and start monitoring
		for _, proxy := range proxies {
			s.healthChecker.AddProxy(proxy)
		}
		s.healthChecker.Start()
	}
}

// WithProxyHealthMonitoring adds health monitoring to existing proxy setup
func WithProxyHealthMonitoring(config ProxyHealthConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		s.healthChecker = NewProxyHealthChecker(config)

		// If we have a proxy rotator, enable health monitoring on it
		if s.proxyRotator != nil {
			s.proxyRotator.EnableHealthMonitoring(config)
		}
	}
}

// WithCaptchaSolver adds CAPTCHA solving capability
func WithCaptchaSolver(config CaptchaSolverConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		s.captchaSolver = NewCaptchaSolver(config)
	}
}

// WithCaptchaDetection enables automatic CAPTCHA detection and solving
func WithCaptchaDetection(solverConfig CaptchaSolverConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		s.captchaSolver = NewCaptchaSolver(solverConfig)
		// CAPTCHA detector will be initialized when JS engine is available
	}
} // WithProxy sets a proxy for the scraper
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

// WithHumanBehavior enables human behavior simulation
func WithHumanBehavior(config HumanBehaviorConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		s.behaviorSimulator = NewHumanBehaviorSimulator(config)
	}
}

// WithBehaviorType sets a specific behavior type
func WithBehaviorType(behaviorType BehaviorType) ScraperOption {
	return func(s *AdvancedScraper) {
		config := GetDefaultBehaviorConfig()
		if s.behaviorSimulator != nil {
			config = s.behaviorSimulator.config
		}
		config.BehaviorType = behaviorType
		s.behaviorSimulator = NewHumanBehaviorSimulator(config)
		s.behaviorSimulator.ApplyBehaviorType(behaviorType)
	}
}

// EnableBehaviorSimulation enables default human behavior simulation
func EnableBehaviorSimulation() ScraperOption {
	return func(s *AdvancedScraper) {
		config := GetDefaultBehaviorConfig()
		s.behaviorSimulator = NewHumanBehaviorSimulator(config)
	}
}

// GetWithRetry performs a GET request with retry logic
func (a *AdvancedScraper) GetWithRetry(targetURL string) (*Response, error) {
	var lastErr error

	for i := 0; i <= a.retryCount; i++ {
		// Configure proxy rotation if enabled
		if a.proxyRotator != nil {
			currentProxy := a.proxyRotator.GetNext()
			if currentProxy != "" {
				a.proxyURL = currentProxy
				if err := a.configureProxy(); err != nil {
					// Mark proxy as failed and continue with next
					a.proxyRotator.MarkFailed(currentProxy)
					continue
				}
			}
		}

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

		// Mark proxy as failed if we're using rotation and got an error
		if a.proxyRotator != nil && a.proxyURL != "" {
			a.proxyRotator.MarkFailed(a.proxyURL)
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
		// Configure proxy rotation if enabled
		if a.proxyRotator != nil {
			currentProxy := a.proxyRotator.GetNext()
			if currentProxy != "" {
				a.proxyURL = currentProxy
				if err := a.configureProxy(); err != nil {
					// Mark proxy as failed and continue with next
					a.proxyRotator.MarkFailed(currentProxy)
					continue
				}
			}
		}

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

		// Mark proxy as failed if we're using rotation and got an error
		if a.proxyRotator != nil && a.proxyURL != "" {
			a.proxyRotator.MarkFailed(a.proxyURL)
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

// GetProxyHealth returns health information for a specific proxy
func (a *AdvancedScraper) GetProxyHealth(proxyURL string) (*ProxyHealth, bool) {
	if a.healthChecker != nil {
		return a.healthChecker.GetProxyHealth(proxyURL)
	}
	if a.proxyRotator != nil {
		return a.proxyRotator.GetProxyHealth(proxyURL)
	}
	return nil, false
}

// GetAllProxiesHealth returns health status of all proxies
func (a *AdvancedScraper) GetAllProxiesHealth() map[string]*ProxyHealth {
	if a.healthChecker != nil {
		return a.healthChecker.GetAllProxiesHealth()
	}
	if a.proxyRotator != nil {
		return a.proxyRotator.GetAllProxiesHealth()
	}
	return make(map[string]*ProxyHealth)
}

// GetProxyMetrics returns comprehensive proxy pool metrics
func (a *AdvancedScraper) GetProxyMetrics() map[string]interface{} {
	if a.healthChecker != nil {
		return a.healthChecker.GetProxyMetrics()
	}
	if a.proxyRotator != nil {
		return a.proxyRotator.GetProxyMetrics()
	}
	return map[string]interface{}{
		"proxy_rotation":    false,
		"health_monitoring": false,
	}
}

// GetHealthyProxies returns list of currently healthy proxies
func (a *AdvancedScraper) GetHealthyProxies() []string {
	if a.healthChecker != nil {
		return a.healthChecker.GetHealthyProxies()
	}
	if a.proxyRotator != nil {
		return a.proxyRotator.GetHealthyProxies()
	}
	return []string{}
}

// EnableProxy manually enables a disabled proxy
func (a *AdvancedScraper) EnableProxy(proxyURL string) {
	if a.healthChecker != nil {
		a.healthChecker.EnableProxy(proxyURL)
	}
}

// DisableProxy manually disables a proxy
func (a *AdvancedScraper) DisableProxy(proxyURL string) {
	if a.healthChecker != nil {
		a.healthChecker.DisableProxy(proxyURL)
	}
}

// CheckProxyHealth manually triggers a health check for a specific proxy
func (a *AdvancedScraper) CheckProxyHealth(proxyURL string) error {
	if a.healthChecker != nil {
		return a.healthChecker.CheckProxyHealth(proxyURL)
	}
	return fmt.Errorf("health monitoring not enabled")
}

// Stop gracefully stops the scraper and health monitoring
func (a *AdvancedScraper) Stop() {
	if a.proxyRotator != nil {
		a.proxyRotator.Stop()
	}
	if a.healthChecker != nil {
		a.healthChecker.Stop()
	}
}

// DetectAndSolveCaptcha detects and solves CAPTCHAs on the current page
func (a *AdvancedScraper) DetectAndSolveCaptcha(ctx context.Context, pageURL string) (*CaptchaDetectionResult, error) {
	if a.captchaDetector == nil {
		return &CaptchaDetectionResult{Found: false, PageURL: pageURL}, nil
	}
	return a.captchaDetector.DetectCaptcha(ctx, pageURL)
}

// SetCaptchaSolver sets or updates the CAPTCHA solver
func (a *AdvancedScraper) SetCaptchaSolver(solver *CaptchaSolver) {
	a.captchaSolver = solver
	if a.captchaDetector != nil {
		a.captchaDetector.SetSolver(solver)
	}
}

// EnableCaptchaDetection enables CAPTCHA detection (requires JS engine)
func (a *AdvancedScraper) EnableCaptchaDetection() error {
	if a.captchaSolver == nil {
		return fmt.Errorf("CAPTCHA solver not configured")
	}

	// Get JS engine from the scraper
	jsEngine := a.getJSEngine()
	if jsEngine == nil {
		return fmt.Errorf("JavaScript engine required for CAPTCHA detection")
	}

	a.captchaDetector = NewCaptchaDetector(a.captchaSolver, jsEngine)
	return nil
}

// DisableCaptchaDetection disables CAPTCHA detection
func (a *AdvancedScraper) DisableCaptchaDetection() {
	if a.captchaDetector != nil {
		a.captchaDetector.Disable()
	}
}

// GetCaptchaSolverBalance returns the balance of the CAPTCHA solving service
func (a *AdvancedScraper) GetCaptchaSolverBalance(ctx context.Context) (float64, error) {
	if a.captchaSolver == nil {
		return 0, fmt.Errorf("CAPTCHA solver not configured")
	}
	return a.captchaSolver.GetBalance(ctx)
}

// IsCaptchaDetectionEnabled returns whether CAPTCHA detection is enabled
func (a *AdvancedScraper) IsCaptchaDetectionEnabled() bool {
	return a.captchaDetector != nil && a.captchaDetector.IsEnabled()
}

// getJSEngine attempts to get the JavaScript engine from the scraper
func (a *AdvancedScraper) getJSEngine() *JSEngine {
	// This is a simplified approach - in a real implementation,
	// we'd need to access the JS engine from the scraper properly
	// For now, return nil if not available
	return nil
}
