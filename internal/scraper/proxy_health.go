package scraper

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ProxyHealth represents the health status of a proxy
type ProxyHealth struct {
	URL          string        `json:"url"`
	IsHealthy    bool          `json:"is_healthy"`
	LastCheck    time.Time     `json:"last_check"`
	Latency      time.Duration `json:"latency"`
	FailureCount int           `json:"failure_count"`
	SuccessCount int           `json:"success_count"`
	Uptime       float64       `json:"uptime"` // Success rate percentage
	LastError    string        `json:"last_error"`
	Status       string        `json:"status"` // "active", "degraded", "failed", "disabled"
	Region       string        `json:"region"` // Geographic region if known
}

// ProxyHealthChecker manages proxy health monitoring
type ProxyHealthChecker struct {
	proxies        map[string]*ProxyHealth
	mutex          sync.RWMutex
	checkInterval  time.Duration
	timeout        time.Duration
	testURL        string
	stopChan       chan struct{}
	running        bool
	maxFailures    int                // Max failures before marking proxy as failed
	checkDelay     time.Duration      // Delay between individual proxy checks
	healthCallback func(*ProxyHealth) // Callback when proxy health changes
}

// ProxyHealthConfig configures the health checker
type ProxyHealthConfig struct {
	CheckInterval  time.Duration      // How often to check all proxies
	Timeout        time.Duration      // Timeout for individual health checks
	TestURL        string             // URL to test proxy connectivity
	MaxFailures    int                // Max consecutive failures before disabling proxy
	CheckDelay     time.Duration      // Delay between checking individual proxies
	HealthCallback func(*ProxyHealth) // Optional callback for health changes
}

// NewProxyHealthChecker creates a new proxy health checker
func NewProxyHealthChecker(config ProxyHealthConfig) *ProxyHealthChecker {
	if config.CheckInterval == 0 {
		config.CheckInterval = 5 * time.Minute
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.TestURL == "" {
		config.TestURL = "https://httpbin.org/ip"
	}
	if config.MaxFailures == 0 {
		config.MaxFailures = 3
	}
	if config.CheckDelay == 0 {
		config.CheckDelay = 1 * time.Second
	}

	return &ProxyHealthChecker{
		proxies:        make(map[string]*ProxyHealth),
		checkInterval:  config.CheckInterval,
		timeout:        config.Timeout,
		testURL:        config.TestURL,
		maxFailures:    config.MaxFailures,
		checkDelay:     config.CheckDelay,
		healthCallback: config.HealthCallback,
		stopChan:       make(chan struct{}),
	}
}

// AddProxy adds a proxy to the health monitoring system
func (phc *ProxyHealthChecker) AddProxy(proxyURL string) {
	phc.mutex.Lock()
	defer phc.mutex.Unlock()

	if _, exists := phc.proxies[proxyURL]; !exists {
		phc.proxies[proxyURL] = &ProxyHealth{
			URL:       proxyURL,
			IsHealthy: true,
			LastCheck: time.Now(),
			Status:    "active",
		}
	}
}

// RemoveProxy removes a proxy from health monitoring
func (phc *ProxyHealthChecker) RemoveProxy(proxyURL string) {
	phc.mutex.Lock()
	defer phc.mutex.Unlock()
	delete(phc.proxies, proxyURL)
}

// GetProxyHealth returns the health status of a specific proxy
func (phc *ProxyHealthChecker) GetProxyHealth(proxyURL string) (*ProxyHealth, bool) {
	phc.mutex.RLock()
	defer phc.mutex.RUnlock()
	health, exists := phc.proxies[proxyURL]
	return health, exists
}

// GetHealthyProxies returns a list of currently healthy proxies
func (phc *ProxyHealthChecker) GetHealthyProxies() []string {
	phc.mutex.RLock()
	defer phc.mutex.RUnlock()

	var healthy []string
	for url, health := range phc.proxies {
		if health.IsHealthy && health.Status == "active" {
			healthy = append(healthy, url)
		}
	}
	return healthy
}

// GetAllProxiesHealth returns health status of all proxies
func (phc *ProxyHealthChecker) GetAllProxiesHealth() map[string]*ProxyHealth {
	phc.mutex.RLock()
	defer phc.mutex.RUnlock()

	result := make(map[string]*ProxyHealth)
	for url, health := range phc.proxies {
		// Create a copy to avoid race conditions
		healthCopy := *health
		result[url] = &healthCopy
	}
	return result
}

// CheckProxyHealth performs a health check on a specific proxy
func (phc *ProxyHealthChecker) CheckProxyHealth(proxyURL string) error {
	start := time.Now()

	// Parse proxy URL
	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		phc.recordFailure(proxyURL, fmt.Sprintf("Invalid proxy URL: %v", err))
		return err
	}

	// Create HTTP client with proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
		DialContext: (&net.Dialer{
			Timeout: phc.timeout,
		}).DialContext,
		TLSHandshakeTimeout: phc.timeout,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   phc.timeout,
	}

	// Create request with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), phc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", phc.testURL, nil)
	if err != nil {
		phc.recordFailure(proxyURL, fmt.Sprintf("Failed to create request: %v", err))
		return err
	}

	// Set User-Agent to avoid detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		phc.recordFailure(proxyURL, fmt.Sprintf("Request failed: %v", err))
		return err
	}
	defer resp.Body.Close()

	latency := time.Since(start)

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		phc.recordSuccess(proxyURL, latency)
		return nil
	}

	errorMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
	phc.recordFailure(proxyURL, errorMsg)
	return fmt.Errorf("proxy returned status %d", resp.StatusCode)
}

// recordSuccess records a successful health check
func (phc *ProxyHealthChecker) recordSuccess(proxyURL string, latency time.Duration) {
	phc.mutex.Lock()
	defer phc.mutex.Unlock()

	health := phc.proxies[proxyURL]
	if health == nil {
		return
	}

	health.LastCheck = time.Now()
	health.Latency = latency
	health.SuccessCount++
	health.LastError = ""

	// Calculate uptime percentage
	total := health.SuccessCount + health.FailureCount
	if total > 0 {
		health.Uptime = float64(health.SuccessCount) / float64(total) * 100
	}

	// Determine status based on performance
	if health.Latency > 10*time.Second {
		health.Status = "degraded"
	} else {
		health.Status = "active"
	}

	// Mark as healthy if it was previously failed
	if !health.IsHealthy {
		health.IsHealthy = true
		if phc.healthCallback != nil {
			go phc.healthCallback(health)
		}
	}
}

// recordFailure records a failed health check
func (phc *ProxyHealthChecker) recordFailure(proxyURL string, errorMsg string) {
	phc.mutex.Lock()
	defer phc.mutex.Unlock()

	health := phc.proxies[proxyURL]
	if health == nil {
		return
	}

	health.LastCheck = time.Now()
	health.FailureCount++
	health.LastError = errorMsg

	// Calculate uptime percentage
	total := health.SuccessCount + health.FailureCount
	if total > 0 {
		health.Uptime = float64(health.SuccessCount) / float64(total) * 100
	}

	// Check if proxy should be marked as failed
	consecutiveFailures := health.FailureCount - health.SuccessCount
	if consecutiveFailures >= phc.maxFailures {
		health.IsHealthy = false
		health.Status = "failed"

		if phc.healthCallback != nil {
			go phc.healthCallback(health)
		}
	} else {
		health.Status = "degraded"
	}
}

// Start begins the health monitoring process
func (phc *ProxyHealthChecker) Start() {
	phc.mutex.Lock()
	if phc.running {
		phc.mutex.Unlock()
		return
	}
	phc.running = true
	phc.mutex.Unlock()

	go phc.healthCheckLoop()
}

// Stop stops the health monitoring process
func (phc *ProxyHealthChecker) Stop() {
	phc.mutex.Lock()
	if !phc.running {
		phc.mutex.Unlock()
		return
	}
	phc.running = false
	phc.mutex.Unlock()

	close(phc.stopChan)
}

// healthCheckLoop runs the continuous health checking
func (phc *ProxyHealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(phc.checkInterval)
	defer ticker.Stop()

	// Run initial health check
	phc.runHealthChecks()

	for {
		select {
		case <-ticker.C:
			phc.runHealthChecks()
		case <-phc.stopChan:
			return
		}
	}
}

// runHealthChecks performs health checks on all proxies
func (phc *ProxyHealthChecker) runHealthChecks() {
	phc.mutex.RLock()
	proxies := make([]string, 0, len(phc.proxies))
	for url := range phc.proxies {
		proxies = append(proxies, url)
	}
	phc.mutex.RUnlock()

	// Check each proxy with a delay to avoid overwhelming
	for _, proxyURL := range proxies {
		go func(url string) {
			phc.CheckProxyHealth(url)
		}(proxyURL)

		// Small delay between checks to avoid overwhelming the system
		time.Sleep(phc.checkDelay)
	}
}

// GetProxyMetrics returns overall proxy pool metrics
func (phc *ProxyHealthChecker) GetProxyMetrics() map[string]interface{} {
	phc.mutex.RLock()
	defer phc.mutex.RUnlock()

	totalProxies := len(phc.proxies)
	healthyProxies := 0
	degradedProxies := 0
	failedProxies := 0
	avgLatency := time.Duration(0)
	avgUptime := 0.0

	for _, health := range phc.proxies {
		switch health.Status {
		case "active":
			healthyProxies++
			avgLatency += health.Latency
		case "degraded":
			degradedProxies++
			avgLatency += health.Latency
		case "failed":
			failedProxies++
		}
		avgUptime += health.Uptime
	}

	if healthyProxies+degradedProxies > 0 {
		avgLatency = avgLatency / time.Duration(healthyProxies+degradedProxies)
	}

	if totalProxies > 0 {
		avgUptime = avgUptime / float64(totalProxies)
	}

	return map[string]interface{}{
		"total_proxies":    totalProxies,
		"healthy_proxies":  healthyProxies,
		"degraded_proxies": degradedProxies,
		"failed_proxies":   failedProxies,
		"avg_latency_ms":   avgLatency.Milliseconds(),
		"avg_uptime":       avgUptime,
		"last_check":       time.Now(),
	}
}

// EnableProxy manually enables a disabled proxy
func (phc *ProxyHealthChecker) EnableProxy(proxyURL string) {
	phc.mutex.Lock()
	defer phc.mutex.Unlock()

	if health, exists := phc.proxies[proxyURL]; exists {
		health.IsHealthy = true
		health.Status = "active"
		health.FailureCount = 0 // Reset failure count
	}
}

// DisableProxy manually disables a proxy
func (phc *ProxyHealthChecker) DisableProxy(proxyURL string) {
	phc.mutex.Lock()
	defer phc.mutex.Unlock()

	if health, exists := phc.proxies[proxyURL]; exists {
		health.IsHealthy = false
		health.Status = "disabled"
	}
}
