package books

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
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

	// Build rows
	rows := make([][]string, len(books))
	for i, book := range books {
		rows[i] = []string{
			book.ID, // Full UUID for copy-paste and approval
			book.Title,        // No truncation
			book.Genre,
			book.TargetAudience,
		}
	}

	// Render
	table := ui.RenderTable(ui.TableConfig{
		Headers:  []string{"ID", "Title", "Genre", "Target Audience"},
		Rows:     rows,
		MaxWidth: ui.GetTerminalWidth(),
	})

	fmt.Println(ui.StyleHeader.Render("📚 Book Catalog"))
	fmt.Println(table)

	fmt.Printf("\nTotal books: %d\n", len(books))

	return nil
}
