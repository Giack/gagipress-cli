package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Gagipress CLI configuration",
	Long: `Interactive setup wizard to configure Gagipress CLI.

This command will guide you through:
  • Supabase connection setup
  • OpenAI API key configuration
  • Social media API credentials (optional)

Configuration is saved to ~/.gagipress/config.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Welcome to Gagipress CLI Setup")
	fmt.Println("==================================")
	fmt.Println()

	// Check if already configured
	if config.IsConfigured() {
		fmt.Print("⚠️  Configuration already exists. Overwrite? (y/N): ")
		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Setup cancelled.")
			return nil
		}
		fmt.Println()
	}

	cfg := &config.Config{}

	// Supabase Configuration
	fmt.Println("📦 Supabase Configuration")
	fmt.Println("-------------------------")
	fmt.Print("Supabase URL: ")
	supabaseURL, _ := reader.ReadString('\n')
	cfg.Supabase.URL = strings.TrimSpace(supabaseURL)

	fmt.Print("Supabase Anon Key: ")
	supabaseAnonKey, _ := reader.ReadString('\n')
	cfg.Supabase.AnonKey = strings.TrimSpace(supabaseAnonKey)

	fmt.Print("Supabase Service Key (optional): ")
	supabaseServiceKey, _ := reader.ReadString('\n')
	cfg.Supabase.ServiceKey = strings.TrimSpace(supabaseServiceKey)
	fmt.Println()

	// OpenAI Configuration
	fmt.Println("🤖 OpenAI Configuration")
	fmt.Println("-----------------------")
	fmt.Print("OpenAI API Key: ")
	openaiKey, _ := reader.ReadString('\n')
	cfg.OpenAI.APIKey = strings.TrimSpace(openaiKey)

	fmt.Print("OpenAI Model (default: gpt-4o-mini): ")
	openaiModel, _ := reader.ReadString('\n')
	openaiModel = strings.TrimSpace(openaiModel)
	if openaiModel == "" {
		openaiModel = "gpt-4o-mini"
	}
	cfg.OpenAI.Model = openaiModel
	fmt.Println()

	// Optional: Instagram Configuration
	fmt.Println("📸 Instagram Configuration (optional - press Enter to skip)")
	fmt.Println("-----------------------------------------------------------")
	fmt.Print("Instagram Access Token: ")
	instagramToken, _ := reader.ReadString('\n')
	cfg.Instagram.AccessToken = strings.TrimSpace(instagramToken)

	if cfg.Instagram.AccessToken != "" {
		fmt.Print("Instagram Account ID: ")
		instagramAccountID, _ := reader.ReadString('\n')
		cfg.Instagram.AccountID = strings.TrimSpace(instagramAccountID)
	}
	fmt.Println()

	// Optional: TikTok Configuration
	fmt.Println("🎵 TikTok Configuration (optional - press Enter to skip)")
	fmt.Println("---------------------------------------------------------")
	fmt.Print("TikTok Access Token: ")
	tiktokToken, _ := reader.ReadString('\n')
	cfg.TikTok.AccessToken = strings.TrimSpace(tiktokToken)

	if cfg.TikTok.AccessToken != "" {
		fmt.Print("TikTok Account ID: ")
		tiktokAccountID, _ := reader.ReadString('\n')
		cfg.TikTok.AccountID = strings.TrimSpace(tiktokAccountID)
	}
	fmt.Println()

	// Optional: Amazon KDP Configuration
	fmt.Println("📚 Amazon KDP Configuration (optional - press Enter to skip)")
	fmt.Println("-------------------------------------------------------------")
	fmt.Print("Amazon KDP Email: ")
	amazonEmail, _ := reader.ReadString('\n')
	cfg.Amazon.Email = strings.TrimSpace(amazonEmail)

	if cfg.Amazon.Email != "" {
		fmt.Print("Amazon KDP Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		cfg.Amazon.Password = string(passwordBytes)
		fmt.Println() // New line after password input
	}
	fmt.Println()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Configuration saved successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'gagipress db migrate' to create database schema")
	fmt.Println("  2. Run 'gagipress books add' to add your first book")
	fmt.Println("  3. Run 'gagipress generate ideas' to start generating content")
	fmt.Println()

	return nil
}
