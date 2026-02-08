package auth

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/ai"
	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/spf13/cobra"
)

var openaiCmd = &cobra.Command{
	Use:   "openai",
	Short: "Test OpenAI API connection",
	Long:  `Test OpenAI API authentication and connection by sending a simple request.`,
	RunE:  runOpenAIAuth,
}

func runOpenAIAuth(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if API key is configured
	if cfg.OpenAI.APIKey == "" {
		return fmt.Errorf("OpenAI API key not configured. Run 'gagipress init' first")
	}

	fmt.Println("🔑 Testing OpenAI API connection...")
	fmt.Printf("   Model: %s\n", cfg.OpenAI.Model)

	// Create OpenAI client
	client := ai.NewOpenAIClient(&cfg.OpenAI)

	// Test connection
	fmt.Print("   Sending test request... ")
	if err := client.TestConnection(); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("OpenAI connection test failed: %w", err)
	}

	fmt.Println("✅ OK")
	fmt.Println("\n✅ OpenAI API is configured correctly!")

	return nil
}
