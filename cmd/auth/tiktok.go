package auth

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/social"
	"github.com/spf13/cobra"
)

var tiktokCmd = &cobra.Command{
	Use:   "tiktok",
	Short: "Test TikTok API connection",
	Long:  `Test TikTok API authentication and connection.`,
	RunE:  runTikTokAuth,
}

func runTikTokAuth(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🎵 Testing TikTok API connection...")

	// Create TikTok client
	client := social.NewTikTokClient(&cfg.TikTok)

	// Test connection
	fmt.Print("   Testing connection... ")
	if err := client.TestConnection(); err != nil {
		fmt.Println("❌ FAILED")
		fmt.Println("\n⚠️  TikTok OAuth flow not yet implemented.")
		fmt.Println("   This will be available in Week 3 of implementation.")
		fmt.Println("   Required steps:")
		fmt.Println("   1. Create TikTok Developer Account")
		fmt.Println("   2. Create App in TikTok Developer Portal")
		fmt.Println("   3. Complete OAuth flow")
		return err
	}

	fmt.Println("✅ OK")
	fmt.Println("\n✅ TikTok API is configured correctly!")

	return nil
}
