package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
)

// SalesRepository handles sales database operations
type SalesRepository struct {
	config *config.SupabaseConfig
	client *http.Client
}

// NewSalesRepository creates a new sales repository
func NewSalesRepository(cfg *config.SupabaseConfig) *SalesRepository {
	return &SalesRepository{
		config: cfg,
		client: &http.Client{},
	}
}

// CreateSale creates a new book sale record
func (r *SalesRepository) CreateSale(input *models.BookSaleInput) (*models.BookSale, error) {
	url := fmt.Sprintf("%s/rest/v1/book_sales", r.config.URL)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sale: %w", err)
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
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create sale: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var sales []models.BookSale
	if err := json.Unmarshal(body, &sales); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(sales) == 0 {
		return nil, fmt.Errorf("no sale returned from API")
	}

	return &sales[0], nil
}

// GetSalesByBook retrieves sales for a specific book
func (r *SalesRepository) GetSalesByBook(bookID string, from, to time.Time) ([]models.BookSale, error) {
	url := fmt.Sprintf("%s/rest/v1/book_sales?book_id=eq.%s&order=sale_date.asc", r.config.URL, bookID)

	if !from.IsZero() {
		url += fmt.Sprintf("&sale_date=gte.%s", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		url += fmt.Sprintf("&sale_date=lte.%s", to.Format("2006-01-02"))
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
		return nil, fmt.Errorf("failed to get sales: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get sales: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var sales []models.BookSale
	if err := json.Unmarshal(body, &sales); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return sales, nil
}

// GetAllSales retrieves all sales
func (r *SalesRepository) GetAllSales(from, to time.Time) ([]models.BookSale, error) {
	url := fmt.Sprintf("%s/rest/v1/book_sales?select=*&order=sale_date.desc", r.config.URL)

	if !from.IsZero() {
		url += fmt.Sprintf("&sale_date=gte.%s", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		url += fmt.Sprintf("&sale_date=lte.%s", to.Format("2006-01-02"))
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
		return nil, fmt.Errorf("failed to get sales: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get sales: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var sales []models.BookSale
	if err := json.Unmarshal(body, &sales); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return sales, nil
}
