package scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// CaptchaType represents different types of CAPTCHAs
type CaptchaType string

const (
	ImageCaptcha      CaptchaType = "image"
	RecaptchaV2       CaptchaType = "recaptchav2"
	RecaptchaV3       CaptchaType = "recaptchav3"
	HCaptcha          CaptchaType = "hcaptcha"
	FunCaptcha        CaptchaType = "funcaptcha"
	GeeTestCaptcha    CaptchaType = "geetest"
	CloudflareCaptcha CaptchaType = "turnstile"
)

// CaptchaService represents different CAPTCHA solving services
type CaptchaService string

const (
	TwoCaptchaService     CaptchaService = "2captcha"
	DeathByCaptchaService CaptchaService = "deathbycaptcha"
	AntiCaptchaService    CaptchaService = "anticaptcha"
	CapMonsterService     CaptchaService = "capmonster"
)

// CaptchaSolverConfig configures the CAPTCHA solver
type CaptchaSolverConfig struct {
	Service        CaptchaService `json:"service"`
	APIKey         string         `json:"api_key"`
	Timeout        time.Duration  `json:"timeout"`
	PollInterval   time.Duration  `json:"poll_interval"`
	MaxRetries     int            `json:"max_retries"`
	SoftID         string         `json:"soft_id"`         // For affiliate tracking
	Language       string         `json:"language"`        // Language preference
	MinScore       float64        `json:"min_score"`       // For reCAPTCHA v3
	DefaultTimeout time.Duration  `json:"default_timeout"` // Default solving timeout
}

// CaptchaTask represents a CAPTCHA solving task
type CaptchaTask struct {
	ID          string                 `json:"id"`
	Type        CaptchaType            `json:"type"`
	SiteKey     string                 `json:"site_key,omitempty"`
	PageURL     string                 `json:"page_url,omitempty"`
	ImageData   []byte                 `json:"image_data,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	SubmittedAt time.Time              `json:"submitted_at"`
	SolvedAt    *time.Time             `json:"solved_at,omitempty"`
	Solution    string                 `json:"solution,omitempty"`
	Cost        float64                `json:"cost,omitempty"`
	Status      string                 `json:"status"`
}

// CaptchaSolver handles CAPTCHA solving across multiple services
type CaptchaSolver struct {
	config     CaptchaSolverConfig
	httpClient *http.Client
	apiURLs    map[CaptchaService]string
}

// NewCaptchaSolver creates a new CAPTCHA solver
func NewCaptchaSolver(config CaptchaSolverConfig) *CaptchaSolver {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}
	if config.PollInterval == 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Language == "" {
		config.Language = "en"
	}
	if config.MinScore == 0 {
		config.MinScore = 0.3
	}
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 120 * time.Second
	}

	apiURLs := map[CaptchaService]string{
		TwoCaptchaService:     "https://2captcha.com",
		DeathByCaptchaService: "https://api.deathbycaptcha.com",
		AntiCaptchaService:    "https://api.anti-captcha.com",
		CapMonsterService:     "https://api.capmonster.cloud",
	}

	return &CaptchaSolver{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiURLs: apiURLs,
	}
}

// SolveImageCaptcha solves image-based CAPTCHAs
func (cs *CaptchaSolver) SolveImageCaptcha(ctx context.Context, imageData []byte, options map[string]interface{}) (*CaptchaTask, error) {
	task := &CaptchaTask{
		Type:        ImageCaptcha,
		ImageData:   imageData,
		SubmittedAt: time.Now(),
		Status:      "pending",
		Extra:       options,
	}

	switch cs.config.Service {
	case TwoCaptchaService:
		return cs.solve2CaptchaImage(ctx, task)
	case DeathByCaptchaService:
		return cs.solveDeathByCaptchaImage(ctx, task)
	case AntiCaptchaService:
		return cs.solveAntiCaptchaImage(ctx, task)
	case CapMonsterService:
		return cs.solveCapMonsterImage(ctx, task)
	default:
		return nil, fmt.Errorf("unsupported service for image CAPTCHA: %s", cs.config.Service)
	}
}

// SolveRecaptchaV2 solves reCAPTCHA v2 challenges
func (cs *CaptchaSolver) SolveRecaptchaV2(ctx context.Context, siteKey, pageURL string, options map[string]interface{}) (*CaptchaTask, error) {
	task := &CaptchaTask{
		Type:        RecaptchaV2,
		SiteKey:     siteKey,
		PageURL:     pageURL,
		SubmittedAt: time.Now(),
		Status:      "pending",
		Extra:       options,
	}

	switch cs.config.Service {
	case TwoCaptchaService:
		return cs.solve2CaptchaRecaptcha(ctx, task)
	case DeathByCaptchaService:
		return cs.solveDeathByCaptchaRecaptcha(ctx, task)
	case AntiCaptchaService:
		return cs.solveAntiCaptchaRecaptcha(ctx, task)
	case CapMonsterService:
		return cs.solveCapMonsterRecaptcha(ctx, task)
	default:
		return nil, fmt.Errorf("unsupported service for reCAPTCHA v2: %s", cs.config.Service)
	}
}

// SolveRecaptchaV3 solves reCAPTCHA v3 challenges
func (cs *CaptchaSolver) SolveRecaptchaV3(ctx context.Context, siteKey, pageURL, action string, options map[string]interface{}) (*CaptchaTask, error) {
	if options == nil {
		options = make(map[string]interface{})
	}
	options["action"] = action
	options["min_score"] = cs.config.MinScore

	task := &CaptchaTask{
		Type:        RecaptchaV3,
		SiteKey:     siteKey,
		PageURL:     pageURL,
		SubmittedAt: time.Now(),
		Status:      "pending",
		Extra:       options,
	}

	switch cs.config.Service {
	case TwoCaptchaService:
		return cs.solve2CaptchaRecaptchaV3(ctx, task)
	case AntiCaptchaService:
		return cs.solveAntiCaptchaRecaptchaV3(ctx, task)
	case CapMonsterService:
		return cs.solveCapMonsterRecaptchaV3(ctx, task)
	default:
		return nil, fmt.Errorf("unsupported service for reCAPTCHA v3: %s", cs.config.Service)
	}
}

// SolveHCaptcha solves hCaptcha challenges
func (cs *CaptchaSolver) SolveHCaptcha(ctx context.Context, siteKey, pageURL string, options map[string]interface{}) (*CaptchaTask, error) {
	task := &CaptchaTask{
		Type:        HCaptcha,
		SiteKey:     siteKey,
		PageURL:     pageURL,
		SubmittedAt: time.Now(),
		Status:      "pending",
		Extra:       options,
	}

	switch cs.config.Service {
	case TwoCaptchaService:
		return cs.solve2CaptchaHCaptcha(ctx, task)
	case AntiCaptchaService:
		return cs.solveAntiCaptchaHCaptcha(ctx, task)
	case CapMonsterService:
		return cs.solveCapMonsterHCaptcha(ctx, task)
	default:
		return nil, fmt.Errorf("unsupported service for hCaptcha: %s", cs.config.Service)
	}
}

// GetBalance returns the account balance for the CAPTCHA service
func (cs *CaptchaSolver) GetBalance(ctx context.Context) (float64, error) {
	switch cs.config.Service {
	case TwoCaptchaService:
		return cs.get2CaptchaBalance(ctx)
	case DeathByCaptchaService:
		return cs.getDeathByCaptchaBalance(ctx)
	case AntiCaptchaService:
		return cs.getAntiCaptchaBalance(ctx)
	case CapMonsterService:
		return cs.getCapMonsterBalance(ctx)
	default:
		return 0, fmt.Errorf("unsupported service: %s", cs.config.Service)
	}
}

// 2Captcha service implementations
func (cs *CaptchaSolver) solve2CaptchaImage(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// Submit the image CAPTCHA
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add API key
	writer.WriteField("key", cs.config.APIKey)
	writer.WriteField("method", "post")
	writer.WriteField("json", "1")

	// Add language if specified
	if cs.config.Language != "" {
		writer.WriteField("lang", cs.config.Language)
	}

	// Add soft ID if specified
	if cs.config.SoftID != "" {
		writer.WriteField("soft_id", cs.config.SoftID)
	}

	// Add the image file
	part, err := writer.CreateFormFile("file", "captcha.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	part.Write(task.ImageData)
	writer.Close()

	// Submit the task
	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[TwoCaptchaService]+"/in.php", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var submitResp struct {
		Status    int    `json:"status"`
		Request   string `json:"request"`
		ErrorText string `json:"error_text"`
	}

	if err := json.Unmarshal(body, &submitResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if submitResp.Status != 1 {
		return nil, fmt.Errorf("2captcha submission failed: %s", submitResp.ErrorText)
	}

	task.ID = submitResp.Request
	task.Status = "solving"

	// Poll for the result
	return cs.poll2CaptchaResult(ctx, task)
}

func (cs *CaptchaSolver) solve2CaptchaRecaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	data := map[string]string{
		"key":       cs.config.APIKey,
		"method":    "userrecaptcha",
		"googlekey": task.SiteKey,
		"pageurl":   task.PageURL,
		"json":      "1",
	}

	if cs.config.Language != "" {
		data["lang"] = cs.config.Language
	}

	if cs.config.SoftID != "" {
		data["soft_id"] = cs.config.SoftID
	}

	return cs.submit2CaptchaTask(ctx, task, data)
}

func (cs *CaptchaSolver) solve2CaptchaRecaptchaV3(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	data := map[string]string{
		"key":       cs.config.APIKey,
		"method":    "userrecaptcha",
		"version":   "v3",
		"googlekey": task.SiteKey,
		"pageurl":   task.PageURL,
		"json":      "1",
	}

	if action, ok := task.Extra["action"].(string); ok {
		data["action"] = action
	}

	if minScore, ok := task.Extra["min_score"].(float64); ok {
		data["min_score"] = fmt.Sprintf("%.1f", minScore)
	}

	return cs.submit2CaptchaTask(ctx, task, data)
}

func (cs *CaptchaSolver) solve2CaptchaHCaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	data := map[string]string{
		"key":     cs.config.APIKey,
		"method":  "hcaptcha",
		"sitekey": task.SiteKey,
		"pageurl": task.PageURL,
		"json":    "1",
	}

	return cs.submit2CaptchaTask(ctx, task, data)
}

func (cs *CaptchaSolver) submit2CaptchaTask(ctx context.Context, task *CaptchaTask, data map[string]string) (*CaptchaTask, error) {
	formData := strings.NewReader(cs.encodeFormData(data))

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[TwoCaptchaService]+"/in.php", formData)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var submitResp struct {
		Status    int    `json:"status"`
		Request   string `json:"request"`
		ErrorText string `json:"error_text"`
	}

	if err := json.Unmarshal(body, &submitResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if submitResp.Status != 1 {
		return nil, fmt.Errorf("2captcha submission failed: %s", submitResp.ErrorText)
	}

	task.ID = submitResp.Request
	task.Status = "solving"

	return cs.poll2CaptchaResult(ctx, task)
}

func (cs *CaptchaSolver) poll2CaptchaResult(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	ticker := time.NewTicker(cs.config.PollInterval)
	defer ticker.Stop()

	timeout := time.After(cs.config.Timeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for CAPTCHA solution")
		case <-ticker.C:
			url := fmt.Sprintf("%s/res.php?key=%s&action=get&id=%s&json=1",
				cs.apiURLs[TwoCaptchaService], cs.config.APIKey, task.ID)

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				continue
			}

			resp, err := cs.httpClient.Do(req)
			if err != nil {
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			var resultResp struct {
				Status    int    `json:"status"`
				Request   string `json:"request"`
				ErrorText string `json:"error_text"`
			}

			if err := json.Unmarshal(body, &resultResp); err != nil {
				continue
			}

			if resultResp.Status == 1 {
				now := time.Now()
				task.SolvedAt = &now
				task.Solution = resultResp.Request
				task.Status = "solved"
				return task, nil
			}

			if resultResp.Request != "CAPCHA_NOT_READY" {
				task.Status = "failed"
				return nil, fmt.Errorf("2captcha solving failed: %s", resultResp.ErrorText)
			}
		}
	}
}

func (cs *CaptchaSolver) get2CaptchaBalance(ctx context.Context) (float64, error) {
	url := fmt.Sprintf("%s/res.php?key=%s&action=getbalance&json=1",
		cs.apiURLs[TwoCaptchaService], cs.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var balanceResp struct {
		Status    int     `json:"status"`
		Request   float64 `json:"request"`
		ErrorText string  `json:"error_text"`
	}

	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if balanceResp.Status != 1 {
		return 0, fmt.Errorf("failed to get balance: %s", balanceResp.ErrorText)
	}

	return balanceResp.Request, nil
}

// DeathByCaptcha implementations (simplified for brevity)
func (cs *CaptchaSolver) solveDeathByCaptchaImage(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// DeathByCaptcha image submission
	payload := map[string]interface{}{
		"username":    cs.config.APIKey, // DeathByCaptcha uses username as API key
		"password":    cs.config.APIKey, // Can use same key for both
		"captchafile": task.ImageData,
		"type":        "0", // Image type
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DeathByCaptcha request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[DeathByCaptchaService]+"/api/captcha", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create DeathByCaptcha request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DeathByCaptcha request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeathByCaptcha response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DeathByCaptcha response: %w", err)
	}

	if status, ok := result["status"].(float64); !ok || status != 0 {
		return nil, fmt.Errorf("DeathByCaptcha submission failed: %s", string(body))
	}

	if captchaID, ok := result["captcha"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", captchaID)
		task.Status = "pending"

		// Poll for solution
		return cs.pollDeathByCaptchaResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid DeathByCaptcha response: %s", string(body))
}

func (cs *CaptchaSolver) solveDeathByCaptchaRecaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// DeathByCaptcha reCAPTCHA submission
	payload := map[string]interface{}{
		"username": cs.config.APIKey,
		"password": cs.config.APIKey,
		"type":     "4", // reCAPTCHA type
		"token_params": map[string]string{
			"googlekey": task.SiteKey,
			"pageurl":   task.PageURL,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DeathByCaptcha reCAPTCHA request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[DeathByCaptchaService]+"/api/captcha", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create DeathByCaptcha reCAPTCHA request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DeathByCaptcha reCAPTCHA request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DeathByCaptcha reCAPTCHA response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DeathByCaptcha reCAPTCHA response: %w", err)
	}

	if status, ok := result["status"].(float64); !ok || status != 0 {
		return nil, fmt.Errorf("DeathByCaptcha reCAPTCHA submission failed: %s", string(body))
	}

	if captchaID, ok := result["captcha"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", captchaID)
		task.Status = "pending"

		// Poll for solution
		return cs.pollDeathByCaptchaResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid DeathByCaptcha reCAPTCHA response: %s", string(body))
}

func (cs *CaptchaSolver) pollDeathByCaptchaResult(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	timeout := time.After(cs.config.Timeout)
	ticker := time.NewTicker(cs.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("DeathByCaptcha solving timed out after %v", cs.config.Timeout)
		case <-ticker.C:
			// Check result
			req, err := http.NewRequestWithContext(ctx, "GET",
				fmt.Sprintf("%s/api/captcha/%s", cs.apiURLs[DeathByCaptchaService], task.ID), nil)
			if err != nil {
				continue
			}

			resp, err := cs.httpClient.Do(req)
			if err != nil {
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				continue
			}

			if status, ok := result["status"].(float64); ok {
				if status == 0 && result["text"] != nil {
					// Solved
					task.Solution = result["text"].(string)
					task.Status = "solved"
					now := time.Now()
					task.SolvedAt = &now
					return task, nil
				} else if status == 255 {
					// Failed
					task.Status = "failed"
					return nil, fmt.Errorf("DeathByCaptcha solving failed: captcha marked as unsolvable")
				}
				// Still processing, continue polling
			}
		}
	}
}

func (cs *CaptchaSolver) getDeathByCaptchaBalance(ctx context.Context) (float64, error) {
	payload := map[string]string{
		"username": cs.config.APIKey,
		"password": cs.config.APIKey,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal DeathByCaptcha balance request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[DeathByCaptchaService]+"/api/user", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create DeathByCaptcha balance request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("DeathByCaptcha balance request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read DeathByCaptcha balance response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse DeathByCaptcha balance response: %w", err)
	}

	if balance, ok := result["balance"].(float64); ok {
		return balance, nil
	}

	return 0, fmt.Errorf("invalid DeathByCaptcha balance response: %s", string(body))
}

// Anti-Captcha implementations (simplified for brevity)
func (cs *CaptchaSolver) solveAntiCaptchaImage(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// Anti-Captcha task creation
	taskData := map[string]interface{}{
		"type": "ImageToTextTask",
		"body": task.ImageData, // Base64 encoded
	}

	if task.Extra != nil {
		if numeric, ok := task.Extra["numeric"].(bool); ok && numeric {
			taskData["numeric"] = 1
		}
		if minLength, ok := task.Extra["min_length"].(int); ok {
			taskData["minLength"] = minLength
		}
		if maxLength, ok := task.Extra["max_length"].(int); ok {
			taskData["maxLength"] = maxLength
		}
	}

	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
		"task":      taskData,
	}

	if cs.config.SoftID != "" {
		payload["softId"] = cs.config.SoftID
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anti-Captcha request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[AntiCaptchaService]+"/createTask", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anti-Captcha request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anti-Captcha request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Anti-Captcha response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Anti-Captcha response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return nil, fmt.Errorf("Anti-Captcha submission failed: %s", errorCode)
	}

	if taskID, ok := result["taskId"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", taskID)
		task.Status = "pending"

		// Poll for solution
		return cs.pollAntiCaptchaResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid Anti-Captcha response: %s", string(body))
}

func (cs *CaptchaSolver) solveAntiCaptchaRecaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// Anti-Captcha reCAPTCHA task creation
	taskData := map[string]interface{}{
		"type":       "NoCaptchaTaskProxyless",
		"websiteURL": task.PageURL,
		"websiteKey": task.SiteKey,
	}

	if task.Extra != nil {
		if userAgent, ok := task.Extra["user_agent"].(string); ok {
			taskData["userAgent"] = userAgent
		}
		if cookies, ok := task.Extra["cookies"].(string); ok {
			taskData["cookies"] = cookies
		}
	}

	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
		"task":      taskData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anti-Captcha reCAPTCHA request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[AntiCaptchaService]+"/createTask", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anti-Captcha reCAPTCHA request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anti-Captcha reCAPTCHA request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Anti-Captcha reCAPTCHA response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Anti-Captcha reCAPTCHA response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return nil, fmt.Errorf("Anti-Captcha reCAPTCHA submission failed: %s", errorCode)
	}

	if taskID, ok := result["taskId"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", taskID)
		task.Status = "pending"

		return cs.pollAntiCaptchaResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid Anti-Captcha reCAPTCHA response: %s", string(body))
}

func (cs *CaptchaSolver) solveAntiCaptchaRecaptchaV3(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// Anti-Captcha reCAPTCHA v3 task creation
	taskData := map[string]interface{}{
		"type":       "RecaptchaV3TaskProxyless",
		"websiteURL": task.PageURL,
		"websiteKey": task.SiteKey,
		"minScore":   cs.config.MinScore,
	}

	if task.Extra != nil {
		if action, ok := task.Extra["action"].(string); ok {
			taskData["pageAction"] = action
		}
	}

	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
		"task":      taskData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anti-Captcha reCAPTCHA v3 request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[AntiCaptchaService]+"/createTask", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anti-Captcha reCAPTCHA v3 request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anti-Captcha reCAPTCHA v3 request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Anti-Captcha reCAPTCHA v3 response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Anti-Captcha reCAPTCHA v3 response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return nil, fmt.Errorf("Anti-Captcha reCAPTCHA v3 submission failed: %s", errorCode)
	}

	if taskID, ok := result["taskId"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", taskID)
		task.Status = "pending"

		return cs.pollAntiCaptchaResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid Anti-Captcha reCAPTCHA v3 response: %s", string(body))
}

func (cs *CaptchaSolver) solveAntiCaptchaHCaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// Anti-Captcha hCaptcha task creation
	taskData := map[string]interface{}{
		"type":       "HCaptchaTaskProxyless",
		"websiteURL": task.PageURL,
		"websiteKey": task.SiteKey,
	}

	if task.Extra != nil {
		if userAgent, ok := task.Extra["user_agent"].(string); ok {
			taskData["userAgent"] = userAgent
		}
	}

	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
		"task":      taskData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anti-Captcha hCaptcha request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[AntiCaptchaService]+"/createTask", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anti-Captcha hCaptcha request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anti-Captcha hCaptcha request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Anti-Captcha hCaptcha response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Anti-Captcha hCaptcha response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return nil, fmt.Errorf("Anti-Captcha hCaptcha submission failed: %s", errorCode)
	}

	if taskID, ok := result["taskId"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", taskID)
		task.Status = "pending"

		return cs.pollAntiCaptchaResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid Anti-Captcha hCaptcha response: %s", string(body))
}

func (cs *CaptchaSolver) pollAntiCaptchaResult(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	timeout := time.After(cs.config.Timeout)
	ticker := time.NewTicker(cs.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("Anti-Captcha solving timed out after %v", cs.config.Timeout)
		case <-ticker.C:
			// Check result
			payload := map[string]interface{}{
				"clientKey": cs.config.APIKey,
				"taskId":    task.ID,
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[AntiCaptchaService]+"/getTaskResult", bytes.NewBuffer(jsonData))
			if err != nil {
				continue
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := cs.httpClient.Do(req)
			if err != nil {
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				continue
			}

			if errorID, ok := result["errorId"].(float64); ok && errorID == 0 {
				if status, exists := result["status"].(string); exists {
					if status == "ready" {
						// Solved
						if solution, ok := result["solution"].(map[string]interface{}); ok {
							if token, exists := solution["gRecaptchaResponse"].(string); exists {
								task.Solution = token
							} else if text, exists := solution["text"].(string); exists {
								task.Solution = text
							}
							task.Status = "solved"
							now := time.Now()
							task.SolvedAt = &now
							return task, nil
						}
					} else if status == "processing" {
						// Still processing, continue polling
						continue
					}
				}
			} else {
				// Error occurred
				task.Status = "failed"
				return nil, fmt.Errorf("Anti-Captcha solving failed")
			}
		}
	}
}

func (cs *CaptchaSolver) getAntiCaptchaBalance(ctx context.Context) (float64, error) {
	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal Anti-Captcha balance request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[AntiCaptchaService]+"/getBalance", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create Anti-Captcha balance request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Anti-Captcha balance request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read Anti-Captcha balance response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse Anti-Captcha balance response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return 0, fmt.Errorf("Anti-Captcha balance check failed: %s", errorCode)
	}

	if balance, ok := result["balance"].(float64); ok {
		return balance, nil
	}

	return 0, fmt.Errorf("invalid Anti-Captcha balance response: %s", string(body))
}

func (cs *CaptchaSolver) getCapMonsterBalance(ctx context.Context) (float64, error) {
	// CapMonster uses the same API as Anti-Captcha
	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal CapMonster balance request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[CapMonsterService]+"/getBalance", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create CapMonster balance request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("CapMonster balance request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read CapMonster balance response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse CapMonster balance response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return 0, fmt.Errorf("CapMonster balance check failed: %s", errorCode)
	}

	if balance, ok := result["balance"].(float64); ok {
		return balance, nil
	}

	return 0, fmt.Errorf("invalid CapMonster balance response: %s", string(body))
}

// CapMonster service implementations (uses same API as Anti-Captcha)
func (cs *CaptchaSolver) solveCapMonsterImage(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// CapMonster uses the same API as Anti-Captcha for image solving
	return cs.solveCapMonsterGeneric(ctx, task, "ImageToTextTask")
}

func (cs *CaptchaSolver) solveCapMonsterRecaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// CapMonster reCAPTCHA v2
	return cs.solveCapMonsterGeneric(ctx, task, "NoCaptchaTaskProxyless")
}

func (cs *CaptchaSolver) solveCapMonsterRecaptchaV3(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// CapMonster reCAPTCHA v3
	return cs.solveCapMonsterGeneric(ctx, task, "RecaptchaV3TaskProxyless")
}

func (cs *CaptchaSolver) solveCapMonsterHCaptcha(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// CapMonster hCaptcha
	return cs.solveCapMonsterGeneric(ctx, task, "HCaptchaTaskProxyless")
}

func (cs *CaptchaSolver) solveCapMonsterGeneric(ctx context.Context, task *CaptchaTask, taskType string) (*CaptchaTask, error) {
	// Generic CapMonster task creation (same as Anti-Captcha)
	taskData := map[string]interface{}{
		"type": taskType,
	}

	// Configure task based on type
	switch taskType {
	case "ImageToTextTask":
		taskData["body"] = task.ImageData
		if task.Extra != nil {
			if numeric, ok := task.Extra["numeric"].(bool); ok && numeric {
				taskData["numeric"] = 1
			}
		}
	case "NoCaptchaTaskProxyless":
		taskData["websiteURL"] = task.PageURL
		taskData["websiteKey"] = task.SiteKey
	case "RecaptchaV3TaskProxyless":
		taskData["websiteURL"] = task.PageURL
		taskData["websiteKey"] = task.SiteKey
		taskData["minScore"] = cs.config.MinScore
		if task.Extra != nil {
			if action, ok := task.Extra["action"].(string); ok {
				taskData["pageAction"] = action
			}
		}
	case "HCaptchaTaskProxyless":
		taskData["websiteURL"] = task.PageURL
		taskData["websiteKey"] = task.SiteKey
	}

	payload := map[string]interface{}{
		"clientKey": cs.config.APIKey,
		"task":      taskData,
	}

	if cs.config.SoftID != "" {
		payload["softId"] = cs.config.SoftID
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CapMonster request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[CapMonsterService]+"/createTask", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create CapMonster request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CapMonster request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CapMonster response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse CapMonster response: %w", err)
	}

	if errorID, ok := result["errorId"].(float64); !ok || errorID != 0 {
		errorCode := "unknown"
		if code, exists := result["errorCode"].(string); exists {
			errorCode = code
		}
		return nil, fmt.Errorf("CapMonster submission failed: %s", errorCode)
	}

	if taskID, ok := result["taskId"].(float64); ok {
		task.ID = fmt.Sprintf("%.0f", taskID)
		task.Status = "pending"

		// Poll for solution using CapMonster API
		return cs.pollCapMonsterResult(ctx, task)
	}

	return nil, fmt.Errorf("invalid CapMonster response: %s", string(body))
}

func (cs *CaptchaSolver) pollCapMonsterResult(ctx context.Context, task *CaptchaTask) (*CaptchaTask, error) {
	// CapMonster uses same polling API as Anti-Captcha
	timeout := time.After(cs.config.Timeout)
	ticker := time.NewTicker(cs.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("CapMonster solving timed out after %v", cs.config.Timeout)
		case <-ticker.C:
			payload := map[string]interface{}{
				"clientKey": cs.config.APIKey,
				"taskId":    task.ID,
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				continue
			}

			req, err := http.NewRequestWithContext(ctx, "POST", cs.apiURLs[CapMonsterService]+"/getTaskResult", bytes.NewBuffer(jsonData))
			if err != nil {
				continue
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := cs.httpClient.Do(req)
			if err != nil {
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				continue
			}

			if errorID, ok := result["errorId"].(float64); ok && errorID == 0 {
				if status, exists := result["status"].(string); exists {
					if status == "ready" {
						// Solved
						if solution, ok := result["solution"].(map[string]interface{}); ok {
							if token, exists := solution["gRecaptchaResponse"].(string); exists {
								task.Solution = token
							} else if text, exists := solution["text"].(string); exists {
								task.Solution = text
							}
							task.Status = "solved"
							now := time.Now()
							task.SolvedAt = &now
							return task, nil
						}
					} else if status == "processing" {
						continue
					}
				}
			} else {
				task.Status = "failed"
				return nil, fmt.Errorf("CapMonster solving failed")
			}
		}
	}
}

// Utility functions
func (cs *CaptchaSolver) encodeFormData(data map[string]string) string {
	var parts []string
	for key, value := range data {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, "&")
}
