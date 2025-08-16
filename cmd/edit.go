package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ahmaruff/hfl/internal/config"
	"github.com/ahmaruff/hfl/internal/parser"
	"github.com/ahmaruff/hfl/internal/state"
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
		parsedDate, err := parseDate(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		date = parsedDate
	} else {
		date = time.Now().Format("2006-01-02") // Today
	}

	// Ensure hfl.md exists and has the entry
	ensureEntryExists(date)

	// Open editor
	openEditor("hfl.md", date)

	// Post-edit validation
	validateAfterEdit()
}

func validateDate(date string) error {
	// Check format with regex
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !re.MatchString(date) {
		return fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", date)
	}

	// Parse to check if it's a valid date
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date: %s", date)
	}

	return nil
}

func validateAfterEdit() {
	fmt.Println("Validating changes...")
	journal, warnings, err := parser.ParseFile("hfl.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		return
	}

	if len(warnings) > 0 {
		fmt.Println("Warnings found:")
		for _, warning := range warnings {
			fmt.Println("  " + warning)
		}
		fmt.Println()
	}

	// Auto-format
	fmt.Println("Formatting to canonical style...")
	err = writer.WriteFile("hfl.md", journal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting file: %v\n", err)
		return
	}

	fmt.Printf("File is valid. Found %d entries, formatted successfully.\n", len(journal.Entries))
}

func updateStateAfterEdit(journal *parser.Journal) {
	state, err := state.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load state: %v\n", err)
		return
	}

	// Update state for all entries
	for _, entry := range journal.Entries {
		state.UpdateEntry(entry.Date, entry.Body)
	}

	// Save state
	if err := state.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
	}
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

func parseDate(dateStr string) (string, error) {
	dateStr = strings.TrimSpace(strings.ToLower(dateStr))
	now := time.Now()

	switch dateStr {
	case "today":
		return now.Format("2006-01-02"), nil
	case "yesterday":
		return now.AddDate(0, 0, -1).Format("2006-01-02"), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format("2006-01-02"), nil
	}

	// Handle n+X or n-X format
	if strings.HasPrefix(dateStr, "n+") || strings.HasPrefix(dateStr, "n-") {
		offsetStr := dateStr[1:] // Remove 'n', keep +X or -X
		days, err := strconv.Atoi(offsetStr)
		if err == nil {
			return now.AddDate(0, 0, days).Format("2006-01-02"), nil
		}
	}

	// Handle weekdays: monday, tuesday, etc.
	if weekday := parseWeekday(dateStr); weekday >= 0 {
		daysUntil := int(weekday - now.Weekday())
		if daysUntil <= 0 {
			daysUntil += 7 // Next week
		}
		return now.AddDate(0, 0, daysUntil).Format("2006-01-02"), nil
	}

	// Try parsing as YYYY-MM-DD
	if err := validateDate(dateStr); err == nil {
		return dateStr, nil
	}

	return "", fmt.Errorf("unable to parse date: %s", dateStr)
}

func parseWeekday(day string) time.Weekday {
	switch day {
	case "sunday", "sun":
		return time.Sunday
	case "monday", "mon":
		return time.Monday
	case "tuesday", "tue", "tues":
		return time.Tuesday
	case "wednesday", "wed":
		return time.Wednesday
	case "thursday", "thu", "thurs":
		return time.Thursday
	case "friday", "fri":
		return time.Friday
	case "saturday", "sat":
		return time.Saturday
	default:
		return -1
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
