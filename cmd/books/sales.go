package books

import (
	"fmt"
	"os"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/parser"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var salesCmd = &cobra.Command{
	Use:   "sales",
	Short: "Manage book sales data",
	Long:  `Import and manage book sales data from Amazon KDP.`,
}

var importCmd = &cobra.Command{
	Use:   "import [csv-file]",
	Short: "Import sales data from KDP CSV report",
	Long: `Import sales data from an Amazon KDP sales report CSV file.

The importer will:
  - Parse the CSV file
  - Match books by ASIN or title
  - Create daily sales records
  - Update total sales counts

Supports various KDP report formats.`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	salesCmd.AddCommand(importCmd)
	BooksCmd.AddCommand(salesCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	csvFile := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📊 KDP Sales Import")
	fmt.Println("═══════════════════")

	// Open CSV file
	fmt.Printf("Reading file: %s\n", csvFile)
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse CSV
	fmt.Println("⏳ Parsing CSV...")
	kdpParser := parser.NewKDPParser()
	rows, err := kdpParser.ParseCSV(file)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	fmt.Printf("✅ Parsed %d rows\n\n", len(rows))

	// Get books from database
	booksRepo := repository.NewBooksRepository(&cfg.Supabase)
	books, err := booksRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get books: %w", err)
	}

	// Create book lookup maps
	booksByASIN := make(map[string]models.Book)
	booksByTitle := make(map[string]models.Book)
	for _, book := range books {
		if book.KDPASIN != "" {
			booksByASIN[book.KDPASIN] = book
		}
		booksByTitle[book.Title] = book
	}

	// Process rows and create sales records
	salesRepo := repository.NewSalesRepository(&cfg.Supabase)
	imported := 0
	skipped := 0

	fmt.Println("💾 Importing sales data...")

	for _, row := range rows {
		// Find matching book
		var book *models.Book

		// Try ASIN match first
		if row.ASIN != "" {
			if b, ok := booksByASIN[row.ASIN]; ok {
				book = &b
			}
		}

		// Fallback to title match
		if book == nil {
			if b, ok := booksByTitle[row.Title]; ok {
				book = &b
			}
		}

		if book == nil {
			fmt.Printf("⚠️  Skipping: no matching book for '%s' (ASIN: %s)\n", row.Title, row.ASIN)
			skipped++
			continue
		}

		// Create sale record
		saleInput := &models.BookSaleInput{
			BookID:    book.ID,
			SaleDate:  row.OrderDate,
			UnitsSold: row.UnitsSold,
			Royalty:   row.Royalty,
			PageReads: row.PageReads,
		}

		if err := saleInput.Validate(); err != nil {
			fmt.Printf("⚠️  Skipping invalid sale: %v\n", err)
			skipped++
			continue
		}

		_, err := salesRepo.CreateSale(saleInput)
		if err != nil {
			// Likely duplicate - skip silently
			skipped++
			continue
		}

		imported++
	}

	fmt.Println()
	fmt.Println("═══════════════════")
	fmt.Printf("✅ Import Complete!\n")
	fmt.Printf("   Imported: %d sales\n", imported)
	fmt.Printf("   Skipped:  %d rows\n\n", skipped)

	if imported > 0 {
		fmt.Println("Next steps:")
		fmt.Println("  • View sales: gagipress stats show")
		fmt.Println("  • Analyze correlation: gagipress stats correlate")
	}

	return nil
}
