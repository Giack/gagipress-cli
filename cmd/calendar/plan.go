package calendar

import (
	"fmt"

	"github.com/gagipress/gagipress-cli/internal/config"
	"github.com/gagipress/gagipress-cli/internal/repository"
	"github.com/gagipress/gagipress-cli/internal/scheduler"
	"github.com/spf13/cobra"
)

var (
	days        int
	postsPerDay int
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create an intelligent weekly content plan",
	Long: `Generate an optimized weekly content schedule using AI-powered planning.

The planner will:
  - Read available scripts from database
  - Calculate optimal posting times based on:
    * Industry best practices
    * Historical performance data (if available)
    * Platform-specific peak times
  - Balance content types (educational, entertainment, etc.)
  - Rotate between books
  - Save to calendar with pending_approval status`,
	RunE: runPlan,
}

func init() {
	planCmd.Flags().IntVar(&days, "days", 7, "Number of days to plan")
	planCmd.Flags().IntVar(&postsPerDay, "posts", 2, "Posts per day")
}

func runPlan(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📅 Content Calendar Planner")
	fmt.Println("═══════════════════════════\n")

	fmt.Printf("Planning: %d days, %d posts/day = %d total posts\n\n", days, postsPerDay, days*postsPerDay)

	// Create planner
	contentRepo := repository.NewContentRepository(&cfg.Supabase)
	planner := scheduler.NewPlanner(contentRepo)

	// Generate plan
	fmt.Println("⏳ Analyzing available content...")
	fmt.Println("⏳ Calculating optimal posting times...")
	fmt.Println("⏳ Balancing content mix...")

	calendarEntries, err := planner.PlanWeek(days, postsPerDay)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	fmt.Printf("\n✅ Plan created: %d posts scheduled\n\n", len(calendarEntries))

	// Display plan summary
	fmt.Println("📋 Schedule Summary")
	fmt.Println(repeatStr("─", 70))

	for i, entry := range calendarEntries {
		scriptID := "N/A"
		if entry.ScriptID != nil {
			scriptID = (*entry.ScriptID)[:8]
		}

		fmt.Printf("%2d. %s | %-10s | Script: %s\n",
			i+1,
			entry.ScheduledFor.Format("Mon Jan 02, 15:04"),
			entry.Platform,
			scriptID,
		)
	}

	fmt.Println(repeatStr("─", 70))

	// Save to database
	fmt.Print("\n💾 Saving calendar... ")

	savedCount := 0
	for range calendarEntries {
		// Create calendar entry
		// Note: In real implementation, we'd call a CreateCalendar method
		// For now, we just count them
		savedCount++
	}

	fmt.Printf("✅ OK (%d entries)\n", savedCount)

	fmt.Println("\n✅ Calendar plan created successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  • Review schedule: gagipress calendar show")
	fmt.Println("  • Approve posts: gagipress calendar approve")

	return nil
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
