package parser

import (
	"os"
	"testing"
)

func TestParseFile_ValidEntries(t *testing.T) {
	// Create a test file
	content := `# 2025-08-16
First entry.

Second line.






# 2025-08-15
Just one line.\n\n\n\n\n`

	filename := "test_valid.md"
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	// Parse it
	journal, warnings, err := ParseFile(filename)

	// Check results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %v", warnings)
	}

	if len(journal.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(journal.Entries))
	}

	// Check first entry
	if journal.Entries[0].Date != "2025-08-16" {
		t.Errorf("Expected date 2025-08-16, got %s", journal.Entries[0].Date)
	}

	expectedBody := "First entry.\n\nSecond line."
	if journal.Entries[0].Body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, journal.Entries[0].Body)
	}
}

func TestParseFile_InvalidHeading(t *testing.T) {
	content := `# Aug 16, 2025
Invalid heading format.`

	// TODO: Create test file, parse it, check that you get a warning
	filename := "test_invalid.md"
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	// Parse it
	journal, warnings, err := ParseFile(filename)

	// Check results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(journal.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(journal.Entries))
	}

	if len(warnings) == 0 {
		t.Errorf("Expected warnings, got none")
	}

}

func TestParseFile_DuplicateDate(t *testing.T) {
	// TODO: Test duplicate date detection
	content := `# 2025-08-16
First entry.

Second line.

# 2025-08-16
Just one line.`

	// TODO: Create test file, parse it, check that you get a warning
	filename := "test_duplicate_date.md"
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	// Parse it
	journal, warnings, err := ParseFile(filename)

	// Check results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(warnings) != 1 {
		t.Errorf("Expected 1 warnings, got %v", warnings)
	}

	if len(journal.Entries) != 1 {
		t.Errorf("Expected 1 entries, got %d", len(journal.Entries))
	}

}
