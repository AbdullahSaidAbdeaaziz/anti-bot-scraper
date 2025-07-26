package scraper

import "time"

// HeaderMimicryConfig configures browser-consistent header mimicry
type HeaderMimicryConfig struct {
	Profile           Fingerprint // Browser profile to mimic
	IncludeSecHeaders bool        // Include security headers (Sec-CH-UA, Sec-Fetch-*)
	AcceptLanguage    string      // Accept-Language header value
	AcceptEncoding    string      // Accept-Encoding header value
	CustomUserAgent   string      // Custom User-Agent override
}

// CookiePersistence defines how cookies are persisted
type CookiePersistence int

const (
	// SessionPersistence keeps cookies only for the session
	SessionPersistence CookiePersistence = iota
	// ProxyPersistence maintains separate cookie jars per proxy
	ProxyPersistence
	// NoPersistence disables cookie persistence
	NoPersistence
)

// CookieConfig configures cookie handling behavior
type CookieConfig struct {
	EnableJar      bool              // Enable in-memory cookie jar
	Persistence    CookiePersistence // Cookie persistence mode
	ClearOnRequest bool              // Clear cookies before each request
	CookieFile     string            // File to save/load cookies
}

// RedirectConfig configures HTTP redirect handling
type RedirectConfig struct {
	FollowRedirects bool          // Follow HTTP redirects
	MaxRedirects    int           // Maximum number of redirects to follow
	Timeout         time.Duration // Timeout for redirect chains
}

// WithHeaderMimicry adds browser-consistent header mimicry to scraper options
func WithHeaderMimicry(config HeaderMimicryConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		// For now, just store the config for future use
		// Implementation would add browser-consistent headers based on the fingerprint
		if config.CustomUserAgent != "" {
			s.SetUserAgent(config.CustomUserAgent)
		}

		// Apply browser-specific headers based on profile
		applyBrowserHeaders(s, config)
	}
}

// WithCookieHandling adds cookie handling configuration to scraper options
func WithCookieHandling(config CookieConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		if !config.EnableJar {
			s.cookieJar = nil
		}
		// Additional cookie handling logic would go here
	}
}

// WithRedirectHandling adds redirect handling configuration to scraper options
func WithRedirectHandling(config RedirectConfig) ScraperOption {
	return func(s *AdvancedScraper) {
		// Configure redirect handling on the underlying HTTP client
		// This would require modifications to the base scraper
	}
}

// WithTimeout adds timeout configuration to scraper options
func WithTimeout(timeout time.Duration) ScraperOption {
	return func(s *AdvancedScraper) {
		if s.client != nil {
			s.client.Timeout = timeout
		}
	}
}

// applyBrowserHeaders applies browser-specific headers based on the fingerprint
func applyBrowserHeaders(s *AdvancedScraper, config HeaderMimicryConfig) {
	switch config.Profile {
	case ChromeFingerprint:
		if config.AcceptLanguage == "auto" || config.AcceptLanguage == "" {
			s.SetHeader("Accept-Language", "en-US,en;q=0.9")
		} else {
			s.SetHeader("Accept-Language", config.AcceptLanguage)
		}

		if config.AcceptEncoding == "auto" || config.AcceptEncoding == "" {
			s.SetHeader("Accept-Encoding", "gzip, deflate, br")
		} else {
			s.SetHeader("Accept-Encoding", config.AcceptEncoding)
		}

		if config.IncludeSecHeaders {
			s.SetHeader("Sec-Fetch-Dest", "document")
			s.SetHeader("Sec-Fetch-Mode", "navigate")
			s.SetHeader("Sec-Fetch-Site", "none")
			s.SetHeader("Sec-Fetch-User", "?1")
		}

	case FirefoxFingerprint:
		if config.AcceptLanguage == "auto" || config.AcceptLanguage == "" {
			s.SetHeader("Accept-Language", "en-US,en;q=0.5")
		} else {
			s.SetHeader("Accept-Language", config.AcceptLanguage)
		}

		if config.AcceptEncoding == "auto" || config.AcceptEncoding == "" {
			s.SetHeader("Accept-Encoding", "gzip, deflate, br")
		} else {
			s.SetHeader("Accept-Encoding", config.AcceptEncoding)
		}

	case SafariFingerprint:
		if config.AcceptLanguage == "auto" || config.AcceptLanguage == "" {
			s.SetHeader("Accept-Language", "en-US,en;q=0.9")
		} else {
			s.SetHeader("Accept-Language", config.AcceptLanguage)
		}

		if config.AcceptEncoding == "auto" || config.AcceptEncoding == "" {
			s.SetHeader("Accept-Encoding", "gzip, deflate, br")
		} else {
			s.SetHeader("Accept-Encoding", config.AcceptEncoding)
		}

	case EdgeFingerprint:
		if config.AcceptLanguage == "auto" || config.AcceptLanguage == "" {
			s.SetHeader("Accept-Language", "en-US,en;q=0.9")
		} else {
			s.SetHeader("Accept-Language", config.AcceptLanguage)
		}

		if config.AcceptEncoding == "auto" || config.AcceptEncoding == "" {
			s.SetHeader("Accept-Encoding", "gzip, deflate, br")
		} else {
			s.SetHeader("Accept-Encoding", config.AcceptEncoding)
		}

		if config.IncludeSecHeaders {
			s.SetHeader("Sec-Fetch-Dest", "document")
			s.SetHeader("Sec-Fetch-Mode", "navigate")
			s.SetHeader("Sec-Fetch-Site", "none")
			s.SetHeader("Sec-Fetch-User", "?1")
		}
	}
}
