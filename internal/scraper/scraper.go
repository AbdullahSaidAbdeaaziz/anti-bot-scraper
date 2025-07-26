package scraper

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// Fingerprint represents different browser TLS fingerprints
type Fingerprint int

const (
	ChromeFingerprint Fingerprint = iota
	FirefoxFingerprint
	SafariFingerprint
	EdgeFingerprint
)

// ProtocolVersion represents HTTP protocol versions
type ProtocolVersion int

const (
	HTTP1_1 ProtocolVersion = iota
	HTTP2
	HTTPAuto // Auto-detect based on server capabilities
)

// Scraper represents the main scraper client
type Scraper struct {
	client      *http.Client
	fingerprint Fingerprint
	userAgent   string
	headers     map[string]string
	protocol    ProtocolVersion
	jsEngine    *JSEngine      // JavaScript engine for dynamic content
	jsConfig    JSEngineConfig // JS engine configuration
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
	URL        string
	JSEnabled  bool     // Whether JavaScript was used
	Console    []string // Console messages from JS execution
	JSErrors   []string // JavaScript errors if any
}

// NewScraper creates a new scraper instance with the specified fingerprint (HTTP/1.1 for compatibility)
func NewScraper(fingerprint Fingerprint) (*Scraper, error) {
	return NewScraperWithProtocol(fingerprint, HTTP1_1)
}

// NewScraperWithProtocol creates a new scraper with specified protocol version
func NewScraperWithProtocol(fingerprint Fingerprint, protocol ProtocolVersion) (*Scraper, error) {
	return NewScraperWithJS(fingerprint, protocol, JSEngineConfig{Enabled: false})
}

// NewScraperWithJS creates a new scraper with JavaScript engine support
func NewScraperWithJS(fingerprint Fingerprint, protocol ProtocolVersion, jsConfig JSEngineConfig) (*Scraper, error) {
	client, err := createTLSClientWithProtocol(fingerprint, protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client: %w", err)
	}

	// Initialize JavaScript engine if enabled
	var jsEngine *JSEngine
	if jsConfig.Enabled {
		// Set browser-specific user agent for JS engine
		if jsConfig.UserAgent == "" {
			jsConfig.UserAgent = getUserAgent(fingerprint)
		}

		jsEngine, err = NewJSEngine(jsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create JavaScript engine: %w", err)
		}
	}

	scraper := &Scraper{
		client:      client,
		fingerprint: fingerprint,
		userAgent:   getUserAgent(fingerprint),
		headers:     getDefaultHeaders(fingerprint),
		protocol:    protocol,
		jsEngine:    jsEngine,
		jsConfig:    jsConfig,
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
		JSEnabled:  false,
		Console:    []string{},
		JSErrors:   []string{},
	}, nil
}

// GetWithJS performs a GET request with JavaScript execution enabled
func (s *Scraper) GetWithJS(url string) (*Response, error) {
	if s.jsEngine == nil {
		return nil, fmt.Errorf("JavaScript engine not initialized - use NewScraperWithJS")
	}

	jsResp, err := s.jsEngine.Scrape(url, s.jsConfig)
	if err != nil {
		return nil, fmt.Errorf("JavaScript scraping failed: %w", err)
	}

	return &Response{
		StatusCode: jsResp.StatusCode,
		Headers:    convertMapToHeader(jsResp.Headers),
		Body:       jsResp.HTML,
		URL:        jsResp.URL,
		JSEnabled:  true,
		Console:    jsResp.Console,
		JSErrors:   jsResp.Errors,
	}, nil
}

// GetWithBehavior performs a GET request with human-like behavior simulation
func (s *Scraper) GetWithBehavior(url string) (*Response, error) {
	if s.jsEngine == nil {
		return nil, fmt.Errorf("JavaScript engine not initialized - use NewScraperWithJS")
	}

	jsResp, err := s.jsEngine.SimulateHumanBehavior(url, s.jsConfig)
	if err != nil {
		return nil, fmt.Errorf("human behavior simulation failed: %w", err)
	}

	return &Response{
		StatusCode: jsResp.StatusCode,
		Headers:    convertMapToHeader(jsResp.Headers),
		Body:       jsResp.HTML,
		URL:        jsResp.URL,
		JSEnabled:  true,
		Console:    jsResp.Console,
		JSErrors:   jsResp.Errors,
	}, nil
}

// WaitForElement waits for a specific element to appear before returning the page
func (s *Scraper) WaitForElement(url string, selector string) (*Response, error) {
	if s.jsEngine == nil {
		return nil, fmt.Errorf("JavaScript engine not initialized - use NewScraperWithJS")
	}

	jsResp, err := s.jsEngine.WaitForElement(url, selector, s.jsConfig)
	if err != nil {
		return nil, fmt.Errorf("waiting for element failed: %w", err)
	}

	return &Response{
		StatusCode: jsResp.StatusCode,
		Headers:    convertMapToHeader(jsResp.Headers),
		Body:       jsResp.HTML,
		URL:        jsResp.URL,
		JSEnabled:  true,
		Console:    jsResp.Console,
		JSErrors:   jsResp.Errors,
	}, nil
}

// ExecuteJS executes custom JavaScript code and returns the result
func (s *Scraper) ExecuteJS(url string, jsCode string) (*Response, error) {
	if s.jsEngine == nil {
		return nil, fmt.Errorf("JavaScript engine not initialized - use NewScraperWithJS")
	}

	jsResp, err := s.jsEngine.ExecuteJS(url, jsCode, s.jsConfig)
	if err != nil {
		return nil, fmt.Errorf("JavaScript execution failed: %w", err)
	}

	return &Response{
		StatusCode: jsResp.StatusCode,
		Headers:    convertMapToHeader(jsResp.Headers),
		Body:       jsResp.HTML,
		URL:        jsResp.URL,
		JSEnabled:  true,
		Console:    jsResp.Console,
		JSErrors:   jsResp.Errors,
	}, nil
}

// Close cleans up the scraper resources
func (s *Scraper) Close() error {
	if s.jsEngine != nil {
		return s.jsEngine.Close()
	}
	return nil
}

// convertMapToHeader converts a map[string]string to http.Header
func convertMapToHeader(headers map[string]string) map[string][]string {
	result := make(map[string][]string)
	for key, value := range headers {
		result[key] = []string{value}
	}
	return result
}

// createTLSClient creates an HTTP client with the specified TLS fingerprint (HTTP/1.1 only)
func createTLSClient(fingerprint Fingerprint) (*http.Client, error) {
	return createTLSClientWithProtocol(fingerprint, HTTP1_1)
}

// createTLSClientWithProtocol creates an HTTP client with specified protocol support
func createTLSClientWithProtocol(fingerprint Fingerprint, protocol ProtocolVersion) (*http.Client, error) {
	switch protocol {
	case HTTP1_1:
		// Create HTTP/1.1 only transport
		transport := &http.Transport{
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper), // Disable HTTP/2
		}

		// Override the DialTLS function to use uTLS
		transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialTLSWithProtocol(network, addr, fingerprint, protocol)
		}

		client := &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
		return client, nil

	case HTTP2, HTTPAuto:
		// Create HTTP/2 transport directly
		transport2 := &http2.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return dialTLSWithProtocol(network, addr, fingerprint, protocol)
			},
		}

		client := &http.Client{
			Transport: transport2,
			Timeout:   30 * time.Second,
		}
		return client, nil
	}

	return nil, fmt.Errorf("unsupported protocol version")
}

// dialTLSWithProtocol handles TLS dialing with protocol-specific configuration
func dialTLSWithProtocol(network, addr string, fingerprint Fingerprint, protocol ProtocolVersion) (net.Conn, error) {
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

	// Configure Next Protocols based on desired protocol version
	var nextProtos []string
	switch protocol {
	case HTTP1_1:
		nextProtos = []string{"http/1.1"}
	case HTTP2:
		nextProtos = []string{"h2", "http/1.1"}
	case HTTPAuto:
		nextProtos = []string{"h2", "http/1.1"}
	}

	// Create uTLS connection
	uconn := utls.UClient(conn, &utls.Config{
		ServerName: host,
		NextProtos: nextProtos,
	}, utls.HelloCustom)

	// Apply enhanced ClientHello based on fingerprint
	err = uconn.ApplyPreset(getBrowserClientHelloSpec(fingerprint, protocol))
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

// getBrowserClientHelloSpec returns browser-specific TLS fingerprints
func getBrowserClientHelloSpec(fingerprint Fingerprint, protocol ProtocolVersion) *utls.ClientHelloSpec {
	switch fingerprint {
	case ChromeFingerprint:
		return getChromeClientHelloSpec(protocol)
	case FirefoxFingerprint:
		return getFirefoxClientHelloSpec(protocol)
	case SafariFingerprint:
		return getSafariClientHelloSpec(protocol)
	case EdgeFingerprint:
		return getEdgeClientHelloSpec(protocol)
	default:
		return getChromeClientHelloSpec(protocol)
	}
}

// getChromeClientHelloSpec returns Chrome-specific TLS fingerprint
func getChromeClientHelloSpec(protocol ProtocolVersion) *utls.ClientHelloSpec {
	return &utls.ClientHelloSpec{
		TLSVersMax: utls.VersionTLS13,
		TLSVersMin: utls.VersionTLS12,
		CipherSuites: []uint16{
			utls.TLS_AES_128_GCM_SHA256,
			utls.TLS_AES_256_GCM_SHA384,
			utls.TLS_CHACHA20_POLY1305_SHA256,
			utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_RSA_WITH_AES_128_CBC_SHA,
			utls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		Extensions: []utls.TLSExtension{
			&utls.SNIExtension{},
			&utls.ExtendedMasterSecretExtension{},
			&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
			&utls.SupportedCurvesExtension{Curves: []utls.CurveID{
				utls.X25519,
				utls.CurveP256,
				utls.CurveP384,
			}},
			&utls.SupportedPointsExtension{SupportedPoints: []byte{0}},
			&utls.SessionTicketExtension{},
			&utls.ALPNExtension{
				AlpnProtocols: getALPNProtocols(protocol),
			},
			&utls.StatusRequestExtension{},
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
			&utls.SCTExtension{},
			&utls.KeyShareExtension{KeyShares: []utls.KeyShare{
				{Group: utls.X25519},
				{Group: utls.CurveP256},
			}},
			&utls.PSKKeyExchangeModesExtension{Modes: []uint8{
				utls.PskModeDHE,
			}},
			&utls.SupportedVersionsExtension{Versions: []uint16{
				utls.VersionTLS13,
				utls.VersionTLS12,
			}},
			&utls.ApplicationSettingsExtension{SupportedProtocols: getALPNProtocols(protocol)},
		},
	}
}

// getFirefoxClientHelloSpec returns Firefox-specific TLS fingerprint
func getFirefoxClientHelloSpec(protocol ProtocolVersion) *utls.ClientHelloSpec {
	return &utls.ClientHelloSpec{
		TLSVersMax: utls.VersionTLS13,
		TLSVersMin: utls.VersionTLS12,
		CipherSuites: []uint16{
			utls.TLS_AES_128_GCM_SHA256,
			utls.TLS_CHACHA20_POLY1305_SHA256,
			utls.TLS_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			utls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_RSA_WITH_AES_128_CBC_SHA,
			utls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		Extensions: []utls.TLSExtension{
			&utls.SNIExtension{},
			&utls.ExtendedMasterSecretExtension{},
			&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
			&utls.SupportedCurvesExtension{Curves: []utls.CurveID{
				utls.X25519,
				utls.CurveP256,
				utls.CurveP384,
				utls.CurveP521,
			}},
			&utls.SupportedPointsExtension{SupportedPoints: []byte{0}},
			&utls.SessionTicketExtension{},
			&utls.ALPNExtension{
				AlpnProtocols: getALPNProtocols(protocol),
			},
			&utls.StatusRequestExtension{},
			&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
				utls.ECDSAWithP256AndSHA256,
				utls.ECDSAWithP384AndSHA384,
				utls.ECDSAWithP521AndSHA512,
				utls.PSSWithSHA256,
				utls.PSSWithSHA384,
				utls.PSSWithSHA512,
				utls.PKCS1WithSHA256,
				utls.PKCS1WithSHA384,
				utls.PKCS1WithSHA512,
			}},
			&utls.KeyShareExtension{KeyShares: []utls.KeyShare{
				{Group: utls.X25519},
				{Group: utls.CurveP256},
			}},
			&utls.PSKKeyExchangeModesExtension{Modes: []uint8{
				utls.PskModeDHE,
			}},
			&utls.SupportedVersionsExtension{Versions: []uint16{
				utls.VersionTLS13,
				utls.VersionTLS12,
			}},
		},
	}
}

// getSafariClientHelloSpec returns Safari-specific TLS fingerprint
func getSafariClientHelloSpec(protocol ProtocolVersion) *utls.ClientHelloSpec {
	return &utls.ClientHelloSpec{
		TLSVersMax: utls.VersionTLS13,
		TLSVersMin: utls.VersionTLS12,
		CipherSuites: []uint16{
			utls.TLS_AES_128_GCM_SHA256,
			utls.TLS_AES_256_GCM_SHA384,
			utls.TLS_CHACHA20_POLY1305_SHA256,
			utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			utls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			utls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		},
		Extensions: []utls.TLSExtension{
			&utls.SNIExtension{},
			&utls.ExtendedMasterSecretExtension{},
			&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
			&utls.SupportedCurvesExtension{Curves: []utls.CurveID{
				utls.X25519,
				utls.CurveP256,
				utls.CurveP384,
				utls.CurveP521,
			}},
			&utls.SupportedPointsExtension{SupportedPoints: []byte{0}},
			&utls.SessionTicketExtension{},
			&utls.ALPNExtension{
				AlpnProtocols: getALPNProtocols(protocol),
			},
			&utls.StatusRequestExtension{},
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
			&utls.KeyShareExtension{KeyShares: []utls.KeyShare{
				{Group: utls.X25519},
			}},
			&utls.PSKKeyExchangeModesExtension{Modes: []uint8{
				utls.PskModeDHE,
			}},
			&utls.SupportedVersionsExtension{Versions: []uint16{
				utls.VersionTLS13,
				utls.VersionTLS12,
			}},
		},
	}
}

// getEdgeClientHelloSpec returns Edge-specific TLS fingerprint (similar to Chrome)
func getEdgeClientHelloSpec(protocol ProtocolVersion) *utls.ClientHelloSpec {
	return getChromeClientHelloSpec(protocol) // Edge uses Chromium base
}

// getALPNProtocols returns protocol list based on desired HTTP version
func getALPNProtocols(protocol ProtocolVersion) []string {
	switch protocol {
	case HTTP1_1:
		return []string{"http/1.1"}
	case HTTP2:
		return []string{"h2", "http/1.1"}
	case HTTPAuto:
		return []string{"h2", "http/1.1"}
	default:
		return []string{"http/1.1"}
	}
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
