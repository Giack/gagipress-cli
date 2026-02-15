package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
)

// ContentRepository handles content database operations
type ContentRepository struct {
	config *config.SupabaseConfig
	client *http.Client
}

// NewContentRepository creates a new content repository
func NewContentRepository(cfg *config.SupabaseConfig) *ContentRepository {
	return &ContentRepository{
		config: cfg,
		client: &http.Client{},
	}
}

// CreateIdea creates a new content idea
func (r *ContentRepository) CreateIdea(input *models.ContentIdeaInput) (*models.ContentIdea, error) {
	url := fmt.Sprintf("%s/rest/v1/content_ideas", r.config.URL)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal idea: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create idea: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create idea: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var ideas []models.ContentIdea
	if err := json.Unmarshal(body, &ideas); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(ideas) == 0 {
		return nil, fmt.Errorf("no idea returned from API")
	}

	return &ideas[0], nil
}

// GetIdeas retrieves content ideas with optional filters
func (r *ContentRepository) GetIdeas(status string, limit int) ([]models.ContentIdea, error) {
	url := fmt.Sprintf("%s/rest/v1/content_ideas?select=*&order=generated_at.desc", r.config.URL)

	if status != "" {
		url += fmt.Sprintf("&status=eq.%s", status)
	}
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get ideas: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get ideas: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var ideas []models.ContentIdea
	if err := json.Unmarshal(body, &ideas); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ideas, nil
}

// UpdateIdeaStatus updates the status of a content idea
func (r *ContentRepository) UpdateIdeaStatus(id string, status string) error {
	url := fmt.Sprintf("%s/rest/v1/content_ideas?id=eq.%s", r.config.URL, id)

	data := map[string]string{"status": status}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update idea: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update idea: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetIdeaByIDPrefix finds a content idea by UUID prefix (minimum 6 characters).
// Returns an error if the prefix is ambiguous (matches multiple ideas) or not found.
// Uses the find_idea_by_prefix PostgreSQL function via PostgREST RPC.
func (r *ContentRepository) GetIdeaByIDPrefix(prefix string) (*models.ContentIdea, error) {
	if len(prefix) < 6 {
		return nil, fmt.Errorf("prefix too short: must be at least 6 characters, got %d", len(prefix))
	}

	// Use PostgreSQL RPC function for UUID prefix matching
	requestURL := fmt.Sprintf("%s/rest/v1/rpc/find_idea_by_prefix", r.config.URL)

	// Create request body with prefix parameter
	reqBody := map[string]string{"prefix_pattern": prefix}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get idea by prefix: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get idea by prefix: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var ideas []models.ContentIdea
	if err := json.Unmarshal(body, &ideas); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	switch len(ideas) {
	case 0:
		return nil, fmt.Errorf("no idea found with ID prefix %q", prefix)
	case 1:
		return &ideas[0], nil
	default:
		ids := make([]string, len(ideas))
		for i, idea := range ideas {
			ids[i] = idea.ID
		}
		return nil, fmt.Errorf("ambiguous prefix %q matches %d ideas: %s", prefix, len(ideas), strings.Join(ids, ", "))
	}
}

// CreateScript creates a new content script
func (r *ContentRepository) CreateScript(input *models.ContentScriptInput) (*models.ContentScript, error) {
	url := fmt.Sprintf("%s/rest/v1/content_scripts", r.config.URL)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal script: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create script: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create script: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var scripts []models.ContentScript
	if err := json.Unmarshal(body, &scripts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(scripts) == 0 {
		return nil, fmt.Errorf("no script returned from API")
	}

	return &scripts[0], nil
}

// GetScripts retrieves content scripts
func (r *ContentRepository) GetScripts(limit int) ([]models.ContentScript, error) {
	url := fmt.Sprintf("%s/rest/v1/content_scripts?select=*&order=created_at.desc", r.config.URL)

	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get scripts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get scripts: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var scripts []models.ContentScript
	if err := json.Unmarshal(body, &scripts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return scripts, nil
}
