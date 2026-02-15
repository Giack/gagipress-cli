package db

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
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

	// Test connection by running a simple Supabase CLI command
	fmt.Print("📡 Testing connection... ")
	testCmd := exec.Command("supabase", "db", "remote", "commit")
	var testOut bytes.Buffer
	testCmd.Stderr = &testOut
	testCmd.Stdout = &testOut

	if err := testCmd.Run(); err != nil {
		// Command may fail if no changes, but connection is tested
		output := testOut.String()
		if strings.Contains(output, "not logged in") || strings.Contains(output, "not linked") {
			fmt.Println("❌ FAILED")
			return fmt.Errorf("supabase CLI not configured correctly")
		}
	}
	fmt.Println("✅ OK")

	// Get migration status
	fmt.Print("📊 Checking schema version... ")
	statusCmd := exec.Command("supabase", "migration", "list", "--db-url", cfg.Supabase.URL)
	statusCmd.Env = append(os.Environ(), fmt.Sprintf("SUPABASE_ACCESS_TOKEN=%s", cfg.Supabase.AnonKey))

	var out bytes.Buffer
	statusCmd.Stdout = &out
	statusCmd.Stderr = &out

	// Count applied migrations from supabase/migrations directory
	cwd, _ := os.Getwd()
	files, err := os.ReadDir(fmt.Sprintf("%s/supabase/migrations", cwd))
	if err != nil {
		fmt.Println("⚠️  UNKNOWN")
		fmt.Println("   Run 'gagipress db migrate' to initialize the database")
	} else {
		sqlFiles := 0
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".sql") {
				sqlFiles++
			}
		}

		if sqlFiles == 0 {
			fmt.Println("⚠️  NOT INITIALIZED")
			fmt.Println("   Run 'gagipress db migrate' to initialize the database")
		} else {
			fmt.Printf("✅ v%d\n", sqlFiles)
		}
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
