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
