package books

import (
	"github.com/spf13/cobra"
)

// BooksCmd represents the books command group
var BooksCmd = &cobra.Command{
	Use:   "books",
	Short: "Manage your book catalog",
	Long: `Manage your Amazon KDP book catalog:
  - Add new books with metadata
  - List all books
  - Edit book information
  - Delete books from catalog`,
}

func init() {
	BooksCmd.AddCommand(addCmd)
	BooksCmd.AddCommand(listCmd)
	BooksCmd.AddCommand(editCmd)
	BooksCmd.AddCommand(deleteCmd)
}
