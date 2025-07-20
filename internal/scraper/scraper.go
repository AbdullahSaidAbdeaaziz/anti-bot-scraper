package scraper

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
)

// Fingerprint represents different browser TLS fingerprints
type Fingerprint int

const (
	ChromeFingerprint Fingerprint = iota
	FirefoxFingerprint
	SafariFingerprint
	EdgeFingerprint
)

// Scraper represents the main scraper client
type Scraper struct {
	client      *http.Client
	fingerprint Fingerprint
	userAgent   string
	headers     map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
	URL        string
}

// NewScraper creates a new scraper instance with the specified fingerprint
func NewScraper(fingerprint Fingerprint) (*Scraper, error) {
	client, err := createTLSClient(fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client: %w", err)
	}

	scraper := &Scraper{
		client:      client,
		fingerprint: fingerprint,
		userAgent:   getUserAgent(fingerprint),
		headers:     getDefaultHeaders(fingerprint),
	}

	return scraper, nil
}

// Get performs a GET request to the specified URL
func (s *Scraper) Get(url string) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	for key, value := range s.headers {
		req.Header.Set(key, value)
	}

	// Set user agent
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(body),
		URL:        url,
	}, nil
}

// createTLSClient creates an HTTP client with the specified TLS fingerprint
func createTLSClient(fingerprint Fingerprint) (*http.Client, error) {
	// Create a custom transport that uses uTLS
	transport := &http.Transport{
		// Disable HTTP/2 to avoid compatibility issues with malformed responses
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	// Override the DialTLS function to use uTLS with forced HTTP/1.1
	transport.DialTLS = func(network, addr string) (net.Conn, error) {
		// Establish TCP connection
		conn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		// Extract hostname for SNI
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		}

		// Create uTLS connection - use custom spec instead of predefined to ensure HTTP/1.1
		uconn := utls.UClient(conn, &utls.Config{
			ServerName: host,
			NextProtos: []string{"http/1.1"}, // Force HTTP/1.1 only
		}, utls.HelloCustom)

		// Apply a basic ClientHello that doesn't advertise HTTP/2
		err = uconn.ApplyPreset(&utls.ClientHelloSpec{
			TLSVersMax: utls.VersionTLS12,
			TLSVersMin: utls.VersionTLS12,
			CipherSuites: []uint16{
				utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			},
			Extensions: []utls.TLSExtension{
				&utls.SNIExtension{},
				&utls.SupportedCurvesExtension{Curves: []utls.CurveID{utls.X25519, utls.CurveP256, utls.CurveP384}},
				&utls.SupportedPointsExtension{SupportedPoints: []byte{0}},
				&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
					utls.ECDSAWithP256AndSHA256,
					utls.PSSWithSHA256,
					utls.PKCS1WithSHA256,
					utls.ECDSAWithP384AndSHA384,
					utls.PSSWithSHA384,
					utls.PKCS1WithSHA384,
					utls.PSSWithSHA512,
					utls.PKCS1WithSHA512,
				}},
			},
		})
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to apply preset: %w", err)
		}

		err = uconn.Handshake()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("TLS handshake failed: %w", err)
		}

		return uconn, nil
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

// getClientHelloID returns the appropriate ClientHelloID for the fingerprint
func getClientHelloID(fingerprint Fingerprint) utls.ClientHelloID {
	switch fingerprint {
	case ChromeFingerprint:
		return utls.HelloChrome_120
	case FirefoxFingerprint:
		return utls.HelloFirefox_120
	case SafariFingerprint:
		return utls.HelloSafari_16_0
	case EdgeFingerprint:
		return utls.HelloChrome_120 // Edge uses Chromium
	default:
		return utls.HelloChrome_120
	}
}
