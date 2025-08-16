package writer

import (
	"github.com/ahmaruff/hfl/internal/parser"
	"os"
	"testing"
)

func TestWriteFile_CanonicalFormat(t *testing.T) {
	// Create test journal
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-15", Body: "Older entry"},
			{Date: "2025-08-16", Body: "Newer entry\n\nWith blank line"},
		},
	}

	filename := "test_output.md"
	defer os.Remove(filename)

	err := WriteFile(filename, journal)
	if err != nil {
		t.Fatal(err)
	}

	// Read back and check format
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	expected := `# 2025-08-16
Newer entry

With blank line

# 2025-08-15
Older entry
`

	if string(content) != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, string(content))
	}
}

func TestWriteFile_EmptyJournal(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{},
	}

	filename := "test_empty.md"
	defer os.Remove(filename)

	err := WriteFile(filename, journal)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "" {
		t.Errorf("Expected empty file, got %q", string(content))
	}
}

func TestWriteFile_SingleEntry(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-16", Body: "Single entry"},
		},
	}

	filename := "test_single.md"
	defer os.Remove(filename)

	err := WriteFile(filename, journal)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	expected := `# 2025-08-16
Single entry
`

	if string(content) != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, string(content))
	}
}
