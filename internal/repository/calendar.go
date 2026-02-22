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

// GetEntryByID gets a specific calendar entry by its ID
func (r *CalendarRepository) GetEntryByID(id string) (*models.ContentCalendar, error) {
	apiKey := r.config.ServiceKey
	if apiKey == "" {
		apiKey = r.config.AnonKey
	}

	url := fmt.Sprintf("%s/rest/v1/content_calendar?id=eq.%s&select=*", r.config.URL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get entry: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var entries []models.ContentCalendar
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("entry not found: %s", id)
	}

	return &entries[0], nil
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

// GetStatusCounts returns a count of calendar entries grouped by status.
func (r *CalendarRepository) GetStatusCounts() (map[string]int, error) {
	url := fmt.Sprintf("%s/rest/v1/content_calendar?select=status", r.config.URL)

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
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get status counts: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rows []map[string]string
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	counts := make(map[string]int)
	for _, row := range rows {
		counts[row["status"]]++
	}
	return counts, nil
}

// RetryFailed resets all calendar entries with status 'failed' back to 'approved'
// so the cron job will pick them up again. Returns the number of entries reset.
func (r *CalendarRepository) RetryFailed() (int, error) {
	url := fmt.Sprintf("%s/rest/v1/content_calendar?status=eq.failed", r.config.URL)

	data := map[string]string{"status": "approved"}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
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
		return 0, fmt.Errorf("failed to retry failed entries: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return 0, fmt.Errorf("failed to retry entries: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var updated []models.ContentCalendar
	if err := json.Unmarshal(body, &updated); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return len(updated), nil
}

// GetEntriesNeedingMedia returns approved/scheduled entries that have generate_media=true
// and no media_url set yet, joined with their script data for prompt building.
func (r *CalendarRepository) GetEntriesNeedingMedia() ([]models.ContentCalendarWithScript, error) {
	url := fmt.Sprintf(
		"%s/rest/v1/content_calendar?status=in.(approved,scheduled)&generate_media=eq.true&media_url=is.null&select=*,content_scripts(*)",
		r.config.URL,
	)

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
		return nil, fmt.Errorf("failed to get entries needing media: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get entries needing media: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var entries []models.ContentCalendarWithScript
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return entries, nil
}

// UpdateMediaURL sets the media_url for a calendar entry.
func (r *CalendarRepository) UpdateMediaURL(entryID, mediaURL string) error {
	url := fmt.Sprintf("%s/rest/v1/content_calendar?id=eq.%s", r.config.URL, entryID)

	data := map[string]string{"media_url": mediaURL}
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

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update media URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update media URL: HTTP %d: %s", resp.StatusCode, string(body))
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
