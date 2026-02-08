package test

import (
	"github.com/spf13/cobra"
)

// TestCmd represents the test command group
var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test various integrations and features",
	Long:  `Test command group for testing browser automation, APIs, and other features.`,
}

func init() {
	TestCmd.AddCommand(geminiCmd)
}
