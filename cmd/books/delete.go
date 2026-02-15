package books

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [book-id]",
	Short: "Delete a book from the catalog",
	Long:  `Delete a book and all its associated content. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	bookID := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repo := repository.NewBooksRepository(&cfg.Supabase)

	// Resolve ID prefix to full book
	book, err := repo.GetBookByIDPrefix(bookID)
	if err != nil {
		return fmt.Errorf("failed to get book: %w", err)
	}
	bookID = book.ID

	fmt.Println("📚 Delete Book")
	fmt.Println("══════════════")
	fmt.Printf("Book: %s\n", book.Title)
	fmt.Printf("Genre: %s\n", book.Genre)
	fmt.Println("\n⚠️  WARNING: This will delete the book and all associated content!")
	fmt.Print("Type 'DELETE' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation)

	if confirmation != "DELETE" {
		fmt.Println("\n❌ Deletion cancelled")
		return nil
	}

	// Delete book
	fmt.Print("\n🗑️  Deleting book... ")
	if err := repo.Delete(bookID); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to delete book: %w", err)
	}

	fmt.Println("✅ OK")
	fmt.Println("\n✅ Book deleted successfully!")

	return nil
}
