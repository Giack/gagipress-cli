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

func newTestBooksRepo(handler http.HandlerFunc) *BooksRepository {
	server := httptest.NewServer(handler)
	cfg := &config.SupabaseConfig{
		URL:     server.URL,
		AnonKey: "test-key",
	}
	return NewBooksRepository(cfg)
}

func TestGetBookByIDPrefix_ValidPrefix_FindsUniqueBook(t *testing.T) {
	book := models.Book{ID: "abcdef12-3456-7890-abcd-ef1234567890", Title: "Test Book", Genre: "Fiction"}
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify RPC endpoint is called with POST
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rpc/find_book_by_prefix") {
			t.Errorf("expected RPC endpoint, got path: %s", r.URL.Path)
		}
		// Verify request body contains prefix pattern
		var reqBody map[string]string
		json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody["prefix_pattern"] != "abcdef12" {
			t.Errorf("expected prefix_pattern 'abcdef12', got %s", reqBody["prefix_pattern"])
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{book})
	}

	repo := newTestBooksRepo(handler)
	result, err := repo.GetBookByIDPrefix("abcdef12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != book.ID {
		t.Errorf("expected ID %s, got %s", book.ID, result.ID)
	}
	if result.Title != book.Title {
		t.Errorf("expected title %s, got %s", book.Title, result.Title)
	}
}

func TestGetBookByIDPrefix_MinimumLength_FindsBook(t *testing.T) {
	book := models.Book{ID: "abcdef12-3456-7890-abcd-ef1234567890", Title: "Test Book", Genre: "Fiction"}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{book})
	}

	repo := newTestBooksRepo(handler)
	result, err := repo.GetBookByIDPrefix("abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != book.ID {
		t.Errorf("expected ID %s, got %s", book.ID, result.ID)
	}
}

func TestGetBookByIDPrefix_PrefixTooShort_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP request for short prefix")
	}

	repo := newTestBooksRepo(handler)
	_, err := repo.GetBookByIDPrefix("abc")
	if err == nil {
		t.Fatal("expected error for short prefix")
	}
	if !strings.Contains(err.Error(), "prefix too short") {
		t.Errorf("expected 'prefix too short' error, got: %v", err)
	}
}

func TestGetBookByIDPrefix_NoMatches_Error(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{})
	}

	repo := newTestBooksRepo(handler)
	_, err := repo.GetBookByIDPrefix("abcdef12")
	if err == nil {
		t.Fatal("expected error for no matches")
	}
	if !strings.Contains(err.Error(), "no book found") {
		t.Errorf("expected 'no book found' error, got: %v", err)
	}
}

func TestGetBookByIDPrefix_MultipleMatches_DisambiguationError(t *testing.T) {
	books := []models.Book{
		{ID: "abcdef12-1111-1111-1111-111111111111", Title: "Book One", Genre: "Fiction"},
		{ID: "abcdef12-2222-2222-2222-222222222222", Title: "Book Two", Genre: "Fantasy"},
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(books)
	}

	repo := newTestBooksRepo(handler)
	_, err := repo.GetBookByIDPrefix("abcdef12")
	if err == nil {
		t.Fatal("expected error for multiple matches")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "multiple books match") {
		t.Errorf("expected 'multiple books match' error, got: %v", err)
	}
	if !strings.Contains(errMsg, "Book One") || !strings.Contains(errMsg, "Book Two") {
		t.Errorf("expected disambiguation list with book titles, got: %v", err)
	}
}

func TestGetBookByIDPrefix_FullUUID_FindsBook(t *testing.T) {
	fullID := "abcdef12-3456-7890-abcd-ef1234567890"
	book := models.Book{ID: fullID, Title: "Test Book", Genre: "Fiction"}
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.Book{book})
	}

	repo := newTestBooksRepo(handler)
	result, err := repo.GetBookByIDPrefix(fullID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != fullID {
		t.Errorf("expected ID %s, got %s", fullID, result.ID)
	}
}
