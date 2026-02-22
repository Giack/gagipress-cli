package calendar

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry failed calendar entries",
	Long:  `Reset all 'failed' calendar entries back to 'approved' so the cron job will re-publish them.`,
	RunE:  runRetry,
}

func runRetry(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)

	// Check how many are currently failed before resetting
	counts, err := calendarRepo.GetStatusCounts()
	if err != nil {
		return fmt.Errorf("failed to get status counts: %w", err)
	}

	failedCount := counts["failed"]
	if failedCount == 0 {
		fmt.Println(ui.StyleSuccess.Render("✓ No failed entries to retry."))
		return nil
	}

	fmt.Printf("Retrying %s failed entries...\n", ui.StyleWarning.Render(fmt.Sprintf("%d", failedCount)))

	retried, err := calendarRepo.RetryFailed()
	if err != nil {
		return fmt.Errorf("failed to retry entries: %w", err)
	}

	fmt.Println(ui.StyleSuccess.Render(fmt.Sprintf("✓ %d entries reset to 'approved' — the cron will publish them within 15 minutes.", retried)))

	return nil
}
