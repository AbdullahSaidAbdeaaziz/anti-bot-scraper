package scraper

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/chromedp/chromedp"
)

// BehaviorType represents different behavioral simulation modes
type BehaviorType string

const (
	NormalBehavior     BehaviorType = "normal"
	CautiousBehavior   BehaviorType = "cautious"
	AggressiveBehavior BehaviorType = "aggressive"
	RandomBehavior     BehaviorType = "random"
)

// HumanBehaviorConfig configures behavioral simulation
type HumanBehaviorConfig struct {
	Enabled             bool          `json:"enabled"`
	BehaviorType        BehaviorType  `json:"behavior_type"`
	MinDelay            time.Duration `json:"min_delay"`
	MaxDelay            time.Duration `json:"max_delay"`
	MouseMovement       bool          `json:"mouse_movement"`
	ScrollSimulation    bool          `json:"scroll_simulation"`
	TypingDelay         bool          `json:"typing_delay"`
	PageLoadWait        bool          `json:"page_load_wait"`
	RandomScrolling     bool          `json:"random_scrolling"`
	RandomClicks        bool          `json:"random_clicks"`
	ViewportVariation   bool          `json:"viewport_variation"`
	TabSwitchSimulation bool          `json:"tab_switch_simulation"`
}

// HumanBehaviorSimulator handles human-like behavior simulation
type HumanBehaviorSimulator struct {
	config     HumanBehaviorConfig
	random     *rand.Rand
	lastAction time.Time
}

// MousePosition represents mouse coordinates
type MousePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ScrollAction represents a scroll event
type ScrollAction struct {
	Direction string `json:"direction"` // "up", "down", "left", "right"
	Distance  int    `json:"distance"`
	Duration  int    `json:"duration"` // milliseconds
}

// TypingPattern represents realistic typing simulation
type TypingPattern struct {
	Text               string        `json:"text"`
	BaseDelay          time.Duration `json:"base_delay"`
	VariationDelay     time.Duration `json:"variation_delay"`
	PauseOnMistakes    bool          `json:"pause_on_mistakes"`
	MistakeProbability float64       `json:"mistake_probability"`
}

// NewHumanBehaviorSimulator creates a new behavior simulator
func NewHumanBehaviorSimulator(config HumanBehaviorConfig) *HumanBehaviorSimulator {
	return &HumanBehaviorSimulator{
		config:     config,
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
		lastAction: time.Now(),
	}
}

// GetDefaultBehaviorConfig returns sensible defaults for behavioral simulation
func GetDefaultBehaviorConfig() HumanBehaviorConfig {
	return HumanBehaviorConfig{
		Enabled:             true,
		BehaviorType:        NormalBehavior,
		MinDelay:            500 * time.Millisecond,
		MaxDelay:            2000 * time.Millisecond,
		MouseMovement:       true,
		ScrollSimulation:    true,
		TypingDelay:         true,
		PageLoadWait:        true,
		RandomScrolling:     false,
		RandomClicks:        false,
		ViewportVariation:   true,
		TabSwitchSimulation: false,
	}
}

// ApplyBehaviorType adjusts config based on behavior type
func (hbs *HumanBehaviorSimulator) ApplyBehaviorType(behaviorType BehaviorType) {
	switch behaviorType {
	case CautiousBehavior:
		hbs.config.MinDelay = 1000 * time.Millisecond
		hbs.config.MaxDelay = 5000 * time.Millisecond
		hbs.config.MouseMovement = true
		hbs.config.ScrollSimulation = true
		hbs.config.RandomScrolling = true
		hbs.config.PageLoadWait = true
	case AggressiveBehavior:
		hbs.config.MinDelay = 100 * time.Millisecond
		hbs.config.MaxDelay = 800 * time.Millisecond
		hbs.config.MouseMovement = false
		hbs.config.ScrollSimulation = false
		hbs.config.RandomScrolling = false
		hbs.config.PageLoadWait = false
	case RandomBehavior:
		hbs.config.MinDelay = time.Duration(hbs.random.Intn(2000)) * time.Millisecond
		hbs.config.MaxDelay = time.Duration(hbs.random.Intn(5000)+1000) * time.Millisecond
		hbs.config.MouseMovement = hbs.random.Float64() > 0.5
		hbs.config.ScrollSimulation = hbs.random.Float64() > 0.3
		hbs.config.RandomScrolling = hbs.random.Float64() > 0.7
		hbs.config.RandomClicks = hbs.random.Float64() > 0.8
	case NormalBehavior:
		// Keep default values
	}
}

// SimulateHumanDelay adds realistic delays between actions
func (hbs *HumanBehaviorSimulator) SimulateHumanDelay() {
	if !hbs.config.Enabled {
		return
	}

	// Calculate time since last action
	timeSinceLastAction := time.Since(hbs.lastAction)

	// Generate random delay within configured range
	minDelay := hbs.config.MinDelay
	maxDelay := hbs.config.MaxDelay

	// Adjust delay based on recent activity (fatigue simulation)
	if timeSinceLastAction < 5*time.Second {
		// Recent activity, slightly longer delays
		minDelay = time.Duration(float64(minDelay) * 1.2)
		maxDelay = time.Duration(float64(maxDelay) * 1.3)
	}

	delay := minDelay + time.Duration(hbs.random.Int63n(int64(maxDelay-minDelay)))

	// Add some natural variation (simulate fatigue, distractions)
	if hbs.random.Float64() < 0.1 {
		// 10% chance of longer pause (distraction simulation)
		delay = delay + time.Duration(hbs.random.Intn(3000))*time.Millisecond
	}

	time.Sleep(delay)
	hbs.lastAction = time.Now()
}

// SimulateMouseMovement performs realistic mouse movement
func (hbs *HumanBehaviorSimulator) SimulateMouseMovement(ctx context.Context, targetX, targetY int) chromedp.Action {
	if !hbs.config.Enabled || !hbs.config.MouseMovement {
		return chromedp.MouseClickXY(float64(targetX), float64(targetY))
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Get current mouse position (simulate from a random starting point)
		startX := hbs.random.Intn(1920)
		startY := hbs.random.Intn(1080)

		// Calculate movement steps for smooth animation
		steps := 5 + hbs.random.Intn(10) // 5-15 steps
		stepX := float64(targetX-startX) / float64(steps)
		stepY := float64(targetY-startY) / float64(steps)

		// Simulate human-like mouse movement with curves
		for i := 0; i < steps; i++ {
			currentX := float64(startX) + stepX*float64(i)
			currentY := float64(startY) + stepY*float64(i)

			// Add slight curve and jitter to movement
			if i > 0 && i < steps-1 {
				currentX += float64(hbs.random.Intn(5) - 2) // ±2 pixel jitter
				currentY += float64(hbs.random.Intn(5) - 2)
			}

			// Move mouse to intermediate position
			if err := chromedp.MouseClickXY(currentX, currentY).Do(ctx); err != nil {
				return err
			}

			// Small delay between movement steps
			time.Sleep(time.Duration(10+hbs.random.Intn(30)) * time.Millisecond)
		}

		// Final click at target position
		return chromedp.MouseClickXY(float64(targetX), float64(targetY)).Do(ctx)
	})
}

// SimulateScrolling performs realistic scrolling behavior
func (hbs *HumanBehaviorSimulator) SimulateScrolling(ctx context.Context, scrollActions []ScrollAction) chromedp.Action {
	if !hbs.config.Enabled || !hbs.config.ScrollSimulation {
		return nil
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, action := range scrollActions {
			// Simulate gradual scrolling
			totalDistance := action.Distance
			steps := 3 + hbs.random.Intn(5) // 3-8 scroll steps
			stepDistance := totalDistance / steps

			for i := 0; i < steps; i++ {
				var scrollJS string
				switch action.Direction {
				case "down":
					scrollJS = fmt.Sprintf("window.scrollBy(0, %d);", stepDistance)
				case "up":
					scrollJS = fmt.Sprintf("window.scrollBy(0, -%d);", stepDistance)
				case "right":
					scrollJS = fmt.Sprintf("window.scrollBy(%d, 0);", stepDistance)
				case "left":
					scrollJS = fmt.Sprintf("window.scrollBy(-%d, 0);", stepDistance)
				}

				if err := chromedp.Evaluate(scrollJS, nil).Do(ctx); err != nil {
					return err
				}

				// Delay between scroll steps
				stepDelay := time.Duration(action.Duration/steps) * time.Millisecond
				time.Sleep(stepDelay + time.Duration(hbs.random.Intn(50))*time.Millisecond)
			}
		}
		return nil
	})
}

// SimulateTyping performs realistic typing with human-like delays and mistakes
func (hbs *HumanBehaviorSimulator) SimulateTyping(ctx context.Context, selector, text string) chromedp.Action {
	if !hbs.config.Enabled || !hbs.config.TypingDelay {
		return chromedp.SendKeys(selector, text)
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Clear existing text first
		if err := chromedp.Clear(selector).Do(ctx); err != nil {
			return err
		}

		// Type character by character with realistic delays
		for i, char := range text {
			// Calculate typing delay (faster for common keys, slower for special chars)
			baseDelay := 50 + hbs.random.Intn(100) // 50-150ms base

			// Adjust for character type
			charStr := string(char)
			if charStr == " " {
				baseDelay += 20 // Slightly longer for space
			} else if charStr >= "A" && charStr <= "Z" {
				baseDelay += 30 // Longer for shift+key
			}

			// Simulate occasional typos (then correction)
			if hbs.random.Float64() < 0.02 && i > 0 { // 2% typo chance
				// Type wrong character
				wrongChar := string(rune('a' + hbs.random.Intn(26)))
				if err := chromedp.SendKeys(selector, wrongChar).Do(ctx); err != nil {
					return err
				}

				// Pause (realize mistake)
				time.Sleep(time.Duration(200+hbs.random.Intn(300)) * time.Millisecond)

				// Backspace to correct
				if err := chromedp.KeyEvent("\b").Do(ctx); err != nil {
					return err
				}

				// Brief pause before correct character
				time.Sleep(time.Duration(100+hbs.random.Intn(200)) * time.Millisecond)
			}

			// Type the actual character
			if err := chromedp.SendKeys(selector, charStr).Do(ctx); err != nil {
				return err
			}

			// Delay before next character
			time.Sleep(time.Duration(baseDelay) * time.Millisecond)
		}

		return nil
	})
}

// SimulatePageLoad waits for page load with human-like behavior
func (hbs *HumanBehaviorSimulator) SimulatePageLoad(ctx context.Context) chromedp.Action {
	if !hbs.config.Enabled || !hbs.config.PageLoadWait {
		return nil
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Wait for initial load
		if err := chromedp.WaitReady("body").Do(ctx); err != nil {
			return err
		}

		// Additional human-like delay after page load
		postLoadDelay := 1000 + hbs.random.Intn(2000) // 1-3 seconds
		time.Sleep(time.Duration(postLoadDelay) * time.Millisecond)

		// Simulate reading time based on page content length
		var contentLength int
		if err := chromedp.Evaluate("document.body.innerText.length", &contentLength).Do(ctx); err == nil {
			// Simulate reading: ~200 words per minute, ~5 chars per word
			readingTime := time.Duration(contentLength/1000) * time.Second
			if readingTime > 10*time.Second {
				readingTime = 10 * time.Second // Cap at 10 seconds
			}
			time.Sleep(readingTime)
		}

		return nil
	})
}

// SimulateRandomActivity performs random page interactions
func (hbs *HumanBehaviorSimulator) SimulateRandomActivity(ctx context.Context) chromedp.Action {
	if !hbs.config.Enabled || !hbs.config.RandomScrolling {
		return nil
	}

	return chromedp.ActionFunc(func(ctx context.Context) error {
		activities := hbs.random.Intn(3) + 1 // 1-3 random activities

		for i := 0; i < activities; i++ {
			activity := hbs.random.Intn(4)

			switch activity {
			case 0: // Random scroll
				direction := []string{"down", "up"}[hbs.random.Intn(2)]
				distance := 100 + hbs.random.Intn(300)
				scrollAction := ScrollAction{
					Direction: direction,
					Distance:  distance,
					Duration:  300 + hbs.random.Intn(500),
				}
				if err := hbs.SimulateScrolling(ctx, []ScrollAction{scrollAction}).Do(ctx); err != nil {
					continue
				}

			case 1: // Random mouse movement
				x := hbs.random.Intn(1920)
				y := hbs.random.Intn(1080)
				if err := hbs.SimulateMouseMovement(ctx, x, y).Do(ctx); err != nil {
					continue
				}

			case 2: // Pause (simulate thinking/reading)
				pauseTime := 500 + hbs.random.Intn(2000)
				time.Sleep(time.Duration(pauseTime) * time.Millisecond)

			case 3: // Focus simulation (tab key)
				if err := chromedp.KeyEvent("\t").Do(ctx); err != nil {
					continue
				}
			}

			// Delay between activities
			time.Sleep(time.Duration(200+hbs.random.Intn(800)) * time.Millisecond)
		}

		return nil
	})
}

// GetRandomViewport returns a random but realistic viewport size
func (hbs *HumanBehaviorSimulator) GetRandomViewport() (int, int) {
	if !hbs.config.Enabled || !hbs.config.ViewportVariation {
		return 1920, 1080 // Default
	}

	// Common viewport sizes with some variation
	viewports := [][]int{
		{1920, 1080}, {1366, 768}, {1536, 864}, {1440, 900},
		{1280, 720}, {1600, 900}, {1024, 768}, {1280, 1024},
	}

	selected := viewports[hbs.random.Intn(len(viewports))]

	// Add slight variation (±50 pixels)
	width := selected[0] + hbs.random.Intn(101) - 50
	height := selected[1] + hbs.random.Intn(101) - 50

	// Ensure minimum size
	if width < 800 {
		width = 800
	}
	if height < 600 {
		height = 600
	}

	return width, height
}

// GenerateRealisticDelay creates delays based on user behavior patterns
func (hbs *HumanBehaviorSimulator) GenerateRealisticDelay(actionType string) time.Duration {
	if !hbs.config.Enabled {
		return 0
	}

	baseDelays := map[string]time.Duration{
		"click":     200 * time.Millisecond,
		"type":      100 * time.Millisecond,
		"scroll":    300 * time.Millisecond,
		"navigate":  2000 * time.Millisecond,
		"form_fill": 1000 * time.Millisecond,
		"search":    1500 * time.Millisecond,
	}

	baseDelay := baseDelays[actionType]
	if baseDelay == 0 {
		baseDelay = 500 * time.Millisecond
	}

	// Add human-like variation (±50% of base delay)
	variation := time.Duration(float64(baseDelay) * (0.5 * (hbs.random.Float64()*2 - 1)))

	return baseDelay + variation
}
