package main

import (
	"fmt"
	"log"
	"time"

	"anti-bot-scraper/internal/scraper"
)

func main() {
	fmt.Println("=== Advanced Anti-Bot Scraper Demo ===")

	// Create an advanced scraper with retry and rate limiting
	advancedScraper, err := scraper.NewAdvancedScraper(
		scraper.ChromeFingerprint,
		scraper.WithRetryCount(3),
		scraper.WithRateLimit(2*time.Second),
	)
	if err != nil {
		log.Fatal("Failed to create advanced scraper:", err)
	}

	// Test sites that might detect bots
	testSites := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/503", // This will trigger retries
		"https://httpbin.org/delay/1",    // Simulates slow response
		"https://httpbin.org/cookies/set?test=value",
		"https://httpbin.org/cookies", // Should show the cookie we set
	}

	for i, site := range testSites {
		fmt.Printf("\n[%d] Testing: %s\n", i+1, site)

		start := time.Now()
		response, err := advancedScraper.GetWithRetry(site)
		duration := time.Since(start)

		if err != nil {
			log.Printf("Error: %v (took %v)", err, duration)
			continue
		}

		fmt.Printf("Status: %d (took %v)\n", response.StatusCode, duration)

		// Show response body for relevant tests
		if len(response.Body) < 500 {
			fmt.Printf("Response: %s\n", response.Body)
		} else {
			fmt.Printf("Response: %s...\n", response.Body[:200])
		}
	}

	// Test POST request
	fmt.Println("\n=== Testing POST Request ===")
	postData := map[string]string{
		"username": "testuser",
		"password": "testpass",
		"action":   "login",
	}

	postResponse, err := advancedScraper.PostWithRetry("https://httpbin.org/post", postData)
	if err != nil {
		log.Printf("POST Error: %v", err)
	} else {
		fmt.Printf("POST Status: %d\n", postResponse.StatusCode)
		if len(postResponse.Body) > 1000 {
			fmt.Printf("POST Response: %s...\n", postResponse.Body[:500])
		} else {
			fmt.Printf("POST Response: %s\n", postResponse.Body)
		}
	}
}
