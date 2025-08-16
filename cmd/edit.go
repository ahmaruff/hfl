package cmd

import (
	"fmt"
	"github.com/ahmaruff/hfl/internal/config"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/ahmaruff/hfl/internal/writer"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var editCmd = &cobra.Command{
	Use:   "edit [DATE]",
	Short: "Edit journal entry for a specific date",
	Long:  "Opens hfl.md in your editor. If DATE is provided, positions cursor at that entry. If no DATE, uses today.",
	Run:   runEdit,
}

func runEdit(cmd *cobra.Command, args []string) {
	// Determine date
	var date string
	if len(args) > 0 {
		date = args[0] // TODO: Validate YYYY-MM-DD format
	} else {
		date = time.Now().Format("2006-01-02") // Today
	}

	// Ensure hfl.md exists and has the entry
	ensureEntryExists(date)

	// Open editor
	openEditor("hfl.md")
}

func ensureEntryExists(date string) {
	// Try to parse existing file
	journal, _, err := parser.ParseFile("hfl.md")
	if err != nil {
		// File doesn't exist, create new journal
		journal = &parser.Journal{Entries: []parser.Entry{}}
	}

	// Check if entry exists
	entryExists := false
	for _, entry := range journal.Entries {
		if entry.Date == date {
			entryExists = true
			break
		}
	}

	// If entry doesn't exist, create it
	if !entryExists {
		newEntry := parser.Entry{
			Date: date,
			Body: "", // Empty body
		}
		journal.Entries = append(journal.Entries, newEntry)

		// Write back to file
		err = writer.WriteFile("hfl.md", journal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating entry: %v\n", err)
			os.Exit(1)
		}
	}
}

func openEditor(filename string) {
	cfg, err := config.Load()
	if err != nil {
		// Fallback if config fails using empty config
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = &config.Config{}
	}

	editor := cfg.GetEditor()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", editor, filename)
	} else {
		cmd = exec.Command("sh", "-c", editor+" "+filename)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(editCmd)
}
