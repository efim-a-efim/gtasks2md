package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var credentialsPath string

var rootCmd = &cobra.Command{
	Use:   "gtasks2md",
	Short: "Google Tasks to Markdown Sync",
	Long:  `A CLI tool to synchronize Google Tasks with local Markdown files.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&credentialsPath, "credentials", "c", "", "Path to the OAuth 2.0 credentials.json file.")
}
