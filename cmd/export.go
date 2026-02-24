package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gtasks2md/internal/sync"
)

var exportListName string

var exportCmd = &cobra.Command{
	Use:   "export [output_path]",
	Short: "Exports task lists from Google Tasks to local Markdown files.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputPath := "."
		if len(args) > 0 {
			outputPath = args[0]
		}
		
		err := sync.ExportTasks(outputPath, exportListName, credentialsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportListName, "list-name", "l", "", "Specify a single Google Task list name to export (required if output_path is a single file).")
}
