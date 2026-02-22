package stats

import (
	"fmt"
	"os"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/parser"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [kdp-sales.csv]",
	Short: "Import Amazon KDP sales report",
	Long: `Import an Amazon KDP sales CSV report to correlate sales with social media activity.

Required columns in the CSV:
  - Title
  - ASIN
  - Royalty
  - Units Sold
  - Date`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	StatsCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	csvPath := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("📊 Import KDP Sales Data"))
	fmt.Printf("File: %s\n\n", csvPath)

	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse CSV
	spinner := ui.NewSpinner("Parsing KDP report...")
	spinner.Start()
	kdpParser := parser.NewKDPParser()
	rows, err := kdpParser.ParseCSV(file)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(rows) == 0 {
		ui.Warning("No data rows found in the CSV.")
		return nil
	}

	fmt.Printf("✅ Parsed %d rows of sales data.\n\n", len(rows))

	salesRepo := repository.NewSalesRepository(&cfg.Supabase)
	booksRepo := repository.NewBooksRepository(&cfg.Supabase)

	spinner = ui.NewSpinner("Saving sales data to database...")
	spinner.Start()

	successCount := 0

	// Cache ASIN to BookID mapping
	asinMap := make(map[string]string)
	books, err := booksRepo.GetAll()
	if err == nil {
		for _, b := range books {
			if b.KDPASIN != "" {
				asinMap[b.KDPASIN] = b.ID
			}
		}
	}

	for _, row := range rows {
		bookID := asinMap[row.ASIN]
		if bookID == "" {
			// Book not found in our catalog, skip
			continue
		}

		input := &models.BookSaleInput{
			BookID:    bookID,
			SaleDate:  models.Date{Time: row.OrderDate},
			UnitsSold: row.UnitsSold,
			Royalty:   row.Royalty,
			PageReads: row.PageReads,
		}

		_, err = salesRepo.CreateSale(input)
		if err == nil {
			successCount++
		}
	}
	spinner.Stop()

	fmt.Printf("✅ Imported %d new sales records.\n", successCount)
	fmt.Println("\nNext steps:")
	fmt.Println("  • Run correlation analysis: gagipress stats correlate")

	return nil
}
