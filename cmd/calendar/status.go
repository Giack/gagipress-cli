package calendar

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show publishing status summary",
	Long:  `Display a count of calendar entries by status (approved, publishing, published, failed).`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("📊 Calendar Status"))

	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
	counts, err := calendarRepo.GetStatusCounts()
	if err != nil {
		return fmt.Errorf("failed to get status counts: %w", err)
	}

	total := 0
	for _, n := range counts {
		total += n
	}

	if total == 0 {
		fmt.Println("No calendar entries found.")
		fmt.Println("\nCreate a plan with: gagipress calendar plan")
		return nil
	}

	statuses := []struct {
		key   string
		label string
	}{
		{"pending_approval", "Pending approval"},
		{"approved", "Approved (queued)"},
		{"publishing", "Publishing (in-flight)"},
		{"published", "Published"},
		{"failed", "Failed"},
	}

	for _, s := range statuses {
		n := counts[s.key]
		if n == 0 {
			continue
		}
		var rendered string
		switch s.key {
		case "failed":
			rendered = ui.StyleError.Render(fmt.Sprintf("%d", n))
		case "published":
			rendered = ui.StyleSuccess.Render(fmt.Sprintf("%d", n))
		case "approved":
			rendered = ui.StyleSuccess.Render(fmt.Sprintf("%d", n))
		default:
			rendered = ui.StyleWarning.Render(fmt.Sprintf("%d", n))
		}
		fmt.Printf("  %-26s %s\n", s.label+":", rendered)
	}

	fmt.Printf("\n  %-26s %d\n", "Total:", total)

	if counts["failed"] > 0 {
		fmt.Printf("\n%s\n", ui.StyleMuted.Render(
			fmt.Sprintf("💡 %d failed entries — use 'gagipress calendar retry' to re-queue them", counts["failed"]),
		))
	}

	return nil
}
