package db

import (
	"fmt"
	"os"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/supabase"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check database connection and migration status",
	Long: `Verify connection to Supabase and show which migrations have been applied.

This command will:
  1. Test the connection to Supabase
  2. Display current schema version
  3. Show database info

Example:
  gagipress db status`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStatus(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runStatus() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'gagipress init' first", err)
	}

	// Validate Supabase config
	if cfg.Supabase.URL == "" || cfg.Supabase.AnonKey == "" {
		return fmt.Errorf("supabase configuration is incomplete\nRun 'gagipress init' to configure")
	}

	fmt.Println("🔍 Checking database status...")
	fmt.Println()

	// Create Supabase client
	client, err := supabase.NewClient(&cfg.Supabase)
	if err != nil {
		return fmt.Errorf("failed to create supabase client: %w", err)
	}

	// Test connection
	fmt.Print("📡 Testing connection... ")
	if err := client.TestConnection(); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("connection test failed: %w", err)
	}
	fmt.Println("✅ OK")

	// Get applied version
	fmt.Print("📊 Checking schema version... ")
	version, err := client.GetAppliedVersion()
	if err != nil {
		fmt.Println("⚠️  UNKNOWN")
		fmt.Printf("   Could not determine version: %v\n", err)
		fmt.Println("   Run 'gagipress db migrate' to initialize the database")
	} else if version == 0 {
		fmt.Println("⚠️  NOT INITIALIZED")
		fmt.Println("   Run 'gagipress db migrate' to initialize the database")
	} else {
		fmt.Printf("✅ v%d\n", version)
	}

	fmt.Println()
	fmt.Println("📋 Configuration:")
	fmt.Printf("   URL: %s\n", cfg.Supabase.URL)
	fmt.Printf("   Using: %s\n", func() string {
		if cfg.Supabase.ServiceKey != "" {
			return "Service Key (admin)"
		}
		return "Anon Key (standard)"
	}())

	fmt.Println()
	fmt.Println("✨ Database status check complete!")

	return nil
}
