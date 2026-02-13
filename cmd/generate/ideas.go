package generate

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/generator"
	"github.com/gagipress/gagipress-cli/internal/prompts"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	count      int
	bookID     string
	useGemini  bool
)

var ideasCmd = &cobra.Command{
	Use:   "ideas",
	Short: "Generate content ideas using AI",
	Long: `Generate 20-30 content ideas for TikTok/Instagram Reels.
Uses OpenAI API as primary, with automatic fallback to Gemini.

The generator will:
  - Read books from your catalog
  - Generate ideas based on book genre and niche
  - Categorize ideas (educational, entertainment, BTS, UGC, trend)
  - Calculate relevance scores
  - Save to database for approval`,
	RunE: runGenerateIdeas,
}

func init() {
	ideasCmd.Flags().IntVar(&count, "count", 20, "Number of ideas to generate")
	ideasCmd.Flags().StringVar(&bookID, "book", "", "Book ID (optional, generates for all books if not specified)")
	ideasCmd.Flags().BoolVar(&useGemini, "gemini", false, "Use Gemini instead of OpenAI")
}

func runGenerateIdeas(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("💡 Content Idea Generator")
	fmt.Println("══════════════════════════")

	// Get books
	booksRepo := repository.NewBooksRepository(&cfg.Supabase)

	var books []struct {
		id    string
		title string
		genre string
		audience string
	}

	if bookID != "" {
		// Single book
		book, err := booksRepo.GetByID(bookID)
		if err != nil {
			return fmt.Errorf("failed to get book: %w", err)
		}
		books = append(books, struct {
			id       string
			title    string
			genre    string
			audience string
		}{book.ID, book.Title, book.Genre, book.TargetAudience})
	} else {
		// All books
		allBooks, err := booksRepo.GetAll()
		if err != nil {
			return fmt.Errorf("failed to get books: %w", err)
		}
		if len(allBooks) == 0 {
			fmt.Println("No books in catalog. Add one with 'gagipress books add'")
			return nil
		}
		for _, book := range allBooks {
			books = append(books, struct {
				id       string
				title    string
				genre    string
				audience string
			}{book.ID, book.Title, book.Genre, book.TargetAudience})
		}
	}

	fmt.Printf("📚 Generating ideas for %d book(s)\n", len(books))
	fmt.Printf("🎯 Target: %d ideas per book\n\n", count)

	// Create generator
	gen := generator.NewIdeaGenerator(cfg, useGemini)

	totalGenerated := 0
	totalSaved := 0

	for _, book := range books {
		fmt.Printf("📖 Book: %s\n", book.title)
		fmt.Printf("   Genre: %s\n", book.genre)

		// Determine niche from genre
		niche := determineNiche(book.genre)
		fmt.Printf("   Niche: %s\n\n", niche)

		// Generate ideas
		spinner := ui.NewSpinner(fmt.Sprintf("Generating %d ideas...", count))
		spinner.Start()
		ideas, err := gen.GenerateIdeas(book.title, book.genre, book.audience, niche, count)
		spinner.Stop()

		if err != nil {
			ui.Error(fmt.Sprintf("Generation failed: %v", err))
			fmt.Println()
			continue
		}

		ui.Success(fmt.Sprintf("Generated %d ideas", len(ideas)))
		totalGenerated += len(ideas)

		// Save to database
		spinner = ui.NewSpinner("Saving to database...")
		spinner.Start()
		savedIdeas, err := gen.SaveIdeas(ideas, &book.id)
		spinner.Stop()

		if err != nil {
			ui.Warning(fmt.Sprintf("Save failed: %v", err))
			fmt.Println()
			continue
		}

		ui.Success(fmt.Sprintf("Saved %d ideas", len(savedIdeas)))
		fmt.Println()
		totalSaved += len(savedIdeas)
	}

	fmt.Println("═══════════════════════════")
	fmt.Printf("✅ Generation Complete!\n")
	fmt.Printf("   Total generated: %d ideas\n", totalGenerated)
	fmt.Printf("   Total saved: %d ideas\n\n", totalSaved)

	fmt.Println("Next steps:")
	fmt.Println("  • Review ideas: gagipress ideas list")
	fmt.Println("  • Approve ideas: gagipress ideas approve <id>")
	fmt.Println("  • Reject ideas: gagipress ideas reject <id>")

	return nil
}

// determineNiche determines the book niche from genre
func determineNiche(genre string) prompts.BookNiche {
	genreLower := genre
	if len(genreLower) > 0 {
		genreLower = genre[:1] + genre[1:]
	}

	switch {
	case contains(genreLower, "children", "bambini", "kids"):
		return prompts.ChildrenBooks
	case contains(genreLower, "puzzle", "enigmi", "quiz"):
		return prompts.Puzzles
	case contains(genreLower, "dialect", "dialetto", "milanese"):
		return prompts.DialectPuzzles
	case contains(genreLower, "saving", "risparmio", "money"):
		return prompts.Savings
	default:
		return prompts.Puzzles // default
	}
}

// contains checks if str contains any of the substrings
func contains(str string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
