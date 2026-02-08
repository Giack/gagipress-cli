package db

import (
	"github.com/spf13/cobra"
)

// DbCmd represents the db command
var DbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Manage database migrations, status, and operations.`,
}

func init() {
	DbCmd.AddCommand(migrateCmd)
	DbCmd.AddCommand(statusCmd)
}
