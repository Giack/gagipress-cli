package models

import (
	"time"
)

// BookSale represents a daily sales record for a book
type BookSale struct {
	ID         string    `json:"id"`
	BookID     string    `json:"book_id"`
	SaleDate   time.Time `json:"sale_date"`
	UnitsSold  int       `json:"units_sold"`
	Royalty    float64   `json:"royalty"`
	PageReads  int       `json:"page_reads"`
	CreatedAt  time.Time `json:"created_at"`
}

// BookSaleInput represents input for creating a book sale record
type BookSaleInput struct {
	BookID    string    `json:"book_id"`
	SaleDate  time.Time `json:"sale_date"`
	UnitsSold int       `json:"units_sold"`
	Royalty   float64   `json:"royalty"`
	PageReads int       `json:"page_reads"`
}

// Validate validates book sale input
func (b *BookSaleInput) Validate() error {
	if b.BookID == "" {
		return ErrInvalidInput{Field: "book_id", Message: "book ID is required"}
	}
	if b.SaleDate.IsZero() {
		return ErrInvalidInput{Field: "sale_date", Message: "sale date is required"}
	}
	if b.UnitsSold < 0 {
		return ErrInvalidInput{Field: "units_sold", Message: "units sold cannot be negative"}
	}
	if b.Royalty < 0 {
		return ErrInvalidInput{Field: "royalty", Message: "royalty cannot be negative"}
	}
	return nil
}

// KDPReportRow represents a row from Amazon KDP sales report CSV
type KDPReportRow struct {
	Title        string
	ASIN         string
	OrderDate    time.Time
	UnitsSold    int
	Royalty      float64
	PageReads    int
	Marketplace  string
}
