package ideas

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var rejectCmd = &cobra.Command{
	Use:   "reject [idea-id]",
	Short: "Reject a content idea",
	Long:  `Mark a content idea as rejected.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runReject,
}

func runReject(cmd *cobra.Command, args []string) error {
	ideaID := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("❌ Rejecting Idea"))
	fmt.Println()

	repo := repository.NewContentRepository(&cfg.Supabase)

	// Resolve ID prefix to full UUID
	fmt.Print(ui.StyleMuted.Render("Resolving idea ID... "))
	idea, err := repo.GetIdeaByIDPrefix(ideaID)
	if err != nil {
		fmt.Println(ui.StyleError.Render("✗ FAILED"))
		return fmt.Errorf("failed to resolve idea ID: %w", err)
	}
	ideaID = idea.ID
	fmt.Println(ui.StyleSuccess.Render("✓ " + ideaID))

	// Update status
	fmt.Print(ui.StyleMuted.Render("Updating status... "))
	if err := repo.UpdateIdeaStatus(ideaID, "rejected"); err != nil {
		fmt.Println(ui.StyleError.Render("✗ FAILED"))
		return fmt.Errorf("failed to reject idea: %w", err)
	}

	fmt.Println(ui.StyleSuccess.Render("✓ OK"))
	fmt.Printf("\n%s\n", ui.StyleError.Render(fmt.Sprintf("❌ Idea %s rejected", ideaID)))

	return nil
}
