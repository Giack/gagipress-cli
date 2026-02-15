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

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new book to the catalog",
	Long:  `Interactively add a new book with title, genre, target audience, and other metadata.`,
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📚 Add New Book")
	fmt.Println("═══════════════")

	reader := bufio.NewReader(os.Stdin)
	input := &models.BookInput{}

	// Title (required)
	fmt.Print("Title: ")
	title, _ := reader.ReadString('\n')
	input.Title = strings.TrimSpace(title)

	// Genre (required)
	fmt.Print("Genre (e.g., children, puzzles, savings): ")
	genre, _ := reader.ReadString('\n')
	input.Genre = strings.TrimSpace(genre)

	// Target Audience (optional)
	fmt.Print("Target Audience (optional, press Enter to skip): ")
	audience, _ := reader.ReadString('\n')
	audienceStr := strings.TrimSpace(audience)
	if audienceStr != "" {
		input.TargetAudience = audienceStr
	}

	// KDP ASIN (optional)
	fmt.Print("KDP ASIN (optional, press Enter to skip): ")
	asin, _ := reader.ReadString('\n')
	asinStr := strings.TrimSpace(asin)
	if asinStr != "" {
		input.KDPASIN = asinStr
	}

	// Cover Image URL (optional)
	fmt.Print("Cover Image URL (optional, press Enter to skip): ")
	coverURL, _ := reader.ReadString('\n')
	coverURLStr := strings.TrimSpace(coverURL)
	if coverURLStr != "" {
		input.CoverImageURL = coverURLStr
	}

	// Publication Date (optional)
	fmt.Print("Publication Date (YYYY-MM-DD, optional, press Enter to skip): ")
	pubDate, _ := reader.ReadString('\n')
	pubDateStr := strings.TrimSpace(pubDate)
	if pubDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", pubDateStr)
		if err != nil {
			fmt.Printf("⚠️  Invalid date format, skipping publication date\n")
		} else {
			input.PublicationDate = &models.Date{Time: parsedDate}
		}
	}

	// Validate
	if err := input.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	fmt.Println("\n💾 Saving book...")
	repo := repository.NewBooksRepository(&cfg.Supabase)
	book, err := repo.Create(input)
	if err != nil {
		return fmt.Errorf("failed to save book: %w", err)
	}

	fmt.Println("\n✅ Book added successfully!")
	fmt.Printf("   ID: %s\n", book.ID)
	fmt.Printf("   Title: %s\n", book.Title)
	fmt.Printf("   Genre: %s\n", book.Genre)

	return nil
}
