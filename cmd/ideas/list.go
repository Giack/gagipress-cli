package ideas

import (
	"fmt"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/spf13/cobra"
)

var (
	statusFilter string
	limitList    int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List content ideas",
	Long:  `Display all generated content ideas with optional filters.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (pending, approved, rejected, scripted)")
	listCmd.Flags().IntVar(&limitList, "limit", 50, "Maximum number of ideas to show")
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("💡 Content Ideas")
	fmt.Println("═════════════════\n")

	// Get ideas
	repo := repository.NewContentRepository(&cfg.Supabase)
	ideas, err := repo.GetIdeas(statusFilter, limitList)
	if err != nil {
		return fmt.Errorf("failed to get ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("No ideas found. Generate some with 'gagipress generate ideas'")
		return nil
	}

	// Print table header
	fmt.Printf("%-8s %-12s %-15s %-50s %-6s\n", "ID", "Type", "Status", "Description", "Score")
	fmt.Println(strings.Repeat("─", 100))

	// Print ideas
	for _, idea := range ideas {
		id := idea.ID
		if len(id) > 8 {
			id = id[:8]
		}

		ideaType := idea.Type
		if len(ideaType) > 12 {
			ideaType = ideaType[:9] + "..."
		}

		status := idea.Status
		if len(status) > 15 {
			status = status[:12] + "..."
		}

		desc := idea.BriefDescription
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		score := "N/A"
		if idea.RelevanceScore != nil {
			score = fmt.Sprintf("%d", *idea.RelevanceScore)
		}

		fmt.Printf("%-8s %-12s %-15s %-50s %-6s\n",
			id, ideaType, status, desc, score)
	}

	fmt.Printf("\nTotal ideas: %d\n", len(ideas))

	if statusFilter == "" || statusFilter == "pending" {
		pendingCount := 0
		for _, idea := range ideas {
			if idea.Status == "pending" {
				pendingCount++
			}
		}
		if pendingCount > 0 {
			fmt.Printf("\n💡 %d pending ideas awaiting approval\n", pendingCount)
			fmt.Println("   Use 'gagipress ideas approve <id>' to approve")
			fmt.Println("   Use 'gagipress ideas reject <id>' to reject")
		}
	}

	return nil
}
