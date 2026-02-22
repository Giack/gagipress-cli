package calendar

import (
	"fmt"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/models"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	statusFilter string
	daysAhead    int
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show scheduled content calendar",
	Long:  `Display the content calendar with all scheduled posts.`,
	RunE:  runShow,
}

func init() {
	showCmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (pending_approval, approved, published)")
	showCmd.Flags().IntVar(&daysAhead, "days", 14, "Show next N days")
}

func runShow(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	header := "📅 Content Calendar"
	if statusFilter != "" {
		header += fmt.Sprintf(" (%s)", statusFilter)
	}
	fmt.Println(ui.StyleHeader.Render(header))
	fmt.Printf("Showing schedule for next %d days\n\n", daysAhead)

	// Get calendar entries from database
	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
	entries, err := calendarRepo.GetEntries(statusFilter, 0)
	if err != nil {
		return fmt.Errorf("failed to get calendar entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No scheduled posts found.")
		fmt.Println("\nCreate a plan with: gagipress calendar plan")
		return nil
	}

	// Group by date
	byDate := make(map[string][]models.ContentCalendar)
	for _, entry := range entries {
		dateKey := entry.ScheduledFor.Format("2006-01-02")
		byDate[dateKey] = append(byDate[dateKey], entry)
	}

	// Display calendar
	currentDate := ""
	for date, dayEntries := range byDate {
		parsedDate, _ := time.Parse("2006-01-02", date)

		if date != currentDate {
			currentDate = date
			dateHeader := ui.StyleHeader.Render(
				"📆 " + parsedDate.Format("Monday, January 2, 2006"),
			)
			fmt.Println(dateHeader)
		}

		for _, entry := range dayEntries {
			// Format status with color
			var status string
			switch entry.Status {
			case "pending_approval":
				status = ui.FormatStatus("pending")
			case "approved":
				status = ui.FormatStatus("approved")
			case "published":
				status = ui.StyleSuccess.Render("published")
			case "failed":
				status = ui.StyleError.Render("failed")
			default:
				status = entry.Status
			}

			time := ui.StyleMuted.Render(entry.ScheduledFor.Format("15:04"))
			entryID := entry.ID
			if len(entryID) > 8 {
				entryID = entryID[:8] + "…"
			}

			fmt.Printf("  %s | %s | %s | %s\n",
				time,
				entry.Platform,
				status,
				entryID,
			)
		}
		fmt.Println()
	}

	// Summary
	total := len(entries)
	pending := 0
	approved := 0
	published := 0

	for _, entry := range entries {
		switch entry.Status {
		case "pending_approval":
			pending++
		case "approved":
			approved++
		case "published":
			published++
		}
	}

	fmt.Println(ui.StyleHeader.Render("Summary"))
	summaryText := fmt.Sprintf("Total: %d posts | Pending: %s | Approved: %s | Published: %s",
		total,
		ui.StyleWarning.Render(fmt.Sprintf("%d", pending)),
		ui.StyleSuccess.Render(fmt.Sprintf("%d", approved)),
		ui.StyleSuccess.Render(fmt.Sprintf("%d", published)),
	)
	fmt.Println(summaryText)

	if pending > 0 {
		fmt.Printf("\n%s\n", ui.StyleMuted.Render("💡 Use 'gagipress calendar approve' to approve pending posts"))
	}

	return nil
}
