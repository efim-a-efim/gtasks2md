package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gtasks2md/internal/sync"
)

var importListName string

var importCmd = &cobra.Command{
	Use:   "import <input_path>",
	Short: "Imports task lists from local Markdown files to Google Tasks.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]
		
		err := sync.ImportTasks(inputPath, importListName, credentialsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&importListName, "list-name", "l", "", "Target Google Tasks list name (optional override).")
}
