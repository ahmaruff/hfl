package cmd

import (
	"fmt"
	"github.com/ahmaruff/hfl/internal/gitignore"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/spf13/cobra"
	"os"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check hfl.md for formatting issues",
	Long:  "Parse hfl.md and report any warnings or formatting issues.",
	Run:   runCheck,
}

func runCheck(cmd *cobra.Command, args []string) {

	// Ensure .hfl/ is gitignored (before creating any .hfl files)
	if err := gitignore.EnsureHFLIgnored(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update .gitignore: %v\n", err)
	}

	_, warnings, err := parser.ParseFile("hfl.md")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, warning := range warnings {
		fmt.Println(warning)
	}

	if len(warnings) > 0 {
		os.Exit(2) // Exit code 2 for warnings
	}

	fmt.Println("hfl.md is valid")
	// Exit code 0 (success)
}

func init() {
	RootCmd.AddCommand(checkCmd)
}
