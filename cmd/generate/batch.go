package generate

import (
	"fmt"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/generator"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	batchPlatform  string
	batchUseGemini bool
	batchLimit     int
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Generate scripts for all approved ideas",
	Long: `Generate TikTok/Instagram Reels scripts for all approved content ideas in the database.

The batch generator will:
  - Query all ideas with 'approved' status
  - Automatically fetch book metadata and ASIN for each idea
  - Generate complete scripts (hook, content, CTA, hashtags)
  - Save scripts to the database
  - Update idea status to 'scripted'`,
	RunE: runBatch,
}

func init() {
	batchCmd.Flags().StringVar(&batchPlatform, "platform", "tiktok", "Target platform for all scripts (tiktok or instagram)")
	batchCmd.Flags().BoolVar(&batchUseGemini, "gemini", false, "Use Gemini instead of OpenAI")
	batchCmd.Flags().IntVar(&batchLimit, "limit", 10, "Maximum number of scripts to generate in this batch")

	GenerateCmd.AddCommand(batchCmd)
}

func runBatch(cmd *cobra.Command, args []string) error {
	// Validate platform
	if batchPlatform != "tiktok" && batchPlatform != "instagram" {
		return fmt.Errorf("invalid platform: %s (must be tiktok or instagram)", batchPlatform)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("📝 Batch Script Generator"))
	fmt.Printf("Platform: %s | Max Scripts: %d\n\n", batchPlatform, batchLimit)

	contentRepo := repository.NewContentRepository(&cfg.Supabase)
	booksRepo := repository.NewBooksRepository(&cfg.Supabase)

	// Get approved ideas
	ideas, err := contentRepo.GetIdeas("approved", batchLimit)
	if err != nil {
		return fmt.Errorf("failed to get approved ideas: %w", err)
	}

	if len(ideas) == 0 {
		ui.Warning("No approved ideas found. Create ideas with 'gagipress generate ideas' and approve them first.")
		return nil
	}

	fmt.Printf("Found %d approved ideas ready for script generation.\n\n", len(ideas))

	gen := generator.NewScriptGenerator(cfg, batchUseGemini)

	successCount := 0
	failedCount := 0

	for i, idea := range ideas {
		fmt.Printf("[%d/%d] Generating script for idea: %s... ", i+1, len(ideas), idea.ID[:8])

		// 1. Get book info
		bookTitle := "Your Book" // default
		amazonURL := ""
		if idea.BookID != nil {
			book, err := booksRepo.GetByID(*idea.BookID)
			if err == nil {
				bookTitle = book.Title
				if book.KDPASIN != "" {
					// Build Amazon URL with UTM tracking parameters
					amazonURL = fmt.Sprintf("https://www.amazon.it/dp/%s?tag=gagipress-21&utm_source=%s&utm_medium=social&utm_campaign=%s",
						book.KDPASIN, batchPlatform, idea.ID)
				}
			}
		}

		// 2. Generate script
		script, err := gen.GenerateScript(&idea, bookTitle, batchPlatform, amazonURL)
		if err != nil {
			fmt.Printf("❌ Failed (generation error: %v)\n", err)
			failedCount++
			continue
		}

		// 3. Save to database
		_, err = gen.SaveScript(script, idea.ID)
		if err != nil {
			fmt.Printf("❌ Failed (save error: %v)\n", err)
			failedCount++
			continue
		}

		fmt.Printf("✅ Success\n")
		successCount++
	}

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("Batch generation complete!\n")
	fmt.Printf("Total: %d | Success: %d | Failed: %d\n", len(ideas), successCount, failedCount)

	if successCount > 0 {
		fmt.Println("\nNext steps:")
		fmt.Println("  • Plan schedule: gagipress calendar plan")
	}

	return nil
}
