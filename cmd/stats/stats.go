package stats

import (
	"github.com/spf13/cobra"
)

// StatsCmd represents the stats command group
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View analytics and performance statistics",
	Long: `View performance analytics and insights:
  - Social media metrics dashboard
  - Sales data visualization
  - Social → Sales correlation analysis
  - Performance trends`,
}

func init() {
	StatsCmd.AddCommand(showCmd)
	StatsCmd.AddCommand(correlateCmd)
}
