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
	url            = flag.String("url", "", "URL to scrape (required)")
	browser        = flag.String("browser", "chrome", "Browser fingerprint (chrome, firefox, safari, edge)")
	method         = flag.String("method", "GET", "HTTP method (GET, POST)")
	headers        = flag.String("headers", "", "Custom headers in JSON format or @filename to read from file")
	data           = flag.String("data", "", "POST data in JSON format or @filename to read from file")
	output         = flag.String("output", "text", "Output format (text, json)")
	retries        = flag.Int("retries", 3, "Number of retries")
	rateLimit      = flag.Duration("rate-limit", 1*time.Second, "Rate limit between requests")
	proxy          = flag.String("proxy", "", "Proxy URL (http://proxy:port or socks5://proxy:port)")
	userAgent      = flag.String("user-agent", "", "Custom User-Agent (overrides browser default)")
	timeout        = flag.Duration("timeout", 30*time.Second, "Request timeout")
	verbose        = flag.Bool("verbose", false, "Verbose output")
	showHeaders    = flag.Bool("show-headers", false, "Show response headers")
	followRedirect = flag.Bool("follow-redirect", true, "Follow redirects")
	version        = flag.Bool("version", false, "Show version information")
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

	// Add proxy if provided
	if *proxy != "" {
		if *verbose {
			log.Printf("Using proxy: %s", *proxy)
		}
		options = append(options, scraper.WithProxy(*proxy))
	}

	// Create scraper with options
	s, err := scraper.NewAdvancedScraper(fingerprint, options...)
	if err != nil {
		log.Fatal("Failed to create scraper:", err)
	}

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
