package auth

import (
	"github.com/spf13/cobra"
)

// AuthCmd represents the auth command group
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Test API authentication and connections",
	Long: `Test authentication and connectivity for various API integrations:
  - OpenAI API
  - Instagram Graph API (requires OAuth)
  - TikTok API (requires OAuth)`,
}

func init() {
	AuthCmd.AddCommand(openaiCmd)
	AuthCmd.AddCommand(instagramCmd)
	AuthCmd.AddCommand(tiktokCmd)
}
