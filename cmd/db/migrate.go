package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/supabase"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w\nRun 'gagipress init' first", err)
	}

	// Validate Supabase config
	if cfg.Supabase.URL == "" || cfg.Supabase.AnonKey == "" {
		return fmt.Errorf("supabase configuration is incomplete\nRun 'gagipress init' to configure")
	}

	// Create Supabase client
	client, err := supabase.NewClient(&cfg.Supabase)
	if err != nil {
		return fmt.Errorf("failed to create supabase client: %w", err)
	}

	fmt.Println("🔗 Connected to Supabase")

	// Get current working directory to find migrations
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	migrationsDir := filepath.Join(cwd, "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsDir)
	}

	// Load migrations
	fmt.Println("📂 Loading migrations...")
	migrations, err := supabase.LoadMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if len(migrations) == 0 {
		fmt.Println("⚠️  No migration files found")
		return nil
	}

	fmt.Printf("Found %d migration(s)\n\n", len(migrations))

	// Get applied version
	appliedVersion, err := client.GetAppliedVersion()
	if err != nil {
		fmt.Printf("⚠️  Could not check applied migrations (table might not exist yet): %v\n", err)
		appliedVersion = 0
	}

	// Run pending migrations
	pendingCount := 0
	for _, migration := range migrations {
		if migration.Version <= appliedVersion {
			fmt.Printf("✓ Migration %d: %s (already applied)\n", migration.Version, migration.Description)
			continue
		}

		pendingCount++
		fmt.Printf("⏳ Applying migration %d: %s...\n", migration.Version, migration.Description)

		if err := client.RunMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		fmt.Printf("✅ Migration %d applied successfully\n", migration.Version)
	}

	if pendingCount == 0 {
		fmt.Println("\n✨ Database is up to date!")
	} else {
		fmt.Printf("\n✨ Successfully applied %d migration(s)!\n", pendingCount)
	}

	return nil
}
