package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/models"
)

// KDPParser parses Amazon KDP sales reports
type KDPParser struct{}

// NewKDPParser creates a new KDP parser
func NewKDPParser() *KDPParser {
	return &KDPParser{}
}

// ParseCSV parses a KDP CSV file
func (p *KDPParser) ParseCSV(reader io.Reader) ([]models.KDPReportRow, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column names to indices
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Required columns (flexible column names for different KDP report formats)
	titleCol := p.findColumn(colMap, "title", "book title", "product")
	asinCol := p.findColumn(colMap, "asin", "asin/isbn")
	dateCol := p.findColumn(colMap, "date", "order date", "transaction date", "sale date")
	unitsCol := p.findColumn(colMap, "units sold", "units", "quantity")
	royaltyCol := p.findColumn(colMap, "royalty", "net units sold", "earnings")
	pageReadsCol := p.findColumn(colMap, "kenp read", "pages read", "page reads")

	if titleCol == -1 || dateCol == -1 {
		return nil, fmt.Errorf("required columns not found (need at least: title, date)")
	}

	var rows []models.KDPReportRow

	// Read data rows
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Warning: skipping malformed row: %v\n", err)
			continue
		}

		// Parse row
		row := models.KDPReportRow{}

		// Title (required)
		if titleCol >= 0 && titleCol < len(record) {
			row.Title = strings.TrimSpace(record[titleCol])
		}

		// ASIN (optional)
		if asinCol >= 0 && asinCol < len(record) {
			row.ASIN = strings.TrimSpace(record[asinCol])
		}

		// Date (required)
		if dateCol >= 0 && dateCol < len(record) {
			dateStr := strings.TrimSpace(record[dateCol])
			// Try multiple date formats
			formats := []string{
				"2006-01-02",
				"01/02/2006",
				"02/01/2006",
				"2006/01/02",
			}
			var parsed bool
			for _, format := range formats {
				if t, err := time.Parse(format, dateStr); err == nil {
					row.OrderDate = t
					parsed = true
					break
				}
			}
			if !parsed {
				fmt.Printf("Warning: could not parse date '%s', skipping row\n", dateStr)
				continue
			}
		}

		// Units sold (optional)
		if unitsCol >= 0 && unitsCol < len(record) {
			unitsStr := strings.TrimSpace(record[unitsCol])
			if units, err := strconv.Atoi(unitsStr); err == nil {
				row.UnitsSold = units
			}
		}

		// Royalty (optional)
		if royaltyCol >= 0 && royaltyCol < len(record) {
			royaltyStr := strings.TrimSpace(record[royaltyCol])
			// Remove currency symbols
			royaltyStr = strings.ReplaceAll(royaltyStr, "$", "")
			royaltyStr = strings.ReplaceAll(royaltyStr, "€", "")
			royaltyStr = strings.ReplaceAll(royaltyStr, ",", "")
			if royalty, err := strconv.ParseFloat(royaltyStr, 64); err == nil {
				row.Royalty = royalty
			}
		}

		// Page reads (optional)
		if pageReadsCol >= 0 && pageReadsCol < len(record) {
			pageReadsStr := strings.TrimSpace(record[pageReadsCol])
			pageReadsStr = strings.ReplaceAll(pageReadsStr, ",", "")
			if pageReads, err := strconv.Atoi(pageReadsStr); err == nil {
				row.PageReads = pageReads
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// findColumn finds a column index by trying multiple possible names
func (p *KDPParser) findColumn(colMap map[string]int, names ...string) int {
	for _, name := range names {
		if idx, ok := colMap[strings.ToLower(name)]; ok {
			return idx
		}
	}
	return -1
}
