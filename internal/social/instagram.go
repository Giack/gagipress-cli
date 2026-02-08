package social

import (
	"fmt"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
)

// InstagramClient handles Instagram Graph API interactions
type InstagramClient struct {
	accessToken string
	accountID   string
}

// Post represents an Instagram post
type Post struct {
	ID           string    `json:"id"`
	Caption      string    `json:"caption"`
	MediaType    string    `json:"media_type"` // IMAGE, VIDEO, CAROUSEL_ALBUM
	MediaURL     string    `json:"media_url"`
	Permalink    string    `json:"permalink"`
	Timestamp    time.Time `json:"timestamp"`
	LikesCount   int       `json:"like_count"`
	CommentsCount int      `json:"comments_count"`
}

// Metrics represents post performance metrics
type Metrics struct {
	Impressions int `json:"impressions"`
	Reach       int `json:"reach"`
	Engagement  int `json:"engagement"`
	Saves       int `json:"saves"`
	Shares      int `json:"shares"`
}

// NewInstagramClient creates a new Instagram API client
func NewInstagramClient(cfg *config.InstagramConfig) *InstagramClient {
	return &InstagramClient{
		accessToken: cfg.AccessToken,
		accountID:   cfg.AccountID,
	}
}

// PublishPost publishes a post to Instagram
// TODO: Implement actual API call
func (c *InstagramClient) PublishPost(caption string, mediaURL string) (*Post, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("not implemented: Instagram publishing requires OAuth flow setup")
}

// GetPostMetrics retrieves metrics for a specific post
// TODO: Implement actual API call
func (c *InstagramClient) GetPostMetrics(postID string) (*Metrics, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("not implemented: requires Instagram Graph API access")
}

// GetRecentPosts retrieves recent posts
// TODO: Implement actual API call
func (c *InstagramClient) GetRecentPosts(limit int) ([]Post, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("not implemented: requires Instagram Graph API access")
}

// TestConnection tests the Instagram API connection
func (c *InstagramClient) TestConnection() error {
	if c.accessToken == "" {
		return fmt.Errorf("Instagram access token not configured")
	}
	// TODO: Implement actual connection test with API
	return fmt.Errorf("not implemented: OAuth flow required first")
}

// Note: OAuth flow will be implemented in Week 3
// Requires:
// 1. Facebook App setup
// 2. Instagram Business Account
// 3. OAuth redirect handling
// 4. Token refresh mechanism
