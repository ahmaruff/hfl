package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ahmaruff/hfl/internal/config"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/ahmaruff/hfl/internal/writer"
	"github.com/spf13/cobra"
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
	openEditor("hfl.md", date)
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

func openEditor(filename string, targetDate string) {
	cfg, err := config.Load()
	if err != nil {
		// Fallback if config fails using empty config
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = &config.Config{}
	}

	editor := cfg.GetEditor()

	// Find line number of target entry
	lineNum := findEntryLine(filename, targetDate)

	cmd := buildEditorCommand(editor, filename, lineNum)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
		os.Exit(1)
	}
}

func findEntryLine(filename, targetDate string) int {
	file, err := os.Open(filename)
	if err != nil {
		return 0 // File doesn't exist yet
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.HasPrefix(line, "# "+targetDate) {
			return lineNum + 1 // Position cursor at the body, not the heading
		}
	}

	return 0 // Entry not found
}

func buildEditorCommand(editor, filename string, lineNum int) *exec.Cmd {
	switch {
	case strings.Contains(editor, "code") || strings.Contains(editor, "codium"):
		// VS Code: code --goto file:line:column
		if lineNum > 0 {
			return exec.Command(editor, "--goto", fmt.Sprintf("%s:%d:1", filename, lineNum))
		}
		return exec.Command(editor, filename)

	case strings.Contains(editor, "vim") || strings.Contains(editor, "nvim"):
		// Vim: vim +line file
		if lineNum > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNum), filename)
		}
		return exec.Command(editor, filename)

	case strings.Contains(editor, "nano"):
		// Nano: nano +line file
		if lineNum > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNum), filename)
		}
		return exec.Command(editor, filename)

	case strings.Contains(editor, "emacs"):
		// Emacs: emacs +line file
		if lineNum > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNum), filename)
		}
		return exec.Command(editor, filename)

	case strings.Contains(editor, "subl") || strings.Contains(editor, "sublime"):
		// Sublime Text: subl file:line
		if lineNum > 0 {
			return exec.Command(editor, fmt.Sprintf("%s:%d", filename, lineNum))
		}
		return exec.Command(editor, filename)

	case strings.Contains(editor, "atom"):
		// Atom: atom file:line
		if lineNum > 0 {
			return exec.Command(editor, fmt.Sprintf("%s:%d", filename, lineNum))
		}
		return exec.Command(editor, filename)

	default:
		// Fallback - just open the file
		if runtime.GOOS == "windows" {
			return exec.Command("cmd", "/c", editor, filename)
		}
		return exec.Command("sh", "-c", editor+" "+filename)
	}
}

func init() {
	RootCmd.AddCommand(editCmd)
}
