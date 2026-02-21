package social

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gagipress/gagipress-cli/internal/errors"
)

const (
	BlotatoBaseURL = "https://backend.blotato.com/v2"
)

// BlotatoClient is the client for interacting with the Blotato API
type BlotatoClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewBlotatoClient creates a new Blotato API client
func NewBlotatoClient(apiKey string) *BlotatoClient {
	return &BlotatoClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// AccountResponse represents an account item from Blotato
type AccountItem struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	FullName string `json:"fullname"`
	Username string `json:"username"`
}

type accountsListResponse struct {
	Items []AccountItem `json:"items"`
}

// GetAccountID fetches the user's connected accounts and returns the account ID for the requested platform.
// For Facebook/LinkedIn, this is the main accountId (subaccounts handling might be needed later).
func (c *BlotatoClient) GetAccountID(platform string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("blotato API key is not configured")
	}

	req, err := http.NewRequest("GET", BlotatoBaseURL+"/users/me/accounts?platform="+platform, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("blotato-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to connect to Blotato API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("blotato API error: %s", string(body))
	}

	var accList accountsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&accList); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "failed to parse Blotato accounts response")
	}

	if len(accList.Items) == 0 {
		return "", fmt.Errorf("no connected accounts found for platform: %s", platform)
	}

	// Just return the first matched account for the platform
	return accList.Items[0].ID, nil
}

// PublishPostRequest represents the request body for publishing a post
type PublishPostRequest struct {
	Post            PostData `json:"post"`
	ScheduledTime   string   `json:"scheduledTime,omitempty"`
	UseNextFreeSlot *bool    `json:"useNextFreeSlot,omitempty"`
}

type PostData struct {
	AccountID string      `json:"accountId"`
	Content   PostContent `json:"content"`
	Target    PostTarget  `json:"target"`
}

type PostContent struct {
	Text      string   `json:"text"`
	MediaURLs []string `json:"mediaUrls"`
	Platform  string   `json:"platform"`
}

type PostTarget struct {
	TargetType string `json:"targetType"`
	// Additional fields like PrivacyLevel could be added here later for TikTok, etc.
}

type PublishResponse struct {
	PostSubmissionID string `json:"postSubmissionId"`
}

// PublishPost creates or schedules a post on Blotato
func (c *BlotatoClient) PublishPost(accountId, platform, text string, mediaUrls []string, scheduledTime *time.Time) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("blotato API key is not configured")
	}

	if mediaUrls == nil {
		mediaUrls = []string{} // ensure it's an empty array and not null
	}

	reqBody := PublishPostRequest{
		Post: PostData{
			AccountID: accountId,
			Content: PostContent{
				Text:      text,
				MediaURLs: mediaUrls,
				Platform:  platform,
			},
			Target: PostTarget{
				TargetType: platform,
			},
		},
	}

	if scheduledTime != nil && !scheduledTime.IsZero() {
		// Use ISO 8601 format like "2025-12-25T15:00:00Z"
		reqBody.ScheduledTime = scheduledTime.UTC().Format(time.RFC3339)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", BlotatoBaseURL+"/posts", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("blotato-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeAPI, "failed to connect to Blotato API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("blotato publish error (%d): %s", resp.StatusCode, string(body))
	}

	var pubResp PublishResponse
	if err := json.NewDecoder(resp.Body).Decode(&pubResp); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, "failed to parse Blotato publish response")
	}

	return pubResp.PostSubmissionID, nil
}
