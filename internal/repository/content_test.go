package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
)

func TestGetIdeaByIDPrefix_PrefixTooShort(t *testing.T) {
	repo := NewContentRepository(&config.SupabaseConfig{URL: "http://localhost", AnonKey: "test"})

	tests := []struct {
		name   string
		prefix string
	}{
		{"empty", ""},
		{"1 char", "a"},
		{"5 chars", "abcde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.GetIdeaByIDPrefix(tt.prefix)
			if err == nil {
				t.Fatal("expected error for short prefix, got nil")
			}
			if got := err.Error(); !contains(got, "prefix too short") {
				t.Errorf("expected 'prefix too short' error, got: %s", got)
			}
		})
	}
}

func TestGetIdeaByIDPrefix_NoMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	_, err := repo.GetIdeaByIDPrefix("abcdef")
	if err == nil {
		t.Fatal("expected error for no matches, got nil")
	}
	if got := err.Error(); !contains(got, "no idea found") {
		t.Errorf("expected 'no idea found' error, got: %s", got)
	}
}

func TestGetIdeaByIDPrefix_SingleMatch(t *testing.T) {
	fullID := "abcdef12-3456-7890-abcd-ef1234567890"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the query uses like filter
		query := r.URL.Query().Get("id")
		if query == "" {
			t.Error("expected id query parameter")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{
			{ID: fullID, Type: "educational", BriefDescription: "Test idea", Status: "pending"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	idea, err := repo.GetIdeaByIDPrefix("abcdef12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idea.ID != fullID {
		t.Errorf("expected ID %s, got %s", fullID, idea.ID)
	}
}

func TestGetIdeaByIDPrefix_6CharPrefix(t *testing.T) {
	fullID := "abcdef12-3456-7890-abcd-ef1234567890"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{
			{ID: fullID, Type: "educational", BriefDescription: "Test idea", Status: "pending"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	idea, err := repo.GetIdeaByIDPrefix("abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idea.ID != fullID {
		t.Errorf("expected ID %s, got %s", fullID, idea.ID)
	}
}

func TestGetIdeaByIDPrefix_MultipleMatches(t *testing.T) {
	id1 := "abcdef12-1111-1111-1111-111111111111"
	id2 := "abcdef12-2222-2222-2222-222222222222"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{
			{ID: id1, Type: "educational", BriefDescription: "Idea 1", Status: "pending"},
			{ID: id2, Type: "entertainment", BriefDescription: "Idea 2", Status: "pending"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	_, err := repo.GetIdeaByIDPrefix("abcdef12")
	if err == nil {
		t.Fatal("expected error for multiple matches, got nil")
	}
	errMsg := err.Error()
	if !contains(errMsg, "ambiguous") {
		t.Errorf("expected 'ambiguous' in error, got: %s", errMsg)
	}
	if !contains(errMsg, id1) || !contains(errMsg, id2) {
		t.Errorf("expected both IDs in error message, got: %s", errMsg)
	}
}

func TestGetIdeaByIDPrefix_FullUUID(t *testing.T) {
	fullID := "abcdef12-3456-7890-abcd-ef1234567890"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{
			{ID: fullID, Type: "educational", BriefDescription: "Test idea", Status: "pending"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{URL: server.URL, AnonKey: "test"})

	idea, err := repo.GetIdeaByIDPrefix(fullID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idea.ID != fullID {
		t.Errorf("expected ID %s, got %s", fullID, idea.ID)
	}
}

func TestGetIdeaByIDPrefix_UsesServiceKey(t *testing.T) {
	var gotAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("apikey")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{
			{ID: "abcdef12-3456-7890-abcd-ef1234567890", Status: "pending"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{
		URL:        server.URL,
		AnonKey:    "anon-key",
		ServiceKey: "service-key",
	})

	_, err := repo.GetIdeaByIDPrefix("abcdef12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAPIKey != "service-key" {
		t.Errorf("expected service-key to be used, got: %s", gotAPIKey)
	}
}

func TestGetIdeaByIDPrefix_FallsBackToAnonKey(t *testing.T) {
	var gotAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("apikey")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.ContentIdea{
			{ID: "abcdef12-3456-7890-abcd-ef1234567890", Status: "pending"},
		})
	}))
	defer server.Close()

	repo := NewContentRepository(&config.SupabaseConfig{
		URL:     server.URL,
		AnonKey: "anon-key",
	})

	_, err := repo.GetIdeaByIDPrefix("abcdef12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAPIKey != "anon-key" {
		t.Errorf("expected anon-key to be used, got: %s", gotAPIKey)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
