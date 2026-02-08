package supabase

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/supabase-community/supabase-go"
)

// Client wraps the Supabase client with our configuration
type Client struct {
	*supabase.Client
	config *config.SupabaseConfig
}

// NewClient creates a new Supabase client from configuration
func NewClient(cfg *config.SupabaseConfig) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("supabase URL is required")
	}
	if cfg.AnonKey == "" {
		return nil, fmt.Errorf("supabase anon key is required")
	}

	// Use service key if available, otherwise anon key
	apiKey := cfg.ServiceKey
	if apiKey == "" {
		apiKey = cfg.AnonKey
	}

	client, err := supabase.NewClient(cfg.URL, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &Client{
		Client: client,
		config: cfg,
	}, nil
}

// TestConnection verifies the connection to Supabase
func (c *Client) TestConnection() error {
	// Simple health check - try to query a system table
	// For now, just verify the client was created successfully
	if c.Client == nil {
		return fmt.Errorf("client is nil")
	}
	return nil
}
