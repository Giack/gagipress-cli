package books

import (
	"fmt"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all books in the catalog",
	Long:  `Display a table of all books with their metadata.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📚 Book Catalog")
	fmt.Println("═══════════════\n")

	// Get all books
	repo := repository.NewBooksRepository(&cfg.Supabase)
	books, err := repo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get books: %w", err)
	}

	if len(books) == 0 {
		fmt.Println("No books in catalog. Add one with 'gagipress books add'")
		return nil
	}

	// Print table header
	fmt.Printf("%-8s %-40s %-20s %-15s %-12s\n", "ID", "Title", "Genre", "ASIN", "Sales")
	fmt.Println(strings.Repeat("─", 100))

	// Print books
	for _, book := range books {
		id := book.ID
		if len(id) > 8 {
			id = id[:8]
		}

		title := book.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		genre := book.Genre
		if len(genre) > 20 {
			genre = genre[:17] + "..."
		}

		asin := book.KDPASIN
		if asin == "" {
			asin = "N/A"
		}
		if len(asin) > 15 {
			asin = asin[:12] + "..."
		}

		fmt.Printf("%-8s %-40s %-20s %-15s %-12d\n",
			id, title, genre, asin, book.TotalSales)
	}

	fmt.Printf("\nTotal books: %d\n", len(books))

	return nil
}
