package test

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/ai"
	"github.com/spf13/cobra"
)

var (
	headless bool
)

var geminiCmd = &cobra.Command{
	Use:   "gemini [prompt]",
	Short: "Test Gemini browser automation",
	Long: `Test Gemini browser automation by sending a prompt and retrieving the response.

Example:
  gagipress test gemini "Ciao, come stai?"
  gagipress test gemini --headless=false "Scrivi una storia breve"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGeminiTest,
}

func init() {
	geminiCmd.Flags().BoolVar(&headless, "headless", true, "Run browser in headless mode")
}

func runGeminiTest(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	fmt.Println("🤖 Testing Gemini browser automation...")
	fmt.Printf("   Prompt: %s\n", prompt)
	fmt.Printf("   Headless: %v\n", headless)
	fmt.Println()

	// Create Gemini client
	client := ai.NewGeminiClient(headless)

	// Send prompt
	fmt.Println("🌐 Launching browser...")
	fmt.Println("📝 Sending prompt to Gemini...")
	fmt.Println("⏳ Waiting for response (this may take a few seconds)...")
	fmt.Println()

	response, err := client.GenerateText(prompt)
	if err != nil {
		return fmt.Errorf("Gemini test failed: %w", err)
	}

	fmt.Println("✅ Response received!")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(response)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}
