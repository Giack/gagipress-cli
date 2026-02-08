package auth

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/social"
	"github.com/spf13/cobra"
)

var instagramCmd = &cobra.Command{
	Use:   "instagram",
	Short: "Test Instagram API connection",
	Long:  `Test Instagram Graph API authentication and connection.`,
	RunE:  runInstagramAuth,
}

func runInstagramAuth(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📸 Testing Instagram API connection...")

	// Create Instagram client
	client := social.NewInstagramClient(&cfg.Instagram)

	// Test connection
	fmt.Print("   Testing connection... ")
	if err := client.TestConnection(); err != nil {
		fmt.Println("❌ FAILED")
		fmt.Println("\n⚠️  Instagram OAuth flow not yet implemented.")
		fmt.Println("   This will be available in Week 3 of implementation.")
		fmt.Println("   Required steps:")
		fmt.Println("   1. Create Facebook App")
		fmt.Println("   2. Connect Instagram Business Account")
		fmt.Println("   3. Complete OAuth flow")
		return err
	}

	fmt.Println("✅ OK")
	fmt.Println("\n✅ Instagram API is configured correctly!")

	return nil
}
