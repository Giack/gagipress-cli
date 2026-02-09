package cmd

import (
	"fmt"
	"os"

	"github.com/gagipress/gagipress-cli/cmd/auth"
	"github.com/gagipress/gagipress-cli/cmd/books"
	"github.com/gagipress/gagipress-cli/cmd/db"
	"github.com/gagipress/gagipress-cli/cmd/test"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gagipress",
	Short: "Gagipress Social Automation CLI",
	Long: `Gagipress CLI is a powerful tool for automating social media content
generation, scheduling, and analytics for your Amazon KDP publishing business.

Features:
  • AI-powered content generation for TikTok and Instagram Reels
  • Intelligent weekly scheduling with approval workflow
  • Automated publishing via cron jobs
  • Performance analytics with KDP sales correlation
  • Self-hosted with minimal recurring costs`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gagipress/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(db.DbCmd)
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(test.TestCmd)
	rootCmd.AddCommand(books.BooksCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gagipress" (without extension).
		viper.AddConfigPath(home + "/.gagipress")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
