package cmd

import (
	"fmt"
	"github.com/ahmaruff/hfl/internal/config"
	"github.com/ahmaruff/hfl/internal/notion"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/ahmaruff/hfl/internal/state"
	"github.com/ahmaruff/hfl/internal/writer"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize journal with Notion",
	Long:  "Two-way sync between local hfl.md and Notion database.",
	Run:   runSync,
}

var (
	pushOnly bool
	pullOnly bool
	dryRun   bool
)

func runSync(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Notion.ApiToken == "" {
		fmt.Fprintf(os.Stderr, "Error: Notion API token not configured\n")
		fmt.Fprintf(os.Stderr, "Set it with: hfl config set notion.api_token \"your-token\"\n")
		os.Exit(1)
	}

	if cfg.Notion.DatabaseID == "" {
		fmt.Fprintf(os.Stderr, "Error: Notion database ID not configured\n")
		fmt.Fprintf(os.Stderr, "Set it with: hfl config set notion.database_id \"your-db-id\"\n")
		os.Exit(1)
	}

	journal, warnings, err := parser.ParseFile("hfl.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading hfl.md: %v\n", err)
		os.Exit(1)
	}

	if len(warnings) > 0 {
		fmt.Println("⚠️  Warnings in hfl.md:")
		for _, warning := range warnings {
			fmt.Println("  " + warning)
		}
		fmt.Println()
	}

	syncState, err := state.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading sync state: %v\n", err)
		os.Exit(1)
	}

	syncService := notion.NewSyncService(cfg.Notion.ApiToken, cfg.Notion.DatabaseID)

	if err := syncService.ValidateAndSetupDatabase(); err != nil {
		fmt.Fprintf(os.Stderr, "Database schema validation failed: %v\n", err)
	}

	if dryRun {
		fmt.Println("🔍 Dry run mode - no changes will be made")
		showSyncPlan(journal, syncState)
		return
	}

	if pullOnly {
		fmt.Println("⬇️  Pulling changes from Notion...")
		if err := performPullSync(syncService, journal, syncState); err != nil {
			fmt.Fprintf(os.Stderr, "Pull sync failed: %v\n", err)
			os.Exit(1)
		}
	} else if pushOnly {
		fmt.Println("⬆️  Pushing changes to Notion...")
		if err := performPushSync(syncService, journal, syncState); err != nil {
			fmt.Fprintf(os.Stderr, "Push sync failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("🔄 Starting two-way sync...")
		if err := performTwoWaySync(syncService, journal, syncState, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Sync failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("✅ Sync completed successfully!")
}

func performPushSync(syncService *notion.SyncService, journal *parser.Journal, syncState *state.State) error {
	return syncService.SyncToNotion(journal, syncState)
}

func performPullSync(syncService *notion.SyncService, journal *parser.Journal, syncState *state.State) error {
	fmt.Printf("📥 Local journal has %d entries before pull\n", len(journal.Entries))

	err := syncService.SyncFromNotion(journal, syncState)
	if err != nil {
		return err
	}

	fmt.Printf("📥 Local journal has %d entries after pull\n", len(journal.Entries))

	fmt.Println("💾 Writing updated journal to hfl.md...")
	err = writer.WriteFile("hfl.md", journal)
	if err != nil {
		return fmt.Errorf("failed to write updated journal: %w", err)
	}

	fmt.Println("✅ Successfully updated hfl.md")
	return nil
}

func performTwoWaySync(syncService *notion.SyncService, journal *parser.Journal, syncState *state.State, cfg *config.Config) error {
	strategy := cfg.ConflictStrategy
	if strategy == "" {
		strategy = "remote"
	}

	if strategy == "merge" {
		return fmt.Errorf("merge conflict strategy not implemented in this version")
	}

	conflicts, err := detectConflicts(syncService, journal, syncState)
	if err != nil {
		return fmt.Errorf("failed to detect conflicts: %w", err)
	}

	if len(conflicts) > 0 {
		fmt.Printf("⚠️  Found %d conflicts:\n", len(conflicts))
		for _, conflict := range conflicts {
			fmt.Printf("  - %s: both local and remote modified\n", conflict)
		}

		switch strategy {
		case "local":
			fmt.Println("🔧 Resolving conflicts: local wins")
			return performPushSync(syncService, journal, syncState)
		case "remote":
			fmt.Println("🔧 Resolving conflicts: remote wins")
			return performPullSync(syncService, journal, syncState)
		default:
			return fmt.Errorf("unknown conflict strategy: %s", strategy)
		}
	}

	fmt.Println("⬆️  Pushing local changes...")
	if err := syncService.SyncToNotion(journal, syncState); err != nil {
		return err
	}

	fmt.Println("⬇️  Pulling remote changes...")
	if err := syncService.SyncFromNotion(journal, syncState); err != nil {
		return err
	}

	return writer.WriteFile("hfl.md", journal)
}

func detectConflicts(syncService *notion.SyncService, journal *parser.Journal, syncState *state.State) ([]string, error) {
	// For now, return empty conflicts - implement proper conflict detection later
	return []string{}, nil
}

func showSyncPlan(journal *parser.Journal, syncState *state.State) {
	fmt.Printf("📊 Sync Plan for %d local entries:\n\n", len(journal.Entries))

	newCount := 0
	modifiedCount := 0
	syncedCount := 0

	for _, entry := range journal.Entries {
		entryState, exists := syncState.GetEntry(entry.Date)
		wordCount := len(strings.Fields(entry.Body))

		if !exists || entryState.NotionID == "" {
			fmt.Printf("  📝 %s: NEW (%d words, will create in Notion)\n", entry.Date, wordCount)
			newCount++
		} else if syncState.HasChanged(entry.Date, entry.Body) {
			fmt.Printf("  ✏️  %s: MODIFIED (%d words, will update in Notion)\n", entry.Date, wordCount)
			modifiedCount++
		} else {
			fmt.Printf("  ✅ %s: SYNCED (%d words, no changes)\n", entry.Date, wordCount)
			syncedCount++
		}
	}

	fmt.Printf("\nSummary: %d new, %d modified, %d synced\n", newCount, modifiedCount, syncedCount)
}

func init() {
	syncCmd.Flags().BoolVar(&pushOnly, "push", false, "Only push local changes to Notion")
	syncCmd.Flags().BoolVar(&pullOnly, "pull", false, "Only pull changes from Notion")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without making changes")

	RootCmd.AddCommand(syncCmd)
}
