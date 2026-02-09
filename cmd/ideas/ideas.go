package ideas

import (
	"github.com/spf13/cobra"
)

// IdeasCmd represents the ideas command group
var IdeasCmd = &cobra.Command{
	Use:   "ideas",
	Short: "Manage content ideas",
	Long: `Manage generated content ideas:
  - List all ideas with filters
  - Approve ideas for script generation
  - Reject ideas`,
}

func init() {
	IdeasCmd.AddCommand(listCmd)
	IdeasCmd.AddCommand(approveCmd)
	IdeasCmd.AddCommand(rejectCmd)
}
