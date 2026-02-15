package ideas

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
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

	// Build table rows
	rows := make([][]string, len(ideas))
	for i, idea := range ideas {
		// Format score
		score := "N/A"
		if idea.RelevanceScore != nil {
			score = fmt.Sprintf("%d", *idea.RelevanceScore)
		}

		// Format status with color
		status := ui.FormatStatus(idea.Status)

		// No manual truncation - let table handle it
		rows[i] = []string{
			ui.FormatUUID(idea.ID, 8),
			idea.Type,
			status,
			idea.BriefDescription, // Full description
			score,
		}
	}

	// Render table
	table := ui.RenderTable(ui.TableConfig{
		Headers:  []string{"ID", "Type", "Status", "Description", "Score"},
		Rows:     rows,
		MaxWidth: ui.GetTerminalWidth(),
	})

	fmt.Println(ui.StyleHeader.Render("💡 Content Ideas"))
	fmt.Println(table)

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
