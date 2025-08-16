package writer

import (
	"fmt"
	"github.com/ahmaruff/hfl/internal/parser"
	"os"
	"sort"
)

func WriteFile(filename string, journal *parser.Journal) error {
	entries := make([]parser.Entry, len(journal.Entries))
	copy(entries, journal.Entries)

	// Sort by date descending (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date > entries[j].Date
	})

	var journalStr string
	for i, entry := range entries {
		journalStr += "# " + entry.Date + "\n"
		journalStr += entry.Body

		// Add spacing between entries (but not after the last one)
		if i < len(entries)-1 {
			journalStr += "\n\n"
		} else {
			journalStr += "\n"
		}
	}

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}

	defer file.Close()

	_, err = file.WriteString(journalStr)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	return nil
}
