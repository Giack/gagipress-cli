package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
)

// TestGetMetrics_TimestampHasNoPlus verifies that timestamps sent in the URL
// query string use UTC format (ending in "Z") and never contain a raw "+"
// character, which Postgres decodes as a space, causing a 400 parse error.
func TestGetMetrics_TimestampHasNoPlus(t *testing.T) {
	var capturedRawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.PostMetric{})
	}))
	defer server.Close()

	repo := NewMetricsRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	// Simulate a +01:00 timezone (Italy) — this is the common production case.
	loc := time.FixedZone("CET", 60*60)
	from := time.Date(2026, 1, 23, 10, 26, 43, 0, loc)
	to := time.Date(2026, 1, 30, 10, 26, 43, 0, loc)

	_, err := repo.GetMetrics("", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(capturedRawQuery, "+") {
		t.Errorf("URL query contains '+' which Postgres decodes as space — use UTC instead: %q", capturedRawQuery)
	}
}

// TestGetMetrics_TimestampIsUTC verifies that from/to timestamps are sent
// as UTC (suffix "Z") regardless of the caller's local timezone.
func TestGetMetrics_TimestampIsUTC(t *testing.T) {
	var capturedRawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.PostMetric{})
	}))
	defer server.Close()

	repo := NewMetricsRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	loc := time.FixedZone("CET", 60*60)
	from := time.Date(2026, 1, 23, 10, 26, 43, 0, loc) // 10:26:43+01:00 == 09:26:43Z
	to := time.Date(2026, 1, 30, 10, 26, 43, 0, loc)   // same

	_, err := repo.GetMetrics("", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both timestamps must appear in the URL in UTC form (ending with 'Z')
	if !strings.Contains(capturedRawQuery, "09%3A26%3A43Z") && !strings.Contains(capturedRawQuery, "09:26:43Z") {
		t.Errorf("expected UTC timestamp '09:26:43Z' in URL (10:26:43+01:00 converted to UTC), got: %q", capturedRawQuery)
	}
}
