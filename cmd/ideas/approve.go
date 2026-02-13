package ideas

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var approveCmd = &cobra.Command{
	Use:   "approve [idea-id]",
	Short: "Approve a content idea",
	Long:  `Mark a content idea as approved for script generation.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runApprove,
}

func runApprove(cmd *cobra.Command, args []string) error {
	ideaID := args[0]

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("✅ Approving Idea")
	fmt.Println("═════════════════")

	// Update status
	repo := repository.NewContentRepository(&cfg.Supabase)
	fmt.Print("Updating status... ")
	if err := repo.UpdateIdeaStatus(ideaID, "approved"); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("failed to approve idea: %w", err)
	}

	fmt.Println("✅ OK")
	fmt.Printf("\n✅ Idea %s approved!\n", ideaID)
	fmt.Println("\nNext step:")
	fmt.Printf("  • Generate script: gagipress generate script %s\n", ideaID)

	return nil
}
