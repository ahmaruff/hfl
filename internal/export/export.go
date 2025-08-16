package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"

	"github.com/ahmaruff/hfl/internal/parser"
)

func ToJSONFile(journal *parser.Journal, filename string) error {
	data, err := json.MarshalIndent(journal.Entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func ToCSVFile(journal *parser.Journal, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"date", "body"}); err != nil {
		return err
	}

	// Write entries
	for _, entry := range journal.Entries {

		// Replace newlines with spaces for CSV compatibility
		cleanBody := strings.ReplaceAll(entry.Body, "\n", " ")
		err := writer.Write([]string{entry.Date, cleanBody})
		if err != nil {
			return err
		}
	}

	return writer.Error()
}

func ToJSON(journal *parser.Journal) ([]byte, error) {
	return json.MarshalIndent(journal.Entries, "", "  ")
}

// Bisa tambahin format lain nanti
func ToCSV(journal *parser.Journal) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = '\t' // using tab for separator

	// Write header
	err := writer.Write([]string{"date", "body"})
	if err != nil {
		return nil, err
	}

	// Write entries
	for _, entry := range journal.Entries {
		// Replace newlines with spaces for CSV compatibility
		cleanBody := strings.ReplaceAll(entry.Body, "\n", " ")
		err := writer.Write([]string{entry.Date, cleanBody})
		if err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err

	}

	return buf.Bytes(), nil
}
