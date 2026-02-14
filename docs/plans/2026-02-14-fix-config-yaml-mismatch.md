# Config Serialization/Deserialization Fix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix YAML field name mismatch between serialization (camelCase) and deserialization (snake_case with mapstructure tags)

**Architecture:** Add `yaml:` struct tags to all config structs to match the `mapstructure:` tags, ensuring Viper writes and reads the same field names. This is a non-breaking change that fixes existing config files via migration.

**Tech Stack:** Go 1.24, Viper, YAML

---

## Root Cause Summary

**Problem:** Viper serialization and deserialization use different conventions:
- **Serialization** (`viper.WriteConfigAs`): Uses Go field names → lowercase (e.g., `anonkey`)
- **Deserialization** (`viper.Unmarshal`): Uses `mapstructure` tags (e.g., `anon_key`)

**Impact:** After running `gagipress init`, the config file exists but commands fail because Viper can't unmarshal the YAML fields.

**Example:**
```yaml
# Written by Save()
supabase:
    url: https://...
    anonkey: xxx        # ← lowercase field name
    servicekey: xxx

# Expected by Load()
supabase:
    url: https://...
    anon_key: xxx       # ← snake_case from mapstructure tag
    service_key: xxx
```

---

## Task 1: Add YAML Tags to Config Structs

**Files:**
- Modify: `internal/config/config.go:20-49`

**Step 1: Add yaml tags to SupabaseConfig**

Modify `internal/config/config.go` lines 20-25:

```go
// SupabaseConfig holds Supabase connection details
type SupabaseConfig struct {
	URL        string `mapstructure:"url" yaml:"url"`
	AnonKey    string `mapstructure:"anon_key" yaml:"anon_key"`
	ServiceKey string `mapstructure:"service_key" yaml:"service_key"`
}
```

**Step 2: Add yaml tags to OpenAIConfig**

Modify `internal/config/config.go` lines 27-31:

```go
// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey string `mapstructure:"api_key" yaml:"api_key"`
	Model  string `mapstructure:"model" yaml:"model"`
}
```

**Step 3: Add yaml tags to InstagramConfig**

Modify `internal/config/config.go` lines 33-37:

```go
// InstagramConfig holds Instagram API configuration
type InstagramConfig struct {
	AccessToken string `mapstructure:"access_token" yaml:"access_token"`
	AccountID   string `mapstructure:"account_id" yaml:"account_id"`
}
```

**Step 4: Add yaml tags to TikTokConfig**

Modify `internal/config/config.go` lines 39-43:

```go
// TikTokConfig holds TikTok API configuration
type TikTokConfig struct {
	AccessToken string `mapstructure:"access_token" yaml:"access_token"`
	AccountID   string `mapstructure:"account_id" yaml:"account_id"`
}
```

**Step 5: Add yaml tags to AmazonConfig**

Modify `internal/config/config.go` lines 45-49:

```go
// AmazonConfig holds Amazon KDP credentials
type AmazonConfig struct {
	Email    string `mapstructure:"email" yaml:"email"`
	Password string `mapstructure:"password" yaml:"password"`
}
```

**Step 6: Verify changes**

Run: `cat internal/config/config.go | grep -A 3 "type.*Config struct"`

Expected: All config structs now have both `mapstructure` and `yaml` tags

**Step 7: Commit**

```bash
git add internal/config/config.go
git commit -m "fix: add yaml tags to config structs for consistent serialization"
```

---

## Task 2: Test Config Save/Load Roundtrip

**Files:**
- Create: `internal/config/config_test.go`

**Step 1: Write failing test for config roundtrip**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigSaveLoadRoundtrip(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create test config
	original := &Config{
		Supabase: SupabaseConfig{
			URL:        "https://test.supabase.co",
			AnonKey:    "test-anon-key",
			ServiceKey: "test-service-key",
		},
		OpenAI: OpenAIConfig{
			APIKey: "test-openai-key",
			Model:  "gpt-4o-mini",
		},
		Instagram: InstagramConfig{
			AccessToken: "test-ig-token",
			AccountID:   "test-ig-account",
		},
		TikTok: TikTokConfig{
			AccessToken: "test-tt-token",
			AccountID:   "test-tt-account",
		},
		Amazon: AmazonConfig{
			Email:    "test@example.com",
			Password: "test-password",
		},
	}

	// Save config using Viper (same as Save() does)
	viper.Set("supabase", original.Supabase)
	viper.Set("openai", original.OpenAI)
	viper.Set("instagram", original.Instagram)
	viper.Set("tiktok", original.TikTok)
	viper.Set("amazon", original.Amazon)

	if err := viper.WriteConfigAs(configFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Reset Viper and load config back
	viper.Reset()
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Unmarshal into new config
	var loaded Config
	if err := viper.Unmarshal(&loaded); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify all fields match
	if loaded.Supabase.URL != original.Supabase.URL {
		t.Errorf("Supabase.URL mismatch: got %q, want %q", loaded.Supabase.URL, original.Supabase.URL)
	}
	if loaded.Supabase.AnonKey != original.Supabase.AnonKey {
		t.Errorf("Supabase.AnonKey mismatch: got %q, want %q", loaded.Supabase.AnonKey, original.Supabase.AnonKey)
	}
	if loaded.Supabase.ServiceKey != original.Supabase.ServiceKey {
		t.Errorf("Supabase.ServiceKey mismatch: got %q, want %q", loaded.Supabase.ServiceKey, original.Supabase.ServiceKey)
	}
	if loaded.OpenAI.APIKey != original.OpenAI.APIKey {
		t.Errorf("OpenAI.APIKey mismatch: got %q, want %q", loaded.OpenAI.APIKey, original.OpenAI.APIKey)
	}
	if loaded.OpenAI.Model != original.OpenAI.Model {
		t.Errorf("OpenAI.Model mismatch: got %q, want %q", loaded.OpenAI.Model, original.OpenAI.Model)
	}
	if loaded.Instagram.AccessToken != original.Instagram.AccessToken {
		t.Errorf("Instagram.AccessToken mismatch: got %q, want %q", loaded.Instagram.AccessToken, original.Instagram.AccessToken)
	}
	if loaded.Instagram.AccountID != original.Instagram.AccountID {
		t.Errorf("Instagram.AccountID mismatch: got %q, want %q", loaded.Instagram.AccountID, original.Instagram.AccountID)
	}
	if loaded.TikTok.AccessToken != original.TikTok.AccessToken {
		t.Errorf("TikTok.AccessToken mismatch: got %q, want %q", loaded.TikTok.AccessToken, original.TikTok.AccessToken)
	}
	if loaded.TikTok.AccountID != original.TikTok.AccountID {
		t.Errorf("TikTok.AccountID mismatch: got %q, want %q", loaded.TikTok.AccountID, original.TikTok.AccountID)
	}
	if loaded.Amazon.Email != original.Amazon.Email {
		t.Errorf("Amazon.Email mismatch: got %q, want %q", loaded.Amazon.Email, original.Amazon.Email)
	}
	if loaded.Amazon.Password != original.Amazon.Password {
		t.Errorf("Amazon.Password mismatch: got %q, want %q", loaded.Amazon.Password, original.Amazon.Password)
	}
}

func TestYAMLFieldNames(t *testing.T) {
	// Setup temp directory
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create test config
	cfg := &Config{
		Supabase: SupabaseConfig{
			URL:        "https://test.supabase.co",
			AnonKey:    "test-key",
			ServiceKey: "test-service",
		},
	}

	// Save config
	viper.Set("supabase", cfg.Supabase)
	if err := viper.WriteConfigAs(configFile); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read raw YAML to verify field names
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	yamlContent := string(content)

	// Verify snake_case field names (not camelCase)
	if !contains(yamlContent, "anon_key") {
		t.Errorf("YAML should contain 'anon_key', got:\n%s", yamlContent)
	}
	if !contains(yamlContent, "service_key") {
		t.Errorf("YAML should contain 'service_key', got:\n%s", yamlContent)
	}

	// Verify NOT camelCase
	if contains(yamlContent, "anonkey") || contains(yamlContent, "anonKey") {
		t.Errorf("YAML should NOT contain camelCase 'anonkey', got:\n%s", yamlContent)
	}
	if contains(yamlContent, "servicekey") || contains(yamlContent, "serviceKey") {
		t.Errorf("YAML should NOT contain camelCase 'servicekey', got:\n%s", yamlContent)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
```

**Step 2: Run test to verify it fails (before yaml tags)**

Run: `mise exec -- go test ./internal/config -v -run TestConfigSaveLoadRoundtrip`

Expected: FAIL with empty field values (if yaml tags not added yet)

**Step 3: Run test to verify it passes (after Task 1)**

Run: `mise exec -- go test ./internal/config -v`

Expected: PASS (both tests)

**Step 4: Commit**

```bash
git add internal/config/config_test.go
git commit -m "test: add config save/load roundtrip tests"
```

---

## Task 3: Create Config Migration Tool

**Files:**
- Create: `cmd/fix-config.go` (temporary migration command)

**Step 1: Create migration command**

Create `cmd/fix-config.go`:

```go
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
	if err := viper.Unmarshal(&cfg); err != nil {
		// If unmarshal fails, try manual mapping for known old formats
		fmt.Println("⚠️  Standard unmarshal failed, trying manual migration...")

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
```

**Step 2: Add command to root**

Verify `cmd/root.go` includes the new command (it's auto-registered via `init()`)

**Step 3: Test migration command**

Run: `mise exec -- go run main.go fix-config`

Expected: Should migrate `~/.gagipress/config.yaml` from camelCase to snake_case

**Step 4: Verify db commands work**

Run: `mise exec -- go run main.go db status`

Expected: Should connect successfully (not "config incomplete")

**Step 5: Commit**

```bash
git add cmd/fix-config.go
git commit -m "feat: add fix-config migration command"
```

---

## Task 4: Update Documentation

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

**Step 1: Add troubleshooting section to README**

Add to `README.md` after the "Configuration" section:

```markdown
### Troubleshooting

**Commands fail with "Run 'gagipress init' first" even after running init:**

This was a bug in earlier versions where the config file was created with incorrect field names.

Fix:
```bash
gagipress fix-config
```

This will migrate your config file to the correct format. You only need to run this once.
```

**Step 2: Update CLAUDE.md**

Add to `CLAUDE.md` under "Important Constraints":

```markdown
### Config File Format

**IMPORTANT**: All config structs use both `mapstructure` and `yaml` tags:

```go
type SupabaseConfig struct {
    URL        string `mapstructure:"url" yaml:"url"`
    AnonKey    string `mapstructure:"anon_key" yaml:"anon_key"`
    ServiceKey string `mapstructure:"service_key" yaml:"service_key"`
}
```

**Why both tags?**
- `mapstructure`: Used by Viper for deserialization (`Unmarshal`)
- `yaml`: Used by Viper for serialization (`WriteConfigAs`)

Without both tags, field names mismatch between save/load, causing config to appear empty.

**Historical Bug**: Versions before 2026-02-14 only had `mapstructure` tags, causing Viper to write camelCase fields (`anonkey`) but expect snake_case (`anon_key`) when reading. The `fix-config` command migrates old configs.
```

**Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: add config troubleshooting and yaml tag explanation"
```

---

## Task 5: Verification and Cleanup

**Files:**
- Test: All config-related code

**Step 1: Run all config tests**

Run: `mise exec -- go test ./internal/config -v`

Expected: All tests PASS

**Step 2: Test full workflow**

```bash
# Backup and remove existing config
mv ~/.gagipress/config.yaml ~/.gagipress/config.yaml.old

# Run init
mise exec -- go run main.go init
# (Enter test values)

# Verify db commands work
mise exec -- go run main.go db status
```

Expected: No "config incomplete" error

**Step 3: Verify YAML format**

Run: `cat ~/.gagipress/config.yaml | grep -E "(anon_key|service_key|api_key)"`

Expected: Should see snake_case field names, not camelCase

**Step 4: Test migration on old config**

```bash
# Restore old config
mv ~/.gagipress/config.yaml.old ~/.gagipress/config.yaml

# Run migration
mise exec -- go run main.go fix-config

# Verify it works
mise exec -- go run main.go db status
```

Expected: Migration successful, db status works

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: verify config fix complete"
```

---

## Success Criteria

- [ ] All config structs have both `mapstructure` and `yaml` tags
- [ ] Test `TestConfigSaveLoadRoundtrip` passes
- [ ] Test `TestYAMLFieldNames` passes
- [ ] `gagipress init` creates config with snake_case fields
- [ ] `gagipress db status` works after init (no "config incomplete")
- [ ] `gagipress fix-config` migrates old configs successfully
- [ ] Documentation updated with troubleshooting info
- [ ] All tests passing

---

## Rollback Plan

If issues occur:

1. Restore backup config: `mv ~/.gagipress/config.yaml.backup ~/.gagipress/config.yaml`
2. Revert commits: `git revert HEAD~5..HEAD`
3. Report issue with error details

---

## Related Files Reference

- `internal/config/config.go` - Config struct definitions
- `cmd/init.go` - Init command that saves config
- `cmd/root.go` - Viper initialization
- `cmd/db/migrate.go` - Example command that loads config
- `~/.gagipress/config.yaml` - User's config file

---

## Testing Notes

- Config tests use `t.TempDir()` to avoid affecting user's real config
- Tests verify both field names in YAML and successful unmarshal
- Migration command creates backup before modifying config
- Manual testing required for full init → db workflow
