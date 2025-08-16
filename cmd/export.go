package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/spf13/cobra"
	"os"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export journal data",
}

var exportJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Export journal as JSON",
	Run:   runExportJson,
}

func runExportJson(cmd *cobra.Command, args []string) {
	journal, _, err := parser.ParseFile("hfl.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonData, err := json.MarshalIndent(journal.Entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}

func init() {
	exportCmd.AddCommand(exportJsonCmd)
	RootCmd.AddCommand(exportCmd)
}
