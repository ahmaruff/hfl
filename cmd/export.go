package cmd

import (
	"fmt"
	"github.com/ahmaruff/hfl/internal/export"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/spf13/cobra"
	"os"
)

var outputFile string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export journal data",
}

var exportJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Export journal as JSON",
	Run:   runExportJson,
}

var exportCsvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Export journal as CSV",
	Run:   runExportCsv,
}

func runExportJson(cmd *cobra.Command, args []string) {
	journal, _, err := parser.ParseFile("hfl.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	filename := outputFile
	if filename == "" {
		filename = "hfl.json"
	}

	err = export.ToJSONFile(journal, filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported to %s\n", filename)
}

func runExportCsv(cmd *cobra.Command, args []string) {
	journal, _, err := parser.ParseFile("hfl.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	filename := outputFile
	if filename == "" {
		filename = "hfl.csv"
	}

	err = export.ToCSVFile(journal, filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported to %s\n", filename)
}

func init() {
	// Add --output flag
	exportCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output filename")

	exportCmd.AddCommand(exportJsonCmd)
	exportCmd.AddCommand(exportCsvCmd)
	RootCmd.AddCommand(exportCmd)
}
