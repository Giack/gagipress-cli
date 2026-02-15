package calendar

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/ui"
	"github.com/spf13/cobra"
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve pending calendar entries",
	Long: `Review and approve pending calendar entries interactively.

For each pending entry, you can:
  - Approve: Mark as approved for publishing
  - Skip: Leave in pending status
  - Reject: Remove from calendar`,
	RunE: runApprove,
}

func init() {
	CalendarCmd.AddCommand(approveCmd)
}

func runApprove(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(ui.StyleHeader.Render("✅ Calendar Approval"))

	// Get pending entries
	calendarRepo := repository.NewCalendarRepository(&cfg.Supabase)
	entries, err := calendarRepo.GetEntries("pending_approval", 0)
	if err != nil {
		return fmt.Errorf("failed to get entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No pending entries to approve.")
		fmt.Println("\nCreate a plan with: gagipress calendar plan")
		return nil
	}

	fmt.Printf("Found %d pending entries\n\n", len(entries))

	reader := bufio.NewReader(os.Stdin)
	approved := 0
	rejected := 0
	skipped := 0

	for i, entry := range entries {
		// Create entry preview box
		previewStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorPrimary).
			Padding(1, 2).
			Width(60)

		scriptInfo := "N/A"
		if entry.ScriptID != nil {
			scriptID := *entry.ScriptID
			if len(scriptID) > 8 {
				scriptID = scriptID[:8] + "…"
			}
			scriptInfo = scriptID
		}

		entryContent := fmt.Sprintf(
			"%s\n"+
				"Scheduled:  %s\n"+
				"Platform:   %s\n"+
				"Script ID:  %s",
			ui.StyleHeader.Render(fmt.Sprintf("Entry %d/%d", i+1, len(entries))),
			ui.FormatDate(entry.ScheduledFor),
			entry.Platform,
			scriptInfo,
		)

		fmt.Println(previewStyle.Render(entryContent))

		// Color-coded prompt
		promptText := ui.StyleSuccess.Render("[A]pprove") + " / " +
			ui.StyleMuted.Render("[S]kip") + " / " +
			ui.StyleError.Render("[R]eject") + "? "
		fmt.Print(promptText)
		action, _ := reader.ReadString('\n')
		action = strings.ToUpper(strings.TrimSpace(action))

		switch action {
		case "A":
			if err := calendarRepo.UpdateEntryStatus(entry.ID, "approved"); err != nil {
				fmt.Printf("%s\n\n", ui.StyleError.Render("❌ Failed to approve: "+err.Error()))
				continue
			}
			fmt.Println(ui.StyleSuccess.Render("✅ Approved"))
			approved++

		case "R":
			if err := calendarRepo.DeleteEntry(entry.ID); err != nil {
				fmt.Printf("%s\n\n", ui.StyleError.Render("❌ Failed to reject: "+err.Error()))
				continue
			}
			fmt.Println(ui.StyleError.Render("❌ Rejected"))
			rejected++

		case "S":
			fmt.Println(ui.StyleMuted.Render("⏭️  Skipped"))
			skipped++

		default:
			fmt.Println(ui.StyleMuted.Render("⏭️  Invalid input, skipped"))
			skipped++
		}
		fmt.Println()
	}

	// Summary
	fmt.Println(ui.StyleHeader.Render("Summary"))
	summaryText := fmt.Sprintf("%s approved | %s rejected | %s skipped",
		ui.StyleSuccess.Render(fmt.Sprintf("%d", approved)),
		ui.StyleError.Render(fmt.Sprintf("%d", rejected)),
		ui.StyleMuted.Render(fmt.Sprintf("%d", skipped)),
	)
	fmt.Println(summaryText)

	if approved > 0 {
		fmt.Printf("\n%s\n", ui.StyleSuccess.Render(fmt.Sprintf("✅ %d entries approved and ready for publishing", approved)))
	}

	return nil
}
