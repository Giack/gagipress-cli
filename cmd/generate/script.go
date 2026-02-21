package generate

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/generator"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	platform       string
	scriptUseGemini bool
)

var scriptCmd = &cobra.Command{
	Use:   "script [idea-id]",
	Short: "Generate a complete script from an approved idea",
	Long: `Generate a complete TikTok/Instagram Reels script from an approved content idea.

The generator will:
  - Read the approved idea from database
  - Generate hook, main content, and CTA
  - Suggest hashtags and music
  - Provide video editing notes
  - Save script to database
  - Mark idea as "scripted"

The idea must be in "approved" status to generate a script.`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerateScript,
}

func init() {
	scriptCmd.Flags().StringVar(&platform, "platform", "tiktok", "Target platform (tiktok or instagram)")
	scriptCmd.Flags().BoolVar(&scriptUseGemini, "gemini", false, "Use Gemini instead of OpenAI")

	GenerateCmd.AddCommand(scriptCmd)
}

func runGenerateScript(cmd *cobra.Command, args []string) error {
	ideaID := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📝 Script Generator")
	fmt.Println("═══════════════════")

	// Get idea from database
	contentRepo := repository.NewContentRepository(&cfg.Supabase)
	booksRepo := repository.NewBooksRepository(&cfg.Supabase)

	fmt.Print("Loading idea... ")
	ideas, err := contentRepo.GetIdeas("", 0)
	if err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to get ideas: %w", err)
	}

	var idea *models.ContentIdea
	for i := range ideas {
		if ideas[i].ID == ideaID || ideas[i].ID[:8] == ideaID {
			idea = &ideas[i]
			break
		}
	}

	if idea == nil {
		fmt.Println("❌ NOT FOUND")
		return fmt.Errorf("idea not found: %s", ideaID)
	}
	fmt.Println("✅ OK")

	// Check status
	if idea.Status != "approved" {
		return fmt.Errorf("idea must be approved first (current status: %s)", idea.Status)
	}

	fmt.Printf("\n💡 Idea: %s\n", idea.BriefDescription)
	fmt.Printf("   Type: %s\n", idea.Type)
	fmt.Printf("   Platform: %s\n\n", platform)

	// Get book info
	bookTitle := "Your Book" // default
	amazonURL := ""
	if idea.BookID != nil {
		book, err := booksRepo.GetByID(*idea.BookID)
		if err == nil {
			bookTitle = book.Title
			if book.KDPASIN != "" {
				amazonURL = fmt.Sprintf("https://www.amazon.it/dp/%s", book.KDPASIN)
			}
		}
	}

	// Generate script
	gen := generator.NewScriptGenerator(cfg, scriptUseGemini)

	spinner := ui.NewSpinner("Generating script with AI...")
	spinner.Start()
	script, err := gen.GenerateScript(idea, bookTitle, platform, amazonURL)
	spinner.Stop()

	if err != nil {
		ui.Error(fmt.Sprintf("Script generation failed: %v", err))
		return err
	}

	ui.Success("Script generated!")
	fmt.Println("\n" + repeatStr("═", 60))

	// Display script
	fmt.Println("\n🎬 HOOK")
	fmt.Println(repeatStr("─", 60))
	fmt.Println(script.Hook)

	fmt.Println("\n📄 MAIN CONTENT")
	fmt.Println(repeatStr("─", 60))
	fmt.Println(script.MainContent)

	fmt.Println("\n🎯 CALL-TO-ACTION")
	fmt.Println(repeatStr("─", 60))
	fmt.Println(script.CTA)

	fmt.Println("\n🏷️  HASHTAGS")
	fmt.Println(repeatStr("─", 60))
	for _, tag := range script.Hashtags {
		fmt.Printf("%s ", tag)
	}
	fmt.Println()

	if script.MusicSuggestion != "" {
		fmt.Println("\n🎵 MUSIC SUGGESTION")
		fmt.Println(repeatStr("─", 60))
		fmt.Println(script.MusicSuggestion)
	}

	if script.VideoNotes != "" {
		fmt.Println("\n🎥 VIDEO NOTES")
		fmt.Println(repeatStr("─", 60))
		fmt.Println(script.VideoNotes)
	}

	fmt.Printf("\n⏱️  Estimated Length: %d seconds\n", script.EstimatedLength)
	fmt.Println(repeatStr("═", 60))

	// Save to database
	fmt.Print("\n💾 Saving script... ")
	savedScript, err := gen.SaveScript(script, idea.ID)
	if err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to save script: %w", err)
	}
	fmt.Println("✅ OK")

	fmt.Printf("\n✅ Script created successfully!\n")
	fmt.Printf("   Script ID: %s\n", savedScript.ID)
	fmt.Println("\nNext steps:")
	fmt.Println("  • Review and edit if needed")
	fmt.Println("  • Create video content")
	fmt.Println("  • Schedule for publishing")

	return nil
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
