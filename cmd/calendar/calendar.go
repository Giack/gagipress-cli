package calendar

import (
	"github.com/spf13/cobra"
)

// CalendarCmd represents the calendar command group
var CalendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Manage content calendar and scheduling",
	Long: `Manage your content publishing calendar:
  - Plan weekly content schedule
  - View scheduled posts
  - Approve or modify schedule
  - Force publish immediately`,
}

func init() {
	CalendarCmd.AddCommand(planCmd)
	CalendarCmd.AddCommand(showCmd)
	CalendarCmd.AddCommand(approveCmd)
}
