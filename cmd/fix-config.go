package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fixConfigCmd migrates old config format to new format
var fixConfigCmd = &cobra.Command{
	Use:   "fix-config",
	Short: "Migrate config file to use correct field names",
	Long: `Migrates ~/.gagipress/config.yaml from camelCase to snake_case field names.

This fixes the issue where 'gagipress init' created a config file that
couldn't be loaded by other commands due to field name mismatch.

This command:
  1. Reads the old config (with any field name format)
  2. Writes it back with correct snake_case field names
  3. Validates the migrated config

Run this once after upgrading if you're experiencing config issues.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runFixConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fixConfigCmd)
}

func runFixConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get home directory: %w", err)
	}

	configFile := filepath.Join(home, ".gagipress", "config.yaml")

	// Check if config exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\nRun 'gagipress init' first", configFile)
	}

	fmt.Println("🔧 Migrating config file...")
	fmt.Printf("   Config: %s\n\n", configFile)

	// Load config with lenient parsing (accept any field names)
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Try to unmarshal with mapstructure tags
	var cfg config.Config
	unmarshalErr := viper.Unmarshal(&cfg)

	// Check if we need manual migration (either unmarshal failed OR validation fails)
	needsManualMigration := unmarshalErr != nil || cfg.Validate() != nil

	if needsManualMigration {
		// If unmarshal/validation fails, try manual mapping for known old formats
		fmt.Println("⚠️  Detected old config format, performing manual migration...")

		// Manually map old camelCase fields
		cfg.Supabase.URL = viper.GetString("supabase.url")
		cfg.Supabase.AnonKey = viper.GetString("supabase.anonkey")
		if cfg.Supabase.AnonKey == "" {
			cfg.Supabase.AnonKey = viper.GetString("supabase.anon_key")
		}
		cfg.Supabase.ServiceKey = viper.GetString("supabase.servicekey")
		if cfg.Supabase.ServiceKey == "" {
			cfg.Supabase.ServiceKey = viper.GetString("supabase.service_key")
		}

		cfg.OpenAI.APIKey = viper.GetString("openai.apikey")
		if cfg.OpenAI.APIKey == "" {
			cfg.OpenAI.APIKey = viper.GetString("openai.api_key")
		}
		cfg.OpenAI.Model = viper.GetString("openai.model")

		cfg.Instagram.AccessToken = viper.GetString("instagram.accesstoken")
		if cfg.Instagram.AccessToken == "" {
			cfg.Instagram.AccessToken = viper.GetString("instagram.access_token")
		}
		cfg.Instagram.AccountID = viper.GetString("instagram.accountid")
		if cfg.Instagram.AccountID == "" {
			cfg.Instagram.AccountID = viper.GetString("instagram.account_id")
		}

		cfg.TikTok.AccessToken = viper.GetString("tiktok.accesstoken")
		if cfg.TikTok.AccessToken == "" {
			cfg.TikTok.AccessToken = viper.GetString("tiktok.access_token")
		}
		cfg.TikTok.AccountID = viper.GetString("tiktok.accountid")
		if cfg.TikTok.AccountID == "" {
			cfg.TikTok.AccountID = viper.GetString("tiktok.account_id")
		}

		cfg.Amazon.Email = viper.GetString("amazon.email")
		cfg.Amazon.Password = viper.GetString("amazon.password")
	}

	// Validate the loaded config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Backup old config
	backupFile := configFile + ".backup"
	fmt.Printf("📦 Creating backup: %s\n", backupFile)
	if err := os.Rename(configFile, backupFile); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Save with new format
	fmt.Println("💾 Saving migrated config...")
	if err := config.Save(&cfg); err != nil {
		// Restore backup on failure
		os.Rename(backupFile, configFile)
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Verify we can load it back
	viper.Reset()
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to verify migrated config: %w", err)
	}

	var verifyConfig config.Config
	if err := viper.Unmarshal(&verifyConfig); err != nil {
		return fmt.Errorf("failed to unmarshal migrated config: %w", err)
	}

	fmt.Println("\n✅ Config migration successful!")
	fmt.Printf("   Backup saved to: %s\n", backupFile)
	fmt.Println("\nYou can now run commands like 'gagipress db migrate'")

	return nil
}
