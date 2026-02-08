package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// GeminiClient handles browser automation for Gemini
type GeminiClient struct {
	headless bool
}

// NewGeminiClient creates a new Gemini browser automation client
func NewGeminiClient(headless bool) *GeminiClient {
	return &GeminiClient{
		headless: headless,
	}
}

// GenerateText sends a prompt to Gemini and retrieves the response
func (g *GeminiClient) GenerateText(prompt string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Setup browser options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", g.headless),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	// Create browser context
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var response string

	// Automate browser interaction
	err := chromedp.Run(browserCtx,
		// Navigate to Gemini
		chromedp.Navigate("https://gemini.google.com"),

		// Wait for page to load
		chromedp.WaitVisible(`textarea[placeholder*="Enter a prompt"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),

		// Enter prompt
		chromedp.SendKeys(`textarea[placeholder*="Enter a prompt"]`, prompt, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),

		// Submit (press Enter)
		chromedp.SendKeys(`textarea[placeholder*="Enter a prompt"]`, "\n", chromedp.ByQuery),

		// Wait for response (this is a placeholder - actual selector depends on Gemini's UI)
		chromedp.Sleep(5*time.Second),
		chromedp.WaitVisible(`div[data-test-id="conversation-turn-2"]`, chromedp.ByQuery),

		// Extract response text
		chromedp.Text(`div[data-test-id="conversation-turn-2"]`, &response, chromedp.ByQuery),
	)

	if err != nil {
		return "", fmt.Errorf("browser automation failed: %w", err)
	}

	if response == "" {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return response, nil
}

// TestConnection tests the Gemini browser automation
func (g *GeminiClient) TestConnection() error {
	_, err := g.GenerateText("Say 'OK' if you can read this.")
	return err
}
