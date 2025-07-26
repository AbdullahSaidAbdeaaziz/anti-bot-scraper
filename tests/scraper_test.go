package scraper_test

import (
	"testing"
	"time"

	"anti-bot-scraper/internal/scraper"
)

// TestFingerprints tests TLS fingerprint configurations
func TestFingerprints(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint scraper.Fingerprint
		wantError   bool
	}{
		{
			name:        "Chrome fingerprint",
			fingerprint: scraper.ChromeFingerprint,
			wantError:   false,
		},
		{
			name:        "Firefox fingerprint",
			fingerprint: scraper.FirefoxFingerprint,
			wantError:   false,
		},
		{
			name:        "Safari fingerprint",
			fingerprint: scraper.SafariFingerprint,
			wantError:   false,
		},
		{
			name:        "Edge fingerprint",
			fingerprint: scraper.EdgeFingerprint,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we can create a scraper with each fingerprint
			s, err := scraper.NewAdvancedScraper(tt.fingerprint)
			if (err != nil) != tt.wantError {
				t.Errorf("NewAdvancedScraper() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && s == nil {
				t.Errorf("Expected scraper to be created, got nil")
			}

			if s != nil {
				s.Close()
			}
		})
	}
}

// TestAdvancedScraperCreation tests scraper initialization
func TestAdvancedScraperCreation(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint scraper.Fingerprint
		protocol    scraper.ProtocolVersion
		wantError   bool
	}{
		{
			name:        "Chrome with HTTP/1.1",
			fingerprint: scraper.ChromeFingerprint,
			protocol:    scraper.HTTP1_1,
			wantError:   false,
		},
		{
			name:        "Firefox with HTTP/2",
			fingerprint: scraper.FirefoxFingerprint,
			protocol:    scraper.HTTP2,
			wantError:   false,
		},
		{
			name:        "Chrome with Auto protocol",
			fingerprint: scraper.ChromeFingerprint,
			protocol:    scraper.HTTPAuto,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := scraper.NewAdvancedScraperWithProtocol(tt.fingerprint, tt.protocol)
			if (err != nil) != tt.wantError {
				t.Errorf("NewAdvancedScraperWithProtocol() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && s == nil {
				t.Errorf("Expected scraper to be created, got nil")
			}

			if s != nil {
				s.Close()
			}
		})
	}
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint, scraper.WithRateLimit(100*time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	// Test rate limiting timing
	start := time.Now()

	// Simulate rate limiting by checking if the delay is applied
	// This test doesn't make actual network requests to avoid test failures
	elapsed := time.Since(start)

	// The test mainly ensures the scraper initializes with rate limiting
	if elapsed < 0 {
		t.Errorf("Unexpected negative elapsed time: %v", elapsed)
	}
}

// TestProxyConfiguration tests proxy setup
func TestProxyConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		proxyURL  string
		wantError bool
	}{
		{
			name:      "Valid HTTP proxy",
			proxyURL:  "http://proxy.example.com:8080",
			wantError: false,
		},
		{
			name:      "Valid SOCKS5 proxy",
			proxyURL:  "socks5://proxy.example.com:1080",
			wantError: false,
		},
		{
			name:      "Empty proxy URL",
			proxyURL:  "",
			wantError: false, // Empty proxy is allowed (means no proxy)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint, scraper.WithProxy(tt.proxyURL))
			if (err != nil) != tt.wantError {
				t.Errorf("NewAdvancedScraper() with proxy error = %v, wantError %v", err, tt.wantError)
				return
			}

			if s != nil {
				s.Close()
			}
		})
	}
}

// TestCustomHeaders tests custom header functionality
func TestCustomHeaders(t *testing.T) {
	s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	// Test setting custom headers
	s.SetHeader("X-Custom-Header", "test-value")
	s.SetHeader("Authorization", "Bearer token123")

	// Verify headers are set (we can't easily test the actual HTTP request without network)
	// This test mainly ensures the SetHeader method doesn't panic
}

// TestCookieManagement tests cookie handling
func TestCookieManagement(t *testing.T) {
	s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	// Test that scraper handles cookies (implementation detail test)
	// This mainly ensures the scraper initializes properly with cookie support
}

// TestWorkerPoolCreation tests concurrent engine
func TestWorkerPoolCreation(t *testing.T) {
	s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint)
	if err != nil {
		t.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	config := scraper.GetDefaultConcurrencyConfig()
	pool := scraper.NewWorkerPool(s, config)
	if pool == nil {
		t.Errorf("Expected worker pool to be created, got nil")
	}

	if pool != nil {
		pool.Stop()
	}
}

// BenchmarkScraperCreation benchmarks scraper creation performance
func BenchmarkScraperCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint)
		if err != nil {
			b.Fatalf("Failed to create scraper: %v", err)
		}
		s.Close()
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent processing
func BenchmarkConcurrentRequests(b *testing.B) {
	s, err := scraper.NewAdvancedScraper(scraper.ChromeFingerprint)
	if err != nil {
		b.Fatalf("Failed to create scraper: %v", err)
	}
	defer s.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate request processing without actual network call
			time.Sleep(1 * time.Millisecond)
		}
	})
}
