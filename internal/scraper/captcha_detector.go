package scraper

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// CaptchaDetector detects and handles CAPTCHAs on web pages
type CaptchaDetector struct {
	solver   *CaptchaSolver
	jsEngine *JSEngine
	patterns map[CaptchaType]*CaptchaPattern
	enabled  bool
}

// CaptchaPattern defines how to detect and handle different CAPTCHA types
type CaptchaPattern struct {
	Type        CaptchaType
	Selectors   []string
	Keywords    []string
	URLPatterns []string
	Handler     func(ctx context.Context, cd *CaptchaDetector, pageURL string) (*CaptchaTask, error)
}

// CaptchaDetectionResult contains the results of CAPTCHA detection
type CaptchaDetectionResult struct {
	Found       bool          `json:"found"`
	Type        CaptchaType   `json:"type,omitempty"`
	SiteKey     string        `json:"site_key,omitempty"`
	PageURL     string        `json:"page_url"`
	ElementData string        `json:"element_data,omitempty"`
	SolveTime   time.Duration `json:"solve_time,omitempty"`
	Task        *CaptchaTask  `json:"task,omitempty"`
}

// NewCaptchaDetector creates a new CAPTCHA detector
func NewCaptchaDetector(solver *CaptchaSolver, jsEngine *JSEngine) *CaptchaDetector {
	cd := &CaptchaDetector{
		solver:   solver,
		jsEngine: jsEngine,
		enabled:  solver != nil,
		patterns: make(map[CaptchaType]*CaptchaPattern),
	}

	// Initialize CAPTCHA detection patterns
	cd.initializePatterns()

	return cd
}

// initializePatterns sets up detection patterns for different CAPTCHA types
func (cd *CaptchaDetector) initializePatterns() {
	// reCAPTCHA v2 pattern
	cd.patterns[RecaptchaV2] = &CaptchaPattern{
		Type: RecaptchaV2,
		Selectors: []string{
			".g-recaptcha",
			"[data-sitekey]",
			"iframe[src*='recaptcha']",
			"#recaptcha",
		},
		Keywords: []string{
			"recaptcha",
			"data-sitekey",
			"grecaptcha",
		},
		URLPatterns: []string{
			"recaptcha",
			"google.com/recaptcha",
		},
		Handler: func(ctx context.Context, detector *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
			return detector.handleRecaptchaV2(ctx, detector, pageURL)
		},
	}

	// reCAPTCHA v3 pattern
	cd.patterns[RecaptchaV3] = &CaptchaPattern{
		Type: RecaptchaV3,
		Selectors: []string{
			"script[src*='recaptcha/releases/'][src*='render']",
			"[data-site-key]",
		},
		Keywords: []string{
			"grecaptcha.execute",
			"recaptcha/api.js",
			"data-site-key",
		},
		URLPatterns: []string{
			"recaptcha/releases",
			"recaptcha/api.js",
		},
		Handler: func(ctx context.Context, detector *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
			return detector.handleRecaptchaV3(ctx, detector, pageURL)
		},
	}

	// hCaptcha pattern
	cd.patterns[HCaptcha] = &CaptchaPattern{
		Type: HCaptcha,
		Selectors: []string{
			".h-captcha",
			"[data-hcaptcha-sitekey]",
			"iframe[src*='hcaptcha']",
		},
		Keywords: []string{
			"hcaptcha",
			"data-hcaptcha-sitekey",
			"hcaptcha.com",
		},
		URLPatterns: []string{
			"hcaptcha",
			"hcaptcha.com",
		},
		Handler: func(ctx context.Context, detector *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
			return detector.handleHCaptcha(ctx, detector, pageURL)
		},
	}

	// Cloudflare Turnstile pattern
	cd.patterns[CloudflareCaptcha] = &CaptchaPattern{
		Type: CloudflareCaptcha,
		Selectors: []string{
			".cf-turnstile",
			"[data-cf-turnstile-sitekey]",
			"iframe[src*='turnstile']",
		},
		Keywords: []string{
			"turnstile",
			"cloudflare",
			"cf-turnstile",
		},
		URLPatterns: []string{
			"turnstile",
			"cloudflare.com",
		},
		Handler: func(ctx context.Context, detector *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
			return detector.handleTurnstile(ctx, detector, pageURL)
		},
	}

	// Image CAPTCHA pattern
	cd.patterns[ImageCaptcha] = &CaptchaPattern{
		Type: ImageCaptcha,
		Selectors: []string{
			"img[src*='captcha']",
			".captcha-image",
			"#captcha_image",
			"input[name*='captcha']",
		},
		Keywords: []string{
			"captcha",
			"verification",
			"security code",
		},
		Handler: func(ctx context.Context, detector *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
			return detector.handleImageCaptcha(ctx, detector, pageURL)
		},
	}
}

// DetectCaptcha detects CAPTCHAs on the current page
func (cd *CaptchaDetector) DetectCaptcha(ctx context.Context, pageURL string) (*CaptchaDetectionResult, error) {
	if !cd.enabled || cd.jsEngine == nil {
		return &CaptchaDetectionResult{Found: false, PageURL: pageURL}, nil
	}

	start := time.Now()
	result := &CaptchaDetectionResult{
		Found:   false,
		PageURL: pageURL,
	}

	// Get page content
	var pageHTML string
	err := chromedp.Run(cd.jsEngine.ctx, chromedp.InnerHTML("html", &pageHTML))
	if err != nil {
		return result, fmt.Errorf("failed to get page HTML: %w", err)
	}

	// Check each CAPTCHA pattern
	for captchaType, pattern := range cd.patterns {
		detected, err := cd.checkPattern(ctx, pattern, pageHTML, pageURL)
		if err != nil {
			continue // Continue checking other patterns
		}

		if detected {
			result.Found = true
			result.Type = captchaType
			result.SolveTime = time.Since(start)

			// Attempt to solve the CAPTCHA
			if cd.solver != nil {
				task, err := pattern.Handler(ctx, cd, pageURL)
				if err == nil && task != nil {
					result.Task = task
					result.SolveTime = time.Since(start)
				}
			}

			return result, nil
		}
	}

	return result, nil
}

// checkPattern checks if a specific CAPTCHA pattern is present
func (cd *CaptchaDetector) checkPattern(ctx context.Context, pattern *CaptchaPattern, pageHTML, pageURL string) (bool, error) {
	// Check HTML content for keywords
	lowerHTML := strings.ToLower(pageHTML)
	for _, keyword := range pattern.Keywords {
		if strings.Contains(lowerHTML, strings.ToLower(keyword)) {
			return true, nil
		}
	}

	// Check for URL patterns
	lowerURL := strings.ToLower(pageURL)
	for _, urlPattern := range pattern.URLPatterns {
		if strings.Contains(lowerURL, strings.ToLower(urlPattern)) {
			return true, nil
		}
	}

	// Check DOM selectors using JavaScript engine
	for _, selector := range pattern.Selectors {
		var elementExists bool
		err := chromedp.Run(cd.jsEngine.ctx,
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector('%s') !== null`, selector), &elementExists),
		)
		if err == nil && elementExists {
			return true, nil
		}
	}

	return false, nil
}

// handleRecaptchaV2 handles reCAPTCHA v2 solving
func (cd *CaptchaDetector) handleRecaptchaV2(ctx context.Context, cd2 *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
	// Extract site key
	var siteKey string
	selectors := []string{
		"[data-sitekey]",
		".g-recaptcha[data-sitekey]",
		"[data-site-key]",
	}

	for _, selector := range selectors {
		err := chromedp.Run(cd.jsEngine.ctx,
			chromedp.Evaluate(fmt.Sprintf(`
				const element = document.querySelector('%s');
				element ? (element.getAttribute('data-sitekey') || element.getAttribute('data-site-key')) : null
			`, selector), &siteKey),
		)
		if err == nil && siteKey != "" {
			break
		}
	}

	if siteKey == "" {
		return nil, fmt.Errorf("could not extract reCAPTCHA site key")
	}

	// Solve using the CAPTCHA service
	task, err := cd.solver.SolveRecaptchaV2(ctx, siteKey, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to solve reCAPTCHA v2: %w", err)
	}

	// Inject the solution into the page
	if task.Solution != "" {
		err = cd.injectRecaptchaV2Solution(ctx, task.Solution)
		if err != nil {
			return task, fmt.Errorf("failed to inject reCAPTCHA solution: %w", err)
		}
	}

	return task, nil
}

// handleRecaptchaV3 handles reCAPTCHA v3 solving
func (cd *CaptchaDetector) handleRecaptchaV3(ctx context.Context, cd2 *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
	// Extract site key and action
	var siteKey, action string

	// Try to get site key
	err := chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(`
			const scripts = document.querySelectorAll('script');
			let siteKey = '';
			for (let script of scripts) {
				const match = script.src.match(/render=([^&]+)/);
				if (match) {
					siteKey = match[1];
					break;
				}
			}
			siteKey;
		`, &siteKey),
	)
	if err != nil || siteKey == "" {
		return nil, fmt.Errorf("could not extract reCAPTCHA v3 site key")
	}

	// Try to extract action from page
	err = chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(`
			const scripts = document.querySelectorAll('script');
			let action = 'submit';
			for (let script of scripts) {
				const content = script.textContent;
				const match = content.match(/action['"]*:\s*['"]([^'"]+)['"]/);
				if (match) {
					action = match[1];
					break;
				}
			}
			action;
		`, &action),
	)
	if err != nil {
		action = "submit" // Default action
	}

	// Solve using the CAPTCHA service
	task, err := cd.solver.SolveRecaptchaV3(ctx, siteKey, pageURL, action, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to solve reCAPTCHA v3: %w", err)
	}

	// Inject the solution into the page
	if task.Solution != "" {
		err = cd.injectRecaptchaV3Solution(ctx, task.Solution, action)
		if err != nil {
			return task, fmt.Errorf("failed to inject reCAPTCHA v3 solution: %w", err)
		}
	}

	return task, nil
}

// handleHCaptcha handles hCaptcha solving
func (cd *CaptchaDetector) handleHCaptcha(ctx context.Context, cd2 *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
	// Extract site key
	var siteKey string
	err := chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(`
			const element = document.querySelector('[data-hcaptcha-sitekey], .h-captcha[data-sitekey]');
			element ? (element.getAttribute('data-hcaptcha-sitekey') || element.getAttribute('data-sitekey')) : null
		`, &siteKey),
	)
	if err != nil || siteKey == "" {
		return nil, fmt.Errorf("could not extract hCaptcha site key")
	}

	// Solve using the CAPTCHA service
	task, err := cd.solver.SolveHCaptcha(ctx, siteKey, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to solve hCaptcha: %w", err)
	}

	// Inject the solution into the page
	if task.Solution != "" {
		err = cd.injectHCaptchaSolution(ctx, task.Solution)
		if err != nil {
			return task, fmt.Errorf("failed to inject hCaptcha solution: %w", err)
		}
	}

	return task, nil
}

// handleTurnstile handles Cloudflare Turnstile solving
func (cd *CaptchaDetector) handleTurnstile(ctx context.Context, cd2 *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
	// Turnstile implementation would be similar to other CAPTCHAs
	return nil, fmt.Errorf("cloudflare Turnstile solving not implemented yet")
}

// handleImageCaptcha handles image-based CAPTCHA solving
func (cd *CaptchaDetector) handleImageCaptcha(ctx context.Context, cd2 *CaptchaDetector, pageURL string) (*CaptchaTask, error) {
	// Find CAPTCHA image
	var imageData string
	err := chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(`
			const img = document.querySelector('img[src*="captcha"], .captcha-image img, #captcha_image');
			if (img) {
				const canvas = document.createElement('canvas');
				const ctx = canvas.getContext('2d');
				canvas.width = img.naturalWidth;
				canvas.height = img.naturalHeight;
				ctx.drawImage(img, 0, 0);
				canvas.toDataURL('image/png');
			} else {
				null;
			}
		`, &imageData),
	)

	if err != nil || imageData == "" {
		return nil, fmt.Errorf("could not extract CAPTCHA image")
	}

	// Convert base64 to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(imageData, "data:image/png;base64,"))
	if err != nil {
		return nil, fmt.Errorf("failed to decode CAPTCHA image: %w", err)
	}

	// Solve using the CAPTCHA service
	task, err := cd.solver.SolveImageCaptcha(ctx, imageBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to solve image CAPTCHA: %w", err)
	}

	// Inject the solution into the input field
	if task.Solution != "" {
		err = cd.injectImageCaptchaSolution(ctx, task.Solution)
		if err != nil {
			return task, fmt.Errorf("failed to inject image CAPTCHA solution: %w", err)
		}
	}

	return task, nil
}

// Solution injection methods
func (cd *CaptchaDetector) injectRecaptchaV2Solution(ctx context.Context, solution string) error {
	return chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			if (typeof grecaptcha !== 'undefined') {
				const element = document.querySelector('.g-recaptcha, [data-sitekey]');
				if (element) {
					const widgetId = grecaptcha.render(element, {
						'sitekey': element.getAttribute('data-sitekey') || element.getAttribute('data-site-key'),
						'callback': function(token) { console.log('reCAPTCHA solved:', token); }
					});
					grecaptcha.getResponse = function() { return '%s'; };
				}
			}
			
			// Also set in hidden textarea if present
			const textarea = document.querySelector('[name="g-recaptcha-response"]');
			if (textarea) {
				textarea.value = '%s';
				textarea.dispatchEvent(new Event('change', { bubbles: true }));
			}
		`, solution, solution), nil),
	)
}

func (cd *CaptchaDetector) injectRecaptchaV3Solution(ctx context.Context, solution, action string) error {
	return chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			if (typeof grecaptcha !== 'undefined') {
				grecaptcha.ready(function() {
					window.captchaToken = '%s';
					
					// Override execute function
					const originalExecute = grecaptcha.execute;
					grecaptcha.execute = function(siteKey, options) {
						return Promise.resolve('%s');
					};
					
					// Trigger callback if present
					if (window.recaptchaCallback) {
						window.recaptchaCallback('%s');
					}
				});
			}
		`, solution, solution, solution), nil),
	)
}

func (cd *CaptchaDetector) injectHCaptchaSolution(ctx context.Context, solution string) error {
	return chromedp.Run(cd.jsEngine.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			// Set in hidden textarea
			const textarea = document.querySelector('[name="h-captcha-response"]');
			if (textarea) {
				textarea.value = '%s';
				textarea.dispatchEvent(new Event('change', { bubbles: true }));
			}
			
			// Override hcaptcha if present
			if (typeof hcaptcha !== 'undefined') {
				hcaptcha.getResponse = function() { return '%s'; };
			}
		`, solution, solution), nil),
	)
}

func (cd *CaptchaDetector) injectImageCaptchaSolution(ctx context.Context, solution string) error {
	// Try common CAPTCHA input field selectors
	selectors := []string{
		"input[name*='captcha']",
		"input[id*='captcha']",
		".captcha-input",
		"#captcha_code",
		"input[placeholder*='captcha']",
	}

	for _, selector := range selectors {
		err := chromedp.Run(cd.jsEngine.ctx,
			chromedp.Evaluate(fmt.Sprintf(`
				const input = document.querySelector('%s');
				if (input) {
					input.value = '%s';
					input.dispatchEvent(new Event('input', { bubbles: true }));
					input.dispatchEvent(new Event('change', { bubbles: true }));
					true;
				} else {
					false;
				}
			`, selector, solution), nil),
		)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not find CAPTCHA input field")
}

// Enable enables CAPTCHA detection and solving
func (cd *CaptchaDetector) Enable() {
	cd.enabled = true
}

// Disable disables CAPTCHA detection and solving
func (cd *CaptchaDetector) Disable() {
	cd.enabled = false
}

// IsEnabled returns whether CAPTCHA detection is enabled
func (cd *CaptchaDetector) IsEnabled() bool {
	return cd.enabled && cd.solver != nil
}

// SetSolver sets the CAPTCHA solver
func (cd *CaptchaDetector) SetSolver(solver *CaptchaSolver) {
	cd.solver = solver
	cd.enabled = solver != nil
}
