package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"

	"github.com/ahmaruff/hfl/internal/parser"
)

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
		err := writer.Write([]string{entry.Date, entry.Body})
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
