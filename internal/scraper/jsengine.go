package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// JSEngine represents a JavaScript execution engine using headless Chrome
type JSEngine struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// JSEngineConfig holds configuration for the JavaScript engine
type JSEngineConfig struct {
	Enabled      bool
	Timeout      time.Duration
	UserAgent    string
	Viewport     Viewport
	Headless     bool
	NoImages     bool
	NoJavaScript bool // Paradoxically, to disable JS in the browser if needed
}

// Viewport represents browser viewport dimensions
type Viewport struct {
	Width  int64
	Height int64
}

// JSResponse represents a response from JavaScript-enabled scraping
type JSResponse struct {
	HTML       string
	StatusCode int
	URL        string
	Headers    map[string]string
	Cookies    map[string]string
	Console    []string // Console logs
	Errors     []string // JavaScript errors
}

// NewJSEngine creates a new JavaScript engine instance
func NewJSEngine(config JSEngineConfig) (*JSEngine, error) {
	if !config.Enabled {
		return &JSEngine{enable: false}, nil
	}

	// Set default configuration
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Viewport.Width == 0 {
		config.Viewport.Width = 1920
	}
	if config.Viewport.Height == 0 {
		config.Viewport.Height = 1080
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	// Create Chrome options
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.UserAgent(config.UserAgent),
		chromedp.WindowSize(int(config.Viewport.Width), int(config.Viewport.Height)),
	}

	// Add headless option
	if config.Headless {
		opts = append(opts, chromedp.Headless)
	}

	// Disable images if requested
	if config.NoImages {
		opts = append(opts, chromedp.Flag("blink-settings", "imagesEnabled=false"))
	}

	// Create allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create chrome context
	ctx, _ := chromedp.NewContext(allocCtx)

	return &JSEngine{
		ctx:    ctx,
		cancel: cancel,
		enable: true,
	}, nil
}

// Close cleans up the JavaScript engine
func (js *JSEngine) Close() error {
	if js.cancel != nil {
		js.cancel()
	}
	return nil
}

// Scrape performs JavaScript-enabled scraping
func (js *JSEngine) Scrape(url string, config JSEngineConfig) (*JSResponse, error) {
	if !js.enable {
		return nil, fmt.Errorf("JavaScript engine is disabled")
	}

	ctx, cancel := context.WithTimeout(js.ctx, config.Timeout)
	defer cancel()

	var html string

	// Navigate and wait for page to load
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Allow dynamic content to load
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scrape with JavaScript engine: %w", err)
	}

	// Get additional page information
	var currentURL string
	err = chromedp.Run(ctx, chromedp.Location(&currentURL))
	if err != nil {
		currentURL = url // Fallback to original URL
	}

	return &JSResponse{
		HTML:       html,
		StatusCode: 200, // chromedp doesn't easily expose status codes
		URL:        currentURL,
		Headers:    make(map[string]string), // Would need additional work to capture
		Cookies:    make(map[string]string), // Would need additional work to capture
		Console:    []string{},              // Simplified - no console capture for now
		Errors:     []string{},              // Simplified - no error capture for now
	}, nil
}

// ExecuteJS executes custom JavaScript code on the page
func (js *JSEngine) ExecuteJS(url string, jsCode string, config JSEngineConfig) (*JSResponse, error) {
	if !js.enable {
		return nil, fmt.Errorf("JavaScript engine is disabled")
	}

	ctx, cancel := context.WithTimeout(js.ctx, config.Timeout)
	defer cancel()

	var html string
	var result interface{}

	// Navigate, execute custom JS, and get results
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(jsCode, &result),
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to execute JavaScript: %w", err)
	}

	var currentURL string
	err = chromedp.Run(ctx, chromedp.Location(&currentURL))
	if err != nil {
		currentURL = url
	}

	response := &JSResponse{
		HTML:       html,
		StatusCode: 200,
		URL:        currentURL,
		Headers:    make(map[string]string),
		Cookies:    make(map[string]string),
		Console:    []string{},
		Errors:     []string{},
	}

	// Add JS execution result to console messages
	if result != nil {
		response.Console = append(response.Console, fmt.Sprintf("[JS_RESULT] %v", result))
	}

	return response, nil
}

// WaitForElement waits for a specific element to appear on the page
func (js *JSEngine) WaitForElement(url string, selector string, config JSEngineConfig) (*JSResponse, error) {
	if !js.enable {
		return nil, fmt.Errorf("JavaScript engine is disabled")
	}

	ctx, cancel := context.WithTimeout(js.ctx, config.Timeout)
	defer cancel()

	var html string

	// Navigate and wait for specific element
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to wait for element '%s': %w", selector, err)
	}

	var currentURL string
	err = chromedp.Run(ctx, chromedp.Location(&currentURL))
	if err != nil {
		currentURL = url
	}

	return &JSResponse{
		HTML:       html,
		StatusCode: 200,
		URL:        currentURL,
		Headers:    make(map[string]string),
		Cookies:    make(map[string]string),
		Console:    []string{},
		Errors:     []string{},
	}, nil
}

// SimulateHumanBehavior performs human-like interactions
func (js *JSEngine) SimulateHumanBehavior(url string, config JSEngineConfig) (*JSResponse, error) {
	if !js.enable {
		return nil, fmt.Errorf("JavaScript engine is disabled")
	}

	ctx, cancel := context.WithTimeout(js.ctx, config.Timeout)
	defer cancel()

	var html string

	// Navigate and simulate human behavior
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),

		// Simulate random mouse movement
		chromedp.MouseClickXY(100, 100),
		chromedp.Sleep(500*time.Millisecond),

		// Simulate scrolling
		chromedp.Evaluate(`window.scrollTo(0, 200)`, nil),
		chromedp.Sleep(1*time.Second),

		chromedp.Evaluate(`window.scrollTo(0, 400)`, nil),
		chromedp.Sleep(1*time.Second),

		// Wait a bit more for dynamic content
		chromedp.Sleep(2*time.Second),

		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to simulate human behavior: %w", err)
	}

	var currentURL string
	err = chromedp.Run(ctx, chromedp.Location(&currentURL))
	if err != nil {
		currentURL = url
	}

	return &JSResponse{
		HTML:       html,
		StatusCode: 200,
		URL:        currentURL,
		Headers:    make(map[string]string),
		Cookies:    make(map[string]string),
		Console:    []string{},
		Errors:     []string{},
	}, nil
}
