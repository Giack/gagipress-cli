package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Supabase  SupabaseConfig  `mapstructure:"supabase"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	Instagram InstagramConfig `mapstructure:"instagram"`
	TikTok    TikTokConfig    `mapstructure:"tiktok"`
	Amazon    AmazonConfig    `mapstructure:"amazon"`
}

// SupabaseConfig holds Supabase connection details
type SupabaseConfig struct {
	URL        string `mapstructure:"url"`
	AnonKey    string `mapstructure:"anon_key"`
	ServiceKey string `mapstructure:"service_key"`
}

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

// InstagramConfig holds Instagram API configuration
type InstagramConfig struct {
	AccessToken string `mapstructure:"access_token"`
	AccountID   string `mapstructure:"account_id"`
}

// TikTokConfig holds TikTok API configuration
type TikTokConfig struct {
	AccessToken string `mapstructure:"access_token"`
	AccountID   string `mapstructure:"account_id"`
}

// AmazonConfig holds Amazon KDP credentials
type AmazonConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

// Load loads configuration from file
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}
	return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".gagipress")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("unable to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	viper.Set("supabase", cfg.Supabase)
	viper.Set("openai", cfg.OpenAI)
	viper.Set("instagram", cfg.Instagram)
	viper.Set("tiktok", cfg.TikTok)
	viper.Set("amazon", cfg.Amazon)

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("unable to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Supabase.URL == "" {
		return fmt.Errorf("supabase URL is required")
	}
	if c.Supabase.AnonKey == "" {
		return fmt.Errorf("supabase anon key is required")
	}
	return nil
}

// IsConfigured checks if the application is configured
func IsConfigured() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	configFile := filepath.Join(home, ".gagipress", "config.yaml")
	_, err = os.Stat(configFile)
	return err == nil
}
