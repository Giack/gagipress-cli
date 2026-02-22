package db

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long: `Run pending database migrations to create or update the schema.

This command will:
  1. Load all migration files from the migrations/ directory
  2. Check which migrations have already been applied
  3. Apply any pending migrations in order
  4. Update the schema_version table

Example:
  gagipress db migrate`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMigrate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runMigrate() error {
	// Load configuration to validate Supabase setup
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'gagipress init' first", err)
	}

	// Validate Supabase config
	if cfg.Supabase.URL == "" || cfg.Supabase.AnonKey == "" {
		return fmt.Errorf("supabase configuration is incomplete\nRun 'gagipress init' to configure")
	}

	// Get current working directory to find migrations
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Sync migrations from migrations/ to supabase/migrations/
	migrationsDir := filepath.Join(cwd, "migrations")
	supabaseMigrationsDir := filepath.Join(cwd, "supabase", "migrations")

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsDir)
	}

	// Ensure supabase/migrations directory exists
	if err := os.MkdirAll(supabaseMigrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create supabase/migrations directory: %w", err)
	}

	// Copy migration files
	fmt.Println("📂 Syncing migrations to supabase/migrations/...")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		srcPath := filepath.Join(migrationsDir, file.Name())
		dstPath := filepath.Join(supabaseMigrationsDir, file.Name())

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write migration file %s: %w", file.Name(), err)
		}
	}

	fmt.Println("✅ Migrations synced")
	fmt.Println("\n🚀 Running Supabase CLI migration...")

	// Use Supabase CLI to apply migrations
	cmd := exec.Command("supabase", "db", "push", "--yes")
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("supabase db push failed: %w", err)
	}

	fmt.Println("\n✨ Database migrations applied successfully!")
	return nil
}
