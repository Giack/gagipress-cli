package calendar

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
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

	fmt.Println("✅ Calendar Approval")
	fmt.Println("════════════════════")

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
		fmt.Printf("Entry %d/%d\n", i+1, len(entries))
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("ID:           %s\n", entry.ID[:8])
		fmt.Printf("Scheduled:    %s\n", entry.ScheduledFor.Format("Mon Jan 02, 2006 at 15:04"))
		fmt.Printf("Platform:     %s\n", entry.Platform)
		if entry.ScriptID != nil {
			fmt.Printf("Script ID:    %s\n", (*entry.ScriptID)[:8])
		}
		fmt.Println(strings.Repeat("─", 60))

		fmt.Print("\n[A]pprove / [S]kip / [R]eject? ")
		action, _ := reader.ReadString('\n')
		action = strings.ToUpper(strings.TrimSpace(action))

		switch action {
		case "A":
			if err := calendarRepo.UpdateEntryStatus(entry.ID, "approved"); err != nil {
				fmt.Printf("❌ Failed to approve: %v\n\n", err)
				continue
			}
			fmt.Println("✅ Approved")
			approved++

		case "R":
			if err := calendarRepo.DeleteEntry(entry.ID); err != nil {
				fmt.Printf("❌ Failed to reject: %v\n\n", err)
				continue
			}
			fmt.Println("❌ Rejected")
			rejected++

		case "S":
			fmt.Println("⏭️  Skipped")
			skipped++

		default:
			fmt.Println("⏭️  Invalid input, skipped")
			skipped++
		}
	}

	// Summary
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("Summary: %d approved | %d rejected | %d skipped\n", approved, rejected, skipped)

	if approved > 0 {
		fmt.Printf("\n✅ %d entries approved and ready for publishing\n", approved)
	}

	return nil
}
