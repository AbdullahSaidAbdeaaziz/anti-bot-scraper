package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"anti-bot-scraper/internal/scraper"
)

var (
	// Core Input Configuration Flags
	url            = flag.String("url", "", "Single URL to scrape (required if urls-file not specified)")
	urlsFile       = flag.String("urls-file", "", "File containing multiple URLs to scrape (one per line)")
	numRequests    = flag.Int("num-requests", 1, "Number of requests to send per URL")
	proxyFile      = flag.String("proxy-file", "", "File containing proxy list (one per line)")
	tlsProfile     = flag.String("tls-profile", "chrome", "TLS profile for fingerprinting (chrome, firefox, safari, edge)")
	tlsRandomize   = flag.Bool("tls-randomize", false, "Randomize TLS profiles across requests")
	delayMin       = flag.Duration("delay-min", 1*time.Second, "Minimum delay between requests")
	delayMax       = flag.Duration("delay-max", 3*time.Second, "Maximum delay between requests")
	delayRandomize = flag.Bool("delay-randomize", true, "Randomize delays within the specified range")

	// Existing flags
	browser       = flag.String("browser", "chrome", "Browser fingerprint (chrome, firefox, safari, edge)")
	method        = flag.String("method", "GET", "HTTP method (GET, POST)")
	headers       = flag.String("headers", "", "Custom headers in JSON format or @filename to read from file")
	data          = flag.String("data", "", "POST data in JSON format or @filename to read from file")
	output        = flag.String("output", "text", "Output format (text, json)")
	retries       = flag.Int("retries", 3, "Number of retries")
	rateLimit     = flag.Duration("rate-limit", 1*time.Second, "Rate limit between requests")
	proxy         = flag.String("proxy", "", "Single proxy URL (http://proxy:port or socks5://proxy:port)")
	proxies       = flag.String("proxies", "", "Multiple proxies separated by comma for rotation")
	proxyRotation = flag.String("proxy-rotation", "per-request", "Proxy rotation mode: 'per-request', 'on-error', or 'health-aware'")

	// Enhanced Header Mimicry Flags
	headerMimicry    = flag.Bool("header-mimicry", true, "Enable browser-consistent header mimicry")
	customUserAgent  = flag.String("custom-user-agent", "", "Custom User-Agent (overrides browser default)")
	headerProfile    = flag.String("header-profile", "auto", "Header profile: 'auto' (match TLS), 'chrome', 'firefox', 'safari', 'edge'")
	enableSecHeaders = flag.Bool("enable-sec-headers", true, "Include security headers (Sec-CH-UA, Sec-Fetch-*)")
	acceptLanguage   = flag.String("accept-language", "auto", "Accept-Language header value ('auto' for browser default)")
	acceptEncoding   = flag.String("accept-encoding", "auto", "Accept-Encoding header value ('auto' for browser default)")

	// Cookie & Redirect Handling Flags
	cookieJar         = flag.Bool("cookie-jar", true, "Enable in-memory cookie jar")
	cookiePersistence = flag.String("cookie-persistence", "session", "Cookie persistence: 'session', 'proxy', 'none'")
	followRedirects   = flag.Bool("follow-redirects", true, "Follow HTTP redirects (302, 301, etc.)")
	maxRedirects      = flag.Int("max-redirects", 10, "Maximum number of redirects to follow")
	redirectTimeout   = flag.Duration("redirect-timeout", 30*time.Second, "Timeout for redirect chains")
	cookieFile        = flag.String("cookie-file", "", "File to save/load cookies for persistence")
	clearCookies      = flag.Bool("clear-cookies", false, "Clear cookies before each request")
	// Proxy Health Monitoring Flags
	enableProxyHealth   = flag.Bool("enable-proxy-health", false, "Enable proxy health monitoring")
	proxyHealthInterval = flag.Duration("proxy-health-interval", 5*time.Minute, "Proxy health check interval")
	proxyHealthTimeout  = flag.Duration("proxy-health-timeout", 10*time.Second, "Proxy health check timeout")
	proxyHealthTestURL  = flag.String("proxy-health-test-url", "https://httpbin.org/ip", "URL to test proxy health")
	proxyMaxFailures    = flag.Int("proxy-max-failures", 3, "Max consecutive failures before disabling proxy")
	showProxyMetrics    = flag.Bool("show-proxy-metrics", false, "Show proxy health metrics after request")
	// CAPTCHA Solving Flags
	enableCaptcha       = flag.Bool("enable-captcha", false, "Enable automatic CAPTCHA detection and solving")
	captchaService      = flag.String("captcha-service", "2captcha", "CAPTCHA solving service: '2captcha', 'deathbycaptcha', 'anticaptcha'")
	captchaAPIKey       = flag.String("captcha-api-key", "", "API key for CAPTCHA solving service")
	captchaTimeout      = flag.Duration("captcha-timeout", 5*time.Minute, "CAPTCHA solving timeout")
	captchaPollInterval = flag.Duration("captcha-poll-interval", 5*time.Second, "CAPTCHA solution polling interval")
	captchaMaxRetries   = flag.Int("captcha-max-retries", 3, "Max retries for CAPTCHA solving")
	captchaMinScore     = flag.Float64("captcha-min-score", 0.3, "Minimum score for reCAPTCHA v3 (0.1-0.9)")
	showCaptchaInfo     = flag.Bool("show-captcha-info", false, "Show CAPTCHA detection and solving information")
	// Human Behavior Simulation Flags
	enableBehavior       = flag.Bool("enable-behavior", false, "Enable human behavior simulation")
	behaviorType         = flag.String("behavior-type", "normal", "Behavior type: 'normal', 'cautious', 'aggressive', 'random'")
	behaviorMinDelay     = flag.Duration("behavior-min-delay", 500*time.Millisecond, "Minimum delay between actions")
	behaviorMaxDelay     = flag.Duration("behavior-max-delay", 2*time.Second, "Maximum delay between actions")
	enableMouseMove      = flag.Bool("enable-mouse-movement", true, "Enable realistic mouse movement simulation")
	enableScrollSim      = flag.Bool("enable-scroll-simulation", true, "Enable realistic scrolling simulation")
	enableTypingDelay    = flag.Bool("enable-typing-delay", true, "Enable realistic typing delays")
	enableRandomActivity = flag.Bool("enable-random-activity", false, "Enable random page interactions")
	showBehaviorInfo     = flag.Bool("show-behavior-info", false, "Show behavior simulation information")
	// Performance Optimization Flags
	enableConcurrent     = flag.Bool("enable-concurrent", false, "Enable concurrent request processing")
	maxConcurrent        = flag.Int("max-concurrent", 10, "Maximum concurrent requests")
	workerPoolSize       = flag.Int("worker-pool-size", 5, "Number of worker goroutines")
	requestsPerSecond    = flag.Float64("requests-per-second", 5.0, "Rate limit in requests per second")
	connectionPoolSize   = flag.Int("connection-pool-size", 20, "HTTP connection pool size")
	maxIdleConns         = flag.Int("max-idle-conns", 10, "Maximum idle connections")
	idleConnTimeout      = flag.Duration("idle-conn-timeout", 90*time.Second, "Idle connection timeout")
	queueSize            = flag.Int("queue-size", 1000, "Request queue size")
	showPerformanceStats = flag.Bool("show-performance-stats", false, "Show performance statistics")
	// Memory Optimization Flags
	enableMemoryOpt        = flag.Bool("enable-memory-optimization", false, "Enable memory optimization")
	maxMemoryMB            = flag.Int64("max-memory-mb", 512, "Maximum memory usage in MB")
	enableIntelligentQueue = flag.Bool("enable-intelligent-queue", false, "Enable priority-based intelligent queueing")
	showMemoryStats        = flag.Bool("show-memory-stats", false, "Show memory usage statistics")
	httpVersion            = flag.String("http-version", "1.1", "HTTP version: '1.1', '2', or 'auto'")
	userAgent              = flag.String("user-agent", "", "Custom User-Agent (overrides browser default)")
	timeout                = flag.Duration("timeout", 30*time.Second, "Request timeout")
	verbose                = flag.Bool("verbose", false, "Verbose output")
	showHeaders            = flag.Bool("show-headers", false, "Show response headers")
	followRedirect         = flag.Bool("follow-redirect", true, "Follow redirects")
	version                = flag.Bool("version", false, "Show version information")
	// JavaScript Engine Flags
	enableJS       = flag.Bool("enable-js", false, "Enable JavaScript engine for dynamic content")
	jsTimeout      = flag.Duration("js-timeout", 30*time.Second, "JavaScript execution timeout")
	jsMode         = flag.String("js-mode", "standard", "JavaScript mode: 'standard', 'behavior', 'wait-element'")
	jsWaitSelector = flag.String("js-wait-selector", "", "CSS selector to wait for (requires js-mode=wait-element)")
	jsCode         = flag.String("js-code", "", "Custom JavaScript code to execute")
	headless       = flag.Bool("headless", true, "Run browser in headless mode")
	viewport       = flag.String("viewport", "1920x1080", "Browser viewport size (WIDTHxHEIGHT)")
	noImages       = flag.Bool("no-images", false, "Disable image loading for faster execution")
)

const (
	appName    = "Anti-Bot Scraper"
	appVersion = "1.0.0"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s v%s\n", appName, appVersion)
		fmt.Fprintf(os.Stderr, "A TLS fingerprinting web scraper that mimics browser behavior\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -url https://httpbin.org/headers\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url https://httpbin.org/post -method POST -data '{\"key\":\"value\"}'\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url https://example.com -browser firefox -verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url https://httpbin.org/headers -output json -show-headers\n", os.Args[0])
	}

	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	// Validate input configuration
	if *url == "" && *urlsFile == "" {
		fmt.Fprintf(os.Stderr, "Error: Either -url or -urls-file is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *url != "" && *urlsFile != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify both -url and -urls-file\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Prepare list of URLs to process
	var urls []string
	if *url != "" {
		urls = []string{*url}
	} else {
		var err error
		urls, err = loadURLsFromFile(*urlsFile)
		if err != nil {
			log.Fatal("Failed to load URLs from file:", err)
		}
	}

	if len(urls) == 0 {
		log.Fatal("No URLs to process")
	}

	if *verbose {
		log.Printf("Starting %s v%s", appName, appVersion)
		log.Printf("URLs to process: %d", len(urls))
		log.Printf("Requests per URL: %d", *numRequests)
		log.Printf("Browser: %s", *browser)
		log.Printf("Method: %s", *method)
		if *tlsRandomize {
			log.Printf("TLS Profile: randomized across requests")
		} else {
			log.Printf("TLS Profile: %s", *tlsProfile)
		}
		if *delayRandomize {
			log.Printf("Request delays: randomized between %v and %v", *delayMin, *delayMax)
		} else {
			log.Printf("Request delay: %v", *delayMin)
		}
	}

	// Load proxy list if specified
	var proxyList []string
	if *proxyFile != "" {
		var err error
		proxyList, err = loadProxiesFromFile(*proxyFile)
		if err != nil {
			log.Fatal("Failed to load proxies from file:", err)
		}
		if *verbose {
			log.Printf("Loaded %d proxies from file", len(proxyList))
		}
	} else if *proxies != "" {
		proxyList = strings.Split(*proxies, ",")
		for i, p := range proxyList {
			proxyList[i] = strings.TrimSpace(p)
		}
	}

	// Execute requests for all URLs
	for urlIndex, targetURL := range urls {
		if len(urls) > 1 && *verbose {
			log.Printf("Processing URL %d/%d: %s", urlIndex+1, len(urls), targetURL)
		}

		// Execute multiple requests per URL if specified
		for reqIndex := 0; reqIndex < *numRequests; reqIndex++ {
			if *numRequests > 1 && *verbose {
				log.Printf("Request %d/%d for URL: %s", reqIndex+1, *numRequests, targetURL)
			}

			err := executeRequest(targetURL, reqIndex, urlIndex, proxyList)
			if err != nil {
				log.Printf("Request failed for %s (attempt %d): %v", targetURL, reqIndex+1, err)
				continue
			}

			// Apply delay between requests (except for the last request)
			if reqIndex < *numRequests-1 || urlIndex < len(urls)-1 {
				delay := calculateDelay()
				if *verbose {
					log.Printf("Waiting %v before next request", delay)
				}
				time.Sleep(delay)
			}
		}
	}
}

func loadURLsFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read URLs file %s: %w", filename, err)
	}

	lines := strings.Split(string(content), "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}

	return urls, nil
}

func loadProxiesFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy file %s: %w", filename, err)
	}

	lines := strings.Split(string(content), "\n")
	var proxies []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			proxies = append(proxies, line)
		}
	}

	return proxies, nil
}

func calculateDelay() time.Duration {
	if !*delayRandomize {
		return *delayMin
	}

	if *delayMin >= *delayMax {
		return *delayMin
	}

	minNanos := delayMin.Nanoseconds()
	maxNanos := delayMax.Nanoseconds()
	randomNanos := minNanos + rand.Int63n(maxNanos-minNanos)
	return time.Duration(randomNanos)
}

func selectTLSProfile() scraper.Fingerprint {
	if !*tlsRandomize {
		fingerprint, _ := parseBrowserFingerprint(*tlsProfile)
		return fingerprint
	}

	profiles := []string{"chrome", "firefox", "safari", "edge"}
	randomProfile := profiles[rand.Intn(len(profiles))]
	fingerprint, _ := parseBrowserFingerprint(randomProfile)
	return fingerprint
}

func executeRequest(targetURL string, reqIndex, urlIndex int, proxyList []string) error {
	// Select TLS fingerprint (randomized or fixed)
	fingerprint := selectTLSProfile()

	// Create scraper options
	options := []scraper.ScraperOption{
		scraper.WithRetryCount(*retries),
		scraper.WithRateLimit(*rateLimit),
		scraper.WithTimeout(*timeout),
	}

	// Add enhanced header mimicry configuration
	if *headerMimicry {
		var headerProfileType scraper.Fingerprint
		if *headerProfile == "auto" {
			headerProfileType = fingerprint // Match TLS profile
		} else {
			var err error
			headerProfileType, err = parseBrowserFingerprint(*headerProfile)
			if err != nil {
				return fmt.Errorf("invalid header profile: %w", err)
			}
		}

		headerConfig := scraper.HeaderMimicryConfig{
			Profile:           headerProfileType,
			IncludeSecHeaders: *enableSecHeaders,
			AcceptLanguage:    *acceptLanguage,
			AcceptEncoding:    *acceptEncoding,
			CustomUserAgent:   *customUserAgent,
		}

		options = append(options, scraper.WithHeaderMimicry(headerConfig))

		if *verbose {
			log.Printf("Header mimicry enabled for profile: %s", getProfileName(headerProfileType))
		}
	}

	// Add cookie and redirect configuration
	cookieConfig := scraper.CookieConfig{
		EnableJar:      *cookieJar,
		Persistence:    parseCookiePersistence(*cookiePersistence),
		ClearOnRequest: *clearCookies,
		CookieFile:     *cookieFile,
	}

	redirectConfig := scraper.RedirectConfig{
		FollowRedirects: *followRedirects,
		MaxRedirects:    *maxRedirects,
		Timeout:         *redirectTimeout,
	}

	options = append(options, scraper.WithCookieHandling(cookieConfig))
	options = append(options, scraper.WithRedirectHandling(redirectConfig))

	// Add proxy configuration if specified
	if len(proxyList) > 0 {
		// Select proxy for this request (round-robin or random)
		selectedProxy := selectProxy(proxyList, reqIndex, urlIndex)

		if *verbose {
			log.Printf("Using proxy: %s", selectedProxy)
		}

		options = append(options, scraper.WithProxy(selectedProxy))
	} else if *proxy != "" {
		options = append(options, scraper.WithProxy(*proxy))
	}

	// Parse HTTP version
	var protocol scraper.ProtocolVersion
	switch *httpVersion {
	case "1.1":
		protocol = scraper.HTTP1_1
	case "2":
		protocol = scraper.HTTP2
	case "auto":
		protocol = scraper.HTTPAuto
	default:
		return fmt.Errorf("invalid HTTP version: %s", *httpVersion)
	}

	// Create scraper
	s, err := scraper.NewAdvancedScraperWithProtocol(fingerprint, protocol, options...)
	if err != nil {
		return fmt.Errorf("failed to create scraper: %w", err)
	}
	defer s.Close()

	// Override user agent if provided
	if *customUserAgent != "" {
		s.SetUserAgent(*customUserAgent)
	}

	// Add custom headers if provided
	if *headers != "" {
		customHeaders, err := parseHeaders(*headers)
		if err != nil {
			return fmt.Errorf("failed to parse headers: %w", err)
		}
		for key, value := range customHeaders {
			s.SetHeader(key, value)
		}
	}

	// Execute request based on method
	var response *scraper.Response
	switch strings.ToUpper(*method) {
	case "GET":
		if *verbose {
			log.Printf("Making GET request to %s", targetURL)
		}
		response, err = s.GetWithRetry(targetURL)
	case "POST":
		if *data == "" {
			return fmt.Errorf("POST data is required for POST requests")
		}
		postData, err := parsePostData(*data)
		if err != nil {
			return fmt.Errorf("failed to parse POST data: %w", err)
		}
		if *verbose {
			log.Printf("Making POST request to %s", targetURL)
		}
		response, err = s.PostWithRetry(targetURL, postData)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", *method)
	}

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// Output results
	outputResponse(response, *output, *showHeaders, *verbose)

	if *verbose {
		log.Printf("Request completed successfully for %s", targetURL)
	}

	return nil
}

func parseCookiePersistence(persistence string) scraper.CookiePersistence {
	switch persistence {
	case "session":
		return scraper.SessionPersistence
	case "proxy":
		return scraper.ProxyPersistence
	case "none":
		return scraper.NoPersistence
	default:
		log.Printf("Invalid cookie persistence mode: %s, using session", persistence)
		return scraper.SessionPersistence
	}
}

func getProfileName(fingerprint scraper.Fingerprint) string {
	switch fingerprint {
	case scraper.ChromeFingerprint:
		return "chrome"
	case scraper.FirefoxFingerprint:
		return "firefox"
	case scraper.SafariFingerprint:
		return "safari"
	case scraper.EdgeFingerprint:
		return "edge"
	default:
		return "unknown"
	}
}

func selectProxy(proxyList []string, reqIndex, urlIndex int) string {
	if len(proxyList) == 0 {
		return ""
	}

	// Use round-robin selection based on total request count
	totalRequests := urlIndex*(*numRequests) + reqIndex
	return proxyList[totalRequests%len(proxyList)]
}

func parseBrowserFingerprint(browser string) (scraper.Fingerprint, error) {
	switch strings.ToLower(browser) {
	case "chrome":
		return scraper.ChromeFingerprint, nil
	case "firefox":
		return scraper.FirefoxFingerprint, nil
	case "safari":
		return scraper.SafariFingerprint, nil
	case "edge":
		return scraper.EdgeFingerprint, nil
	default:
		return scraper.ChromeFingerprint, fmt.Errorf("unsupported browser: %s", browser)
	}
}

func parseHeaders(headersJSON string) (map[string]string, error) {
	var jsonData string

	// Check if it's a file reference
	if strings.HasPrefix(headersJSON, "@") {
		filename := headersJSON[1:]
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		jsonData = string(content)
	} else {
		jsonData = headersJSON
	}

	var headers map[string]string
	err := json.Unmarshal([]byte(jsonData), &headers)
	return headers, err
}

func parsePostData(dataJSON string) (map[string]string, error) {
	var jsonData string

	// Check if it's a file reference
	if strings.HasPrefix(dataJSON, "@") {
		filename := dataJSON[1:]
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		jsonData = string(content)
	} else {
		jsonData = dataJSON
	}

	var data map[string]string
	err := json.Unmarshal([]byte(jsonData), &data)
	return data, err
}

func outputResponse(response *scraper.Response, outputFormat string, showHeaders, verbose bool) {
	switch strings.ToLower(outputFormat) {
	case "json":
		outputJSON(response, showHeaders)
	case "text":
		outputText(response, showHeaders, verbose)
	default:
		log.Fatal("Unsupported output format:", outputFormat)
	}
}

func outputJSON(response *scraper.Response, showHeaders bool) {
	result := map[string]interface{}{
		"status_code": response.StatusCode,
		"url":         response.URL,
		"body":        response.Body,
	}

	if showHeaders {
		result["headers"] = response.Headers
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal JSON:", err)
	}

	fmt.Println(string(jsonOutput))
}

func outputText(response *scraper.Response, showHeaders, verbose bool) {
	if verbose {
		fmt.Printf("=== Response ===\n")
	}

	fmt.Printf("Status: %d\n", response.StatusCode)
	fmt.Printf("URL: %s\n", response.URL)

	if showHeaders {
		fmt.Println("Headers:")
		for key, values := range response.Headers {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}

	if verbose {
		fmt.Printf("Body Length: %d bytes\n", len(response.Body))
		fmt.Println("--- Body ---")
	}

	fmt.Println(response.Body)
}

// createJSConfig creates JavaScript engine configuration from CLI flags
func createJSConfig() scraper.JSEngineConfig {
	// Parse viewport
	width, height := int64(1920), int64(1080)
	if *viewport != "" {
		if parts := strings.Split(*viewport, "x"); len(parts) == 2 {
			if w, err := parseViewportDimension(parts[0]); err == nil {
				width = w
			}
			if h, err := parseViewportDimension(parts[1]); err == nil {
				height = h
			}
		}
	}

	return scraper.JSEngineConfig{
		Enabled:  *enableJS,
		Timeout:  *jsTimeout,
		Headless: *headless,
		NoImages: *noImages,
		Viewport: scraper.Viewport{
			Width:  width,
			Height: height,
		},
	}
}

// parseViewportDimension parses a viewport dimension string to int64
func parseViewportDimension(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// handleConcurrentRequests handles concurrent request processing
func handleConcurrentRequests(s *scraper.AdvancedScraper) {
	var err error

	if *verbose {
		log.Printf("Starting concurrent request processing...")
		log.Printf("Configuration: max_concurrent=%d, worker_pool_size=%d, requests_per_second=%.2f",
			*maxConcurrent, *workerPoolSize, *requestsPerSecond)
		log.Printf("Connection pool: size=%d, max_idle=%d, idle_timeout=%v",
			*connectionPoolSize, *maxIdleConns, *idleConnTimeout)
		if *enableMemoryOpt {
			log.Printf("Memory optimization enabled: max_memory=%dMB", *maxMemoryMB)
		}
		if *enableIntelligentQueue {
			log.Printf("Intelligent priority queueing enabled")
		}
	}

	// Create memory optimizer if enabled
	var memOptimizer *scraper.MemoryOptimizer
	if *enableMemoryOpt {
		memOptimizer = scraper.NewMemoryOptimizer(*maxMemoryMB)
		if *verbose {
			log.Printf("Memory optimizer initialized with %dMB limit", *maxMemoryMB)
		}
	}

	// Create intelligent queue if enabled
	var intelligentQueue *scraper.IntelligentQueue
	if *enableIntelligentQueue {
		intelligentQueue = scraper.NewIntelligentQueue(*queueSize)
		if *verbose {
			log.Printf("Intelligent queue initialized with priority levels")
		}
	}

	// Create concurrent engine configuration
	config := scraper.ConcurrencyConfig{
		MaxConcurrent:     *maxConcurrent,
		WorkerPoolSize:    *workerPoolSize,
		RequestBuffer:     *queueSize,
		RateLimitPerSec:   int(*requestsPerSecond),
		ConnectionTimeout: *timeout,
		IdleTimeout:       *idleConnTimeout,
		MaxIdleConns:      *maxIdleConns,
		MaxConnsPerHost:   *connectionPoolSize,
	}

	// Create concurrent engine (worker pool)
	engine := scraper.NewWorkerPool(s, config)
	defer engine.Stop()

	// Start the worker pool
	err = engine.Start()
	if err != nil {
		log.Fatal("Failed to start worker pool:", err)
	}

	if *verbose {
		log.Printf("Concurrent engine initialized and started successfully")
	}

	// Create a sample job (in real usage, jobs would come from various sources)
	var postData []byte
	if strings.ToUpper(*method) == "POST" {
		if *data == "" {
			log.Fatal("POST data is required for POST requests")
		}

		// Parse POST data
		postDataMap, err := parsePostData(*data)
		if err != nil {
			log.Fatal("Failed to parse POST data:", err)
		}

		// Convert to form encoded data
		formData := make([]string, 0, len(postDataMap))
		for key, value := range postDataMap {
			formData = append(formData, fmt.Sprintf("%s=%s", key, value))
		}
		postData = []byte(strings.Join(formData, "&"))
	}

	// Create request job with priority
	priority := 5 // Default normal priority
	if *enableIntelligentQueue {
		priority = 7 // Higher priority for demonstration
	}

	job := &scraper.RequestJob{
		ID:         "sample-job-1",
		URL:        *url,
		Method:     strings.ToUpper(*method),
		Data:       postData,
		Options:    make(map[string]interface{}),
		ResultChan: make(chan *scraper.JobResult, 1),
		CreatedAt:  time.Now(),
		Priority:   priority,
	}

	// Check memory before submitting job if optimization is enabled
	if memOptimizer != nil {
		if memOptimizer.CheckMemoryUsage() {
			if *verbose {
				log.Printf("Memory usage high, garbage collection triggered")
			}
		}
	}

	// Submit job via intelligent queue or regular queue
	if intelligentQueue != nil {
		err = intelligentQueue.EnqueueJob(job)
		if err != nil {
			log.Fatal("Failed to enqueue job:", err)
		}

		// Dequeue and submit to worker pool
		if dequeuedJob, ok := intelligentQueue.DequeueJob(); ok {
			err = engine.SubmitJob(dequeuedJob)
			if err != nil {
				log.Fatal("Failed to submit job:", err)
			}
		}
	} else {
		err = engine.SubmitJob(job)
		if err != nil {
			log.Fatal("Failed to submit job:", err)
		}
	}

	// Submit job
	if *verbose {
		log.Printf("Submitting %s request job for: %s (priority: %d)", job.Method, job.URL, job.Priority)
	}

	// Wait for result
	if *verbose {
		log.Printf("Waiting for job result...")
	}

	select {
	case result := <-job.ResultChan:
		if result.Error != nil {
			log.Fatal("Job failed:", result.Error)
		}

		if *verbose {
			log.Printf("Job completed successfully in %v", result.Duration)
		}

		// Output results
		outputResponse(result.Response, *output, *showHeaders, *verbose)

		// Show performance stats if requested
		if *showPerformanceStats {
			stats := engine.GetStats()
			fmt.Printf("\n=== Performance Statistics ===\n")
			fmt.Printf("Total Jobs Processed: %d\n", stats.TotalJobs)
			fmt.Printf("Completed Jobs: %d\n", stats.CompletedJobs)
			fmt.Printf("Failed Jobs: %d\n", stats.FailedJobs)
			fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
			fmt.Printf("Requests Per Second: %.2f\n", stats.RequestsPerSec)
			fmt.Printf("Active Workers: %d\n", stats.ActiveWorkers)
			fmt.Printf("Queue Length: %d\n", stats.QueueLength)
		}

		// Show memory stats if requested
		if *showMemoryStats && memOptimizer != nil {
			memStats := memOptimizer.GetMemoryStats()
			fmt.Printf("\n=== Memory Statistics ===\n")
			if statsJSON, err := json.MarshalIndent(memStats, "", "  "); err == nil {
				fmt.Println(string(statsJSON))
			}
		}

		// Show intelligent queue stats if enabled
		if *enableIntelligentQueue && intelligentQueue != nil {
			queueStats := intelligentQueue.GetQueueStats()
			fmt.Printf("\n=== Queue Statistics ===\n")
			fmt.Printf("High Priority Jobs: %d\n", queueStats.HighPriorityJobs)
			fmt.Printf("Normal Priority Jobs: %d\n", queueStats.NormalPriorityJobs)
			fmt.Printf("Low Priority Jobs: %d\n", queueStats.LowPriorityJobs)
			fmt.Printf("Total Queued: %d\n", queueStats.TotalQueued)
			fmt.Printf("Dropped Jobs: %d\n", queueStats.DroppedJobs)
		}

	case <-time.After(60 * time.Second):
		log.Fatal("Job timed out after 60 seconds")
	}

	// Cleanup
	if intelligentQueue != nil {
		intelligentQueue.Close()
	}
}
