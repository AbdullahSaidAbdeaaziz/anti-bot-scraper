package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"anti-bot-scraper/internal/scraper"
)

var (
	// Command line flags
	url           = flag.String("url", "", "URL to scrape (required)")
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
	httpVersion          = flag.String("http-version", "1.1", "HTTP version: '1.1', '2', or 'auto'")
	userAgent            = flag.String("user-agent", "", "Custom User-Agent (overrides browser default)")
	timeout              = flag.Duration("timeout", 30*time.Second, "Request timeout")
	verbose              = flag.Bool("verbose", false, "Verbose output")
	showHeaders          = flag.Bool("show-headers", false, "Show response headers")
	followRedirect       = flag.Bool("follow-redirect", true, "Follow redirects")
	version              = flag.Bool("version", false, "Show version information")
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

	if *url == "" {
		fmt.Fprintf(os.Stderr, "Error: URL is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *verbose {
		log.Printf("Starting %s v%s", appName, appVersion)
		log.Printf("Target URL: %s", *url)
		log.Printf("Browser: %s", *browser)
		log.Printf("Method: %s", *method)
	}

	// Parse browser fingerprint
	fingerprint, err := parseBrowserFingerprint(*browser)
	if err != nil {
		log.Fatal("Error:", err)
	}

	// Create scraper options
	options := []scraper.ScraperOption{
		scraper.WithRetryCount(*retries),
		scraper.WithRateLimit(*rateLimit),
	}

	// Add proxy configuration
	if *proxies != "" {
		// Multiple proxies for rotation
		proxyList := strings.Split(*proxies, ",")
		for i, p := range proxyList {
			proxyList[i] = strings.TrimSpace(p)
		}

		var rotationMode scraper.ProxyRotationMode
		switch *proxyRotation {
		case "per-request":
			rotationMode = scraper.RotatePerRequest
		case "on-error":
			rotationMode = scraper.RotateOnError
		case "health-aware":
			rotationMode = scraper.HealthAware
		default:
			log.Fatal("Invalid proxy rotation mode. Use 'per-request', 'on-error', or 'health-aware'")
		}

		if *verbose {
			log.Printf("Using proxy rotation with %d proxies, mode: %s", len(proxyList), *proxyRotation)
		}

		// Use health-aware proxy rotation if requested or health monitoring is enabled
		if rotationMode == scraper.HealthAware || *enableProxyHealth {
			healthConfig := scraper.ProxyHealthConfig{
				CheckInterval: *proxyHealthInterval,
				Timeout:       *proxyHealthTimeout,
				TestURL:       *proxyHealthTestURL,
				MaxFailures:   *proxyMaxFailures,
			}

			if *verbose {
				log.Printf("Enabling proxy health monitoring with %d proxies", len(proxyList))
				log.Printf("Health check interval: %v, timeout: %v", *proxyHealthInterval, *proxyHealthTimeout)
			}

			options = append(options, scraper.WithHealthAwareProxyRotation(proxyList, healthConfig))
		} else {
			options = append(options, scraper.WithProxyRotation(proxyList, rotationMode))
		}

	} else if *proxy != "" {
		// Single proxy
		if *verbose {
			log.Printf("Using single proxy: %s", *proxy)
		}
		options = append(options, scraper.WithProxy(*proxy))
	}

	// Add CAPTCHA configuration
	if *enableCaptcha {
		if *captchaAPIKey == "" {
			log.Fatal("CAPTCHA API key is required when CAPTCHA solving is enabled")
		}

		var service scraper.CaptchaService
		switch *captchaService {
		case "2captcha":
			service = scraper.TwoCaptchaService
		case "deathbycaptcha":
			service = scraper.DeathByCaptchaService
		case "anticaptcha":
			service = scraper.AntiCaptchaService
		case "capmonster":
			service = scraper.CapMonsterService
		default:
			log.Fatal("Invalid CAPTCHA service. Use '2captcha', 'deathbycaptcha', 'anticaptcha', or 'capmonster'")
		}

		captchaConfig := scraper.CaptchaSolverConfig{
			Service:      service,
			APIKey:       *captchaAPIKey,
			Timeout:      *captchaTimeout,
			PollInterval: *captchaPollInterval,
			MaxRetries:   *captchaMaxRetries,
			MinScore:     *captchaMinScore,
			Language:     "en",
		}

		if *verbose {
			log.Printf("CAPTCHA solving enabled with service: %s", *captchaService)
			log.Printf("CAPTCHA timeout: %v, poll interval: %v", *captchaTimeout, *captchaPollInterval)
		}

		options = append(options, scraper.WithCaptchaDetection(captchaConfig))
	}

	// Add Human Behavior Simulation configuration
	if *enableBehavior {
		var behaviorTypeEnum scraper.BehaviorType
		switch *behaviorType {
		case "normal":
			behaviorTypeEnum = scraper.NormalBehavior
		case "cautious":
			behaviorTypeEnum = scraper.CautiousBehavior
		case "aggressive":
			behaviorTypeEnum = scraper.AggressiveBehavior
		case "random":
			behaviorTypeEnum = scraper.RandomBehavior
		default:
			log.Fatal("Invalid behavior type. Use 'normal', 'cautious', 'aggressive', or 'random'")
		}

		behaviorConfig := scraper.HumanBehaviorConfig{
			Enabled:             true,
			BehaviorType:        behaviorTypeEnum,
			MinDelay:            *behaviorMinDelay,
			MaxDelay:            *behaviorMaxDelay,
			MouseMovement:       *enableMouseMove,
			ScrollSimulation:    *enableScrollSim,
			TypingDelay:         *enableTypingDelay,
			PageLoadWait:        true,
			RandomScrolling:     *enableRandomActivity,
			RandomClicks:        *enableRandomActivity,
			ViewportVariation:   true,
			TabSwitchSimulation: false,
		}

		if *verbose || *showBehaviorInfo {
			log.Printf("Human behavior simulation enabled with type: %s", *behaviorType)
			log.Printf("Behavior delays: min=%v, max=%v", *behaviorMinDelay, *behaviorMaxDelay)
			log.Printf("Mouse movement: %v, Scroll simulation: %v, Typing delay: %v",
				*enableMouseMove, *enableScrollSim, *enableTypingDelay)
			if *enableRandomActivity {
				log.Printf("Random activity simulation enabled")
			}
		}

		options = append(options, scraper.WithHumanBehavior(behaviorConfig))
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
		log.Fatal("Invalid HTTP version. Use '1.1', '2', or 'auto'")
	}

	if *verbose {
		log.Printf("HTTP Version: %s", *httpVersion)
		if *enableJS {
			log.Printf("JavaScript Engine: enabled")
			log.Printf("JavaScript Mode: %s", *jsMode)
		}
	}

	// Configure JavaScript engine if enabled
	var jsConfig scraper.JSEngineConfig
	if *enableJS {
		jsConfig = createJSConfig()
	}

	// Create scraper with JavaScript support if enabled
	var s *scraper.AdvancedScraper
	if *enableJS {
		s, err = scraper.NewAdvancedScraperWithJS(fingerprint, protocol, jsConfig, options...)
	} else {
		s, err = scraper.NewAdvancedScraperWithProtocol(fingerprint, protocol, options...)
	}
	if err != nil {
		log.Fatal("Failed to create scraper:", err)
	}
	defer s.Close()

	// Override user agent if provided
	if *userAgent != "" {
		s.SetUserAgent(*userAgent)
	}

	// Add custom headers if provided
	if *headers != "" {
		customHeaders, err := parseHeaders(*headers)
		if err != nil {
			log.Fatal("Failed to parse headers:", err)
		}
		for key, value := range customHeaders {
			s.SetHeader(key, value)
		}
	}

	var response *scraper.Response

	// Execute request based on method
	switch strings.ToUpper(*method) {
	case "GET":
		if *verbose {
			log.Printf("Making GET request to %s", *url)
		}
		response, err = s.GetWithRetry(*url)
	case "POST":
		if *data == "" {
			log.Fatal("POST data is required for POST requests")
		}
		postData, err := parsePostData(*data)
		if err != nil {
			log.Fatal("Failed to parse POST data:", err)
		}
		if *verbose {
			log.Printf("Making POST request to %s", *url)
		}
		response, err = s.PostWithRetry(*url, postData)
	default:
		log.Fatal("Unsupported HTTP method:", *method)
	}

	if err != nil {
		log.Fatal("Request failed:", err)
	}

	// Output results
	outputResponse(response, *output, *showHeaders, *verbose)

	// Show proxy metrics if requested
	if *showProxyMetrics && (*proxies != "" || *enableProxyHealth) {
		metrics := s.GetProxyMetrics()
		fmt.Printf("\n=== Proxy Metrics ===\n")
		if metricsJSON, err := json.MarshalIndent(metrics, "", "  "); err == nil {
			fmt.Println(string(metricsJSON))
		}

		// Show detailed proxy health if available
		if *verbose {
			health := s.GetAllProxiesHealth()
			if len(health) > 0 {
				fmt.Printf("\n=== Proxy Health Details ===\n")
				if healthJSON, err := json.MarshalIndent(health, "", "  "); err == nil {
					fmt.Println(string(healthJSON))
				}
			}
		}
	}
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
