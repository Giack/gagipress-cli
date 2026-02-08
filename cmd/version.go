package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version   = "0.1.0"
	BuildDate = "2026-02-08"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Gagipress CLI",
	Long:  `Display the current version and build information of Gagipress CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Gagipress CLI v%s (built %s)\n", Version, BuildDate)
		fmt.Println("https://github.com/gagipress/gagipress-cli")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
