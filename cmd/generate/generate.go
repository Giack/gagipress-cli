package generate

import (
	"github.com/spf13/cobra"
)

// GenerateCmd represents the generate command group
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate content ideas and scripts",
	Long: `Generate AI-powered content for your social media:
  - Generate content ideas from your books
  - Generate scripts from approved ideas
  - Batch generation for weekly planning`,
}

func init() {
	GenerateCmd.AddCommand(ideasCmd)
}
