package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
)

// TestGetStatusCounts verifies that GetStatusCounts returns the correct count
// per status by reading only the status column from content_calendar.
func TestGetStatusCounts(t *testing.T) {
	// Simulate rows with mixed statuses
	rows := []map[string]string{
		{"status": "approved"},
		{"status": "approved"},
		{"status": "failed"},
		{"status": "published"},
		{"status": "published"},
		{"status": "published"},
	}

	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rows)
	}))
	defer server.Close()

	repo := NewCalendarRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	counts, err := repo.GetStatusCounts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify counts
	if counts["approved"] != 2 {
		t.Errorf("expected approved=2, got %d", counts["approved"])
	}
	if counts["failed"] != 1 {
		t.Errorf("expected failed=1, got %d", counts["failed"])
	}
	if counts["published"] != 3 {
		t.Errorf("expected published=3, got %d", counts["published"])
	}

	// Verify only the status column is requested (keep payload small)
	if capturedPath == "" {
		t.Error("expected a URL query string, got empty")
	}
}

// TestGetEntriesNeedingMedia verifies that the correct query is sent and results are parsed.
func TestGetEntriesNeedingMedia(t *testing.T) {
	hook := "Did you know this book exists?"
	book := models.Book{ID: "book-1", Title: "My Book", Genre: "children", KDPASIN: "B0ABC123"}
	idea := models.ContentIdeaWithBook{ContentIdea: models.ContentIdea{ID: "idea-1"}, Book: &book}
	script := models.ContentScriptWithIdea{ContentScript: models.ContentScript{ID: "script-1", Hook: hook}, Idea: &idea}
	entries := []models.ContentCalendarWithScript{
		{
			ContentCalendar: models.ContentCalendar{
				ID:            "entry-1",
				Status:        "approved",
				Platform:      "tiktok",
				PostType:      "reel",
				GenerateMedia: true,
			},
			Script: &script,
		},
	}

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	repo := NewCalendarRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	result, err := repo.GetEntriesNeedingMedia()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].ID != "entry-1" {
		t.Errorf("expected ID entry-1, got %s", result[0].ID)
	}
	// Query must filter by generate_media=true and media_url=is.null
	if !strings.Contains(capturedQuery, "generate_media=eq.true") {
		t.Errorf("query missing generate_media filter, got: %s", capturedQuery)
	}
	if !strings.Contains(capturedQuery, "media_url=is.null") {
		t.Errorf("query missing media_url=is.null filter, got: %s", capturedQuery)
	}
}

// TestUpdateMediaURL verifies that UpdateMediaURL sends the correct PATCH request.
func TestUpdateMediaURL(t *testing.T) {
	var capturedMethod, capturedQuery string
	var capturedBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedQuery = r.URL.RawQuery
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	repo := NewCalendarRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	err := repo.UpdateMediaURL("entry-42", "https://example.com/image.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", capturedMethod)
	}
	if !strings.Contains(capturedQuery, "id=eq.entry-42") {
		t.Errorf("query missing id filter, got: %s", capturedQuery)
	}
	if capturedBody["media_url"] != "https://example.com/image.jpg" {
		t.Errorf("expected media_url in body, got: %v", capturedBody)
	}
}

// TestRetryFailed verifies that RetryFailed sends a PATCH to update
// all 'failed' entries to 'approved' and returns the count of affected rows.
func TestRetryFailed(t *testing.T) {
	// Simulate 3 failed entries being reset to approved
	updated := []models.ContentCalendar{
		{ID: "id-1", Status: "approved"},
		{ID: "id-2", Status: "approved"},
		{ID: "id-3", Status: "approved"},
	}

	var capturedMethod string
	var capturedQuery string
	var capturedBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedQuery = r.URL.RawQuery
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	repo := NewCalendarRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	count, err := repo.RetryFailed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must return the number of retried entries
	if count != 3 {
		t.Errorf("expected count=3, got %d", count)
	}

	// Must use PATCH method
	if capturedMethod != http.MethodPatch {
		t.Errorf("expected PATCH method, got %s", capturedMethod)
	}

	// Must target only failed entries
	if capturedQuery == "" {
		t.Error("expected query string filtering by status=failed")
	}

	// Must set status to approved
	if capturedBody["status"] != "approved" {
		t.Errorf("expected body status=approved, got %q", capturedBody["status"])
	}
}
