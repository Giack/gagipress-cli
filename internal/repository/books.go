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

// BooksRepository handles book database operations
type BooksRepository struct {
	config *config.SupabaseConfig
	client *http.Client
}

// NewBooksRepository creates a new books repository
func NewBooksRepository(cfg *config.SupabaseConfig) *BooksRepository {
	return &BooksRepository{
		config: cfg,
		client: &http.Client{},
	}
}

// Create creates a new book
func (r *BooksRepository) Create(input *models.BookInput) (*models.Book, error) {
	url := fmt.Sprintf("%s/rest/v1/books", r.config.URL)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal book: %w", err)
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
		return nil, fmt.Errorf("failed to create book: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create book: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("no book returned from API")
	}

	return &books[0], nil
}

// GetAll retrieves all books
func (r *BooksRepository) GetAll() ([]models.Book, error) {
	url := fmt.Sprintf("%s/rest/v1/books?select=*&order=created_at.desc", r.config.URL)

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
		return nil, fmt.Errorf("failed to get books: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get books: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return books, nil
}

// GetByID retrieves a book by ID
func (r *BooksRepository) GetByID(id string) (*models.Book, error) {
	url := fmt.Sprintf("%s/rest/v1/books?id=eq.%s&select=*", r.config.URL, id)

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
		return nil, fmt.Errorf("failed to get book: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get book: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("book not found")
	}

	return &books[0], nil
}

// Update updates a book
func (r *BooksRepository) Update(id string, input *models.BookInput) (*models.Book, error) {
	url := fmt.Sprintf("%s/rest/v1/books?id=eq.%s", r.config.URL, id)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal book: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonData)))
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
		return nil, fmt.Errorf("failed to update book: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update book: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(books) == 0 {
		return nil, fmt.Errorf("no book returned from API")
	}

	return &books[0], nil
}

// GetBookByIDPrefix retrieves a book by ID prefix (minimum 6 characters).
// Returns the book if exactly one match is found. Returns an error with
// disambiguation list if multiple books match.
func (r *BooksRepository) GetBookByIDPrefix(prefix string) (*models.Book, error) {
	if len(prefix) < 6 {
		return nil, fmt.Errorf("ID prefix too short: must be at least 6 characters (got %d)", len(prefix))
	}

	url := fmt.Sprintf("%s/rest/v1/books?id=like.%s*&select=*", r.config.URL, prefix)

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
		return nil, fmt.Errorf("failed to get book: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get book: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var books []models.Book
	if err := json.Unmarshal(body, &books); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	switch len(books) {
	case 0:
		return nil, fmt.Errorf("no book found with ID prefix %q", prefix)
	case 1:
		return &books[0], nil
	default:
		ids := make([]string, len(books))
		for i, b := range books {
			ids[i] = fmt.Sprintf("  %s (%s)", b.ID, b.Title)
		}
		return nil, fmt.Errorf("multiple books match prefix %q:\n%s\nPlease use a longer prefix to disambiguate", prefix, strings.Join(ids, "\n"))
	}
}

// Delete deletes a book
func (r *BooksRepository) Delete(id string) error {
	url := fmt.Sprintf("%s/rest/v1/books?id=eq.%s", r.config.URL, id)

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
		return fmt.Errorf("failed to delete book: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete book: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
