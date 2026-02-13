package parser

import (
	"strings"
	"testing"
	"time"
)

func TestKDPParser_ParseCSV(t *testing.T) {
	tests := []struct {
		name      string
		csvData   string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid CSV with standard columns",
			csvData: `Title,ASIN,Date,Units Sold,Royalty
Test Book,B0ABC123,2024-01-15,5,$3.50`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple rows",
			csvData: `Title,ASIN,Date,Units Sold,Royalty
Book 1,B0ABC123,2024-01-15,5,$3.50
Book 2,B0DEF456,2024-01-16,3,$2.10`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "alternative column names",
			csvData: `Book Title,ASIN/ISBN,Order Date,Quantity,Earnings
Test Book,B0ABC123,01/15/2024,5,3.50`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "empty CSV",
			csvData:   "",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "missing required columns",
			csvData: `Some Column,Another Column
value1,value2`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "invalid date format skips row",
			csvData: `Title,Date
Test Book,invalid-date
Good Book,2024-01-15`,
			wantCount: 1, // Should skip invalid row
			wantErr:   false,
		},
	}

	parser := NewKDPParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			rows, err := parser.ParseCSV(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(rows) != tt.wantCount {
				t.Errorf("ParseCSV() got %d rows, want %d", len(rows), tt.wantCount)
			}
		})
	}
}

func TestKDPParser_DateFormats(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectedDay int
	}{
		{"ISO format", "2024-01-15", 15},
		{"US format", "01/15/2024", 15},
		{"EU format", "15/01/2024", 15},
		{"slash format", "2024/01/15", 15},
	}

	parser := NewKDPParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvData := `Title,Date
Test Book,` + tt.dateStr

			reader := strings.NewReader(csvData)
			rows, err := parser.ParseCSV(reader)

			if err != nil {
				t.Fatalf("ParseCSV() unexpected error: %v", err)
			}

			if len(rows) != 1 {
				t.Fatalf("Expected 1 row, got %d", len(rows))
			}

			if rows[0].OrderDate.Day() != tt.expectedDay {
				t.Errorf("Expected day %d, got %d", tt.expectedDay, rows[0].OrderDate.Day())
			}
		})
	}
}

func TestKDPParser_RoyaltyParsing(t *testing.T) {
	tests := []struct {
		name           string
		royaltyStr     string
		expectedRoyalty float64
	}{
		{"with dollar sign", "$3.50", 3.50},
		{"with euro sign", "€3.50", 3.50},
		{"no currency", "3.50", 3.50},
		{"with comma thousands", `"$1,234.56"`, 1234.56},
		{"zero", "$0.00", 0.00},
	}

	parser := NewKDPParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvData := `Title,Date,Royalty
Test Book,2024-01-15,` + tt.royaltyStr

			reader := strings.NewReader(csvData)
			rows, err := parser.ParseCSV(reader)

			if err != nil {
				t.Fatalf("ParseCSV() unexpected error: %v", err)
			}

			if len(rows) != 1 {
				t.Fatalf("Expected 1 row, got %d", len(rows))
			}

			if rows[0].Royalty != tt.expectedRoyalty {
				t.Errorf("Expected royalty %v, got %v", tt.expectedRoyalty, rows[0].Royalty)
			}
		})
	}
}

func TestKDPParser_ParseCSV_CompleteRow(t *testing.T) {
	csvData := `Title,ASIN,Date,Units Sold,Royalty,KENP Read
My Amazing Book,B0TEST123,2024-01-15,10,$25.50,1500`

	parser := NewKDPParser()
	reader := strings.NewReader(csvData)
	rows, err := parser.ParseCSV(reader)

	if err != nil {
		t.Fatalf("ParseCSV() unexpected error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	row := rows[0]

	if row.Title != "My Amazing Book" {
		t.Errorf("Expected title 'My Amazing Book', got '%s'", row.Title)
	}

	if row.ASIN != "B0TEST123" {
		t.Errorf("Expected ASIN 'B0TEST123', got '%s'", row.ASIN)
	}

	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !row.OrderDate.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, row.OrderDate)
	}

	if row.UnitsSold != 10 {
		t.Errorf("Expected units sold 10, got %d", row.UnitsSold)
	}

	if row.Royalty != 25.50 {
		t.Errorf("Expected royalty 25.50, got %v", row.Royalty)
	}

	if row.PageReads != 1500 {
		t.Errorf("Expected page reads 1500, got %d", row.PageReads)
	}
}
