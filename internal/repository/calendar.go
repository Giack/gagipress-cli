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

// CalendarRepository handles calendar database operations
type CalendarRepository struct {
	config *config.SupabaseConfig
	client *http.Client
}

// NewCalendarRepository creates a new calendar repository
func NewCalendarRepository(cfg *config.SupabaseConfig) *CalendarRepository {
	return &CalendarRepository{
		config: cfg,
		client: &http.Client{},
	}
}

// CreateEntry creates a new calendar entry
func (r *CalendarRepository) CreateEntry(input *models.ContentCalendarInput) (*models.ContentCalendar, error) {
	url := fmt.Sprintf("%s/rest/v1/content_calendar", r.config.URL)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entry: %w", err)
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
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create entry: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var entries []models.ContentCalendar
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no entry returned from API")
	}

	return &entries[0], nil
}

// GetEntries retrieves calendar entries with optional filters
func (r *CalendarRepository) GetEntries(status string, limit int) ([]models.ContentCalendar, error) {
	url := fmt.Sprintf("%s/rest/v1/content_calendar?select=*&order=scheduled_for.asc", r.config.URL)

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
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get entries: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var entries []models.ContentCalendar
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return entries, nil
}

// UpdateEntryStatus updates the status of a calendar entry
func (r *CalendarRepository) UpdateEntryStatus(id string, status string) error {
	url := fmt.Sprintf("%s/rest/v1/content_calendar?id=eq.%s", r.config.URL, id)

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
		return fmt.Errorf("failed to update entry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update entry: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteEntry deletes a calendar entry
func (r *CalendarRepository) DeleteEntry(id string) error {
	url := fmt.Sprintf("%s/rest/v1/content_calendar?id=eq.%s", r.config.URL, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete entry: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
