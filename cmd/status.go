package cmd

import (
	"fmt"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/ahmaruff/hfl/internal/state"
	"github.com/spf13/cobra"
	"os"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status of journal entries",
	Long:  "Display which entries have been modified since last sync.",
	Run:   runStatus,
}

func runStatus(cmd *cobra.Command, args []string) {
	// Load journal
	journal, _, err := parser.ParseFile("hfl.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading journal: %v\n", err)
		os.Exit(1)
	}

	// Load state
	state, err := state.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Check status of each entry
	var newEntries []string
	var modifiedEntries []string
	var syncedEntries []string

	for _, entry := range journal.Entries {
		_, exists := state.GetEntry(entry.Date)

		if !exists {
			// Entry doesn't exist in state = new
			newEntries = append(newEntries, entry.Date)
		} else if state.HasChanged(entry.Date, entry.Body) {
			// Entry exists but content changed = modified
			modifiedEntries = append(modifiedEntries, entry.Date)
		} else {
			// Entry exists and hasn't changed = synced
			syncedEntries = append(syncedEntries, entry.Date)
		}
	}

	// Display results
	fmt.Printf("Journal Status (%d total entries)\n\n", len(journal.Entries))

	if len(newEntries) > 0 {
		fmt.Printf("New entries (%d):\n", len(newEntries))
		for _, date := range newEntries {
			fmt.Printf("  %s\n", date)
		}
		fmt.Println()
	}

	if len(modifiedEntries) > 0 {
		fmt.Printf("Modified entries (%d):\n", len(modifiedEntries))
		for _, date := range modifiedEntries {
			fmt.Printf("  %s\n", date)
		}
		fmt.Println()
	}

	if len(syncedEntries) > 0 {
		fmt.Printf("Synced entries (%d):\n", len(syncedEntries))
		for _, date := range syncedEntries {
			fmt.Printf("  %s\n", date)
		}
		fmt.Println()
	}

	if state.LastSynced != "" {
		fmt.Printf("Last sync: %s\n", state.LastSynced)
	} else {
		fmt.Println("Never synced")
	}
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
