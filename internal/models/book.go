package models

import (
	"time"
)

// Book represents a book in the catalog
type Book struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Genre           string     `json:"genre"`
	TargetAudience  string     `json:"target_audience,omitempty"`
	KDPASIN         string     `json:"kdp_asin,omitempty"`
	CoverImageURL   string     `json:"cover_image_url,omitempty"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
	CurrentRank     *int       `json:"current_rank,omitempty"`
	TotalSales      int        `json:"total_sales"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// BookInput represents input for creating/updating a book
type BookInput struct {
	Title           string     `json:"title"`
	Genre           string     `json:"genre"`
	TargetAudience  string     `json:"target_audience,omitempty"`
	KDPASIN         string     `json:"kdp_asin,omitempty"`
	CoverImageURL   string     `json:"cover_image_url,omitempty"`
	PublicationDate *time.Time `json:"publication_date,omitempty"`
}

// Validate validates book input
func (b *BookInput) Validate() error {
	if b.Title == "" {
		return ErrInvalidInput{Field: "title", Message: "title is required"}
	}
	if b.Genre == "" {
		return ErrInvalidInput{Field: "genre", Message: "genre is required"}
	}
	return nil
}

// ErrInvalidInput represents a validation error
type ErrInvalidInput struct {
	Field   string
	Message string
}

func (e ErrInvalidInput) Error() string {
	return e.Message
}
