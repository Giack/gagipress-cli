package books

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [book-id]",
	Short: "Edit a book in the catalog",
	Long:  `Edit an existing book's metadata. Provide the book ID as argument.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	bookID := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repo := repository.NewBooksRepository(&cfg.Supabase)

	// Get existing book
	fmt.Println("📚 Edit Book")
	fmt.Println("════════════")
	fmt.Print("Loading book... ")

	book, err := repo.GetByID(bookID)
	if err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to get book: %w", err)
	}
	fmt.Println("✅ OK")

	fmt.Printf("Current book: %s\n", book.Title)
	fmt.Println("Press Enter to keep current value, or enter new value:")

	reader := bufio.NewReader(os.Stdin)
	input := &models.BookInput{
		Title:           book.Title,
		Genre:           book.Genre,
		TargetAudience:  book.TargetAudience,
		KDPASIN:         book.KDPASIN,
		CoverImageURL:   book.CoverImageURL,
		PublicationDate: book.PublicationDate,
	}

	// Title
	fmt.Printf("Title [%s]: ", book.Title)
	title, _ := reader.ReadString('\n')
	titleStr := strings.TrimSpace(title)
	if titleStr != "" {
		input.Title = titleStr
	}

	// Genre
	fmt.Printf("Genre [%s]: ", book.Genre)
	genre, _ := reader.ReadString('\n')
	genreStr := strings.TrimSpace(genre)
	if genreStr != "" {
		input.Genre = genreStr
	}

	// Target Audience
	currentAudience := book.TargetAudience
	if currentAudience == "" {
		currentAudience = "N/A"
	}
	fmt.Printf("Target Audience [%s]: ", currentAudience)
	audience, _ := reader.ReadString('\n')
	audienceStr := strings.TrimSpace(audience)
	if audienceStr != "" {
		input.TargetAudience = audienceStr
	}

	// KDP ASIN
	currentASIN := book.KDPASIN
	if currentASIN == "" {
		currentASIN = "N/A"
	}
	fmt.Printf("KDP ASIN [%s]: ", currentASIN)
	asin, _ := reader.ReadString('\n')
	asinStr := strings.TrimSpace(asin)
	if asinStr != "" {
		input.KDPASIN = asinStr
	}

	// Cover Image URL
	currentCover := book.CoverImageURL
	if currentCover == "" {
		currentCover = "N/A"
	}
	fmt.Printf("Cover Image URL [%s]: ", currentCover)
	coverURL, _ := reader.ReadString('\n')
	coverURLStr := strings.TrimSpace(coverURL)
	if coverURLStr != "" {
		input.CoverImageURL = coverURLStr
	}

	// Publication Date
	currentPubDate := "N/A"
	if book.PublicationDate != nil {
		currentPubDate = book.PublicationDate.Format("2006-01-02")
	}
	fmt.Printf("Publication Date (YYYY-MM-DD) [%s]: ", currentPubDate)
	pubDate, _ := reader.ReadString('\n')
	pubDateStr := strings.TrimSpace(pubDate)
	if pubDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", pubDateStr)
		if err != nil {
			fmt.Printf("⚠️  Invalid date format, keeping current value\n")
		} else {
			input.PublicationDate = &parsedDate
		}
	}

	// Validate
	if err := input.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	fmt.Println("\n💾 Updating book...")
	updatedBook, err := repo.Update(bookID, input)
	if err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}

	fmt.Println("\n✅ Book updated successfully!")
	fmt.Printf("   ID: %s\n", updatedBook.ID)
	fmt.Printf("   Title: %s\n", updatedBook.Title)
	fmt.Printf("   Genre: %s\n", updatedBook.Genre)

	return nil
}
