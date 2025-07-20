package main

import (
	"fmt"
	"log"

	"anti-bot-scraper/internal/scraper"
)

func main() {
	fmt.Println("=== Anti-Bot TLS Fingerprint Scraper Demo ===")

	// Test different browser fingerprints
	fingerprints := []struct {
		name        string
		fingerprint scraper.Fingerprint
	}{
		{"Chrome", scraper.ChromeFingerprint},
		{"Firefox", scraper.FirefoxFingerprint},
		{"Safari", scraper.SafariFingerprint},
		{"Edge", scraper.EdgeFingerprint},
	}

	testURL := "https://httpbin.org/headers"

	for _, fp := range fingerprints {
		fmt.Printf("\n--- Testing %s Fingerprint ---\n", fp.name)

		// Create scraper with specific fingerprint
		s, err := scraper.NewScraper(fp.fingerprint)
		if err != nil {
			log.Printf("Failed to create %s scraper: %v", fp.name, err)
			continue
		}

		// Make request
		response, err := s.Get(testURL)
		if err != nil {
			log.Printf("Failed to scrape with %s: %v", fp.name, err)
			continue
		}

		fmt.Printf("Status: %d\n", response.StatusCode)
		truncated := truncateString(response.Body, 500)
		fmt.Printf("Response Body (first 500 chars):\n%s\n", truncated)
	}

	// Test a more complex scenario
	fmt.Println("\n=== Testing with Real Website ===")
	s, err := scraper.NewScraper(scraper.ChromeFingerprint)
	if err != nil {
		log.Fatal("Failed to create Chrome scraper:", err)
	}

	// Test with a website that might have bot detection
	testSites := []string{
		"https://httpbin.org/user-agent",
		"https://httpbin.org/ip",
		"https://httpbin.org/json",
	}

	for _, site := range testSites {
		fmt.Printf("\nTesting: %s\n", site)
		response, err := s.Get(site)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("Status: %d\n", response.StatusCode)
		truncated := truncateString(response.Body, 200)
		fmt.Printf("Response: %s\n", truncated)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
