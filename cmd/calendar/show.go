package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/gagipress/gagipress-cli/internal/config"
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
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📅 Content Calendar")
	fmt.Println("═══════════════════")

	// Get calendar entries
	// Note: We'd need to add a GetCalendar method to the repository
	// For now, we'll show a placeholder

	fmt.Printf("Showing schedule for next %d days", daysAhead)
	if statusFilter != "" {
		fmt.Printf(" (status: %s)", statusFilter)
	}
	fmt.Println()

	// Placeholder data structure
	type CalendarEntry struct {
		ID           string
		ScheduledFor time.Time
		Platform     string
		Status       string
		ScriptID     string
	}

	// In real implementation, fetch from database
	entries := []CalendarEntry{
		// Placeholder - would come from repository.GetCalendar()
	}

	if len(entries) == 0 {
		fmt.Println("No scheduled posts found.")
		fmt.Println("\nCreate a plan with: gagipress calendar plan")
		return nil
	}

	// Group by date
	byDate := make(map[string][]CalendarEntry)
	for _, entry := range entries {
		dateKey := entry.ScheduledFor.Format("2006-01-02")
		byDate[dateKey] = append(byDate[dateKey], entry)
	}

	// Display calendar
	for date, dayEntries := range byDate {
		parsedDate, _ := time.Parse("2006-01-02", date)
		fmt.Printf("📆 %s\n", parsedDate.Format("Monday, January 2, 2006"))
		fmt.Println(strings.Repeat("─", 70))

		for _, entry := range dayEntries {
			statusEmoji := "⏳"
			switch entry.Status {
			case "approved":
				statusEmoji = "✅"
			case "published":
				statusEmoji = "🎉"
			case "failed":
				statusEmoji = "❌"
			}

			fmt.Printf("%s %s | %-10s | %-18s | %s\n",
				statusEmoji,
				entry.ScheduledFor.Format("15:04"),
				entry.Platform,
				entry.Status,
				entry.ID[:8],
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

	fmt.Println(strings.Repeat("═", 70))
	fmt.Printf("Total: %d posts | Pending: %d | Approved: %d | Published: %d\n",
		total, pending, approved, published)

	if pending > 0 {
		fmt.Println("\n💡 Use 'gagipress calendar approve' to approve pending posts")
	}

	return nil
}
