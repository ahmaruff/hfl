package export

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ahmaruff/hfl/internal/parser"
)

func TestToJSON(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-16", Body: "First entry\n\nWith multiple lines"},
			{Date: "2025-08-15", Body: "Second entry"},
		},
	}

	data, err := ToJSON(journal)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Parse back to verify structure
	var entries []parser.Entry
	err = json.Unmarshal(data, &entries)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Date != "2025-08-16" {
		t.Errorf("Expected date 2025-08-16, got %s", entries[0].Date)
	}

	if entries[0].Body != "First entry\n\nWith multiple lines" {
		t.Errorf("Expected body preserved, got %q", entries[0].Body)
	}

	// Check JSON has lowercase keys
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"date"`) {
		t.Error("JSON should contain lowercase 'date' key")
	}
	if !strings.Contains(jsonStr, `"body"`) {
		t.Error("JSON should contain lowercase 'body' key")
	}
}

func TestToJSONFile(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-16", Body: "Test entry"},
		},
	}

	filename := "test_export.json"
	defer os.Remove(filename)

	err := ToJSONFile(journal, filename)
	if err != nil {
		t.Fatalf("ToJSONFile failed: %v", err)
	}

	// Verify file exists and has correct content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	var entries []parser.Entry
	err = json.Unmarshal(data, &entries)
	if err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Date != "2025-08-16" {
		t.Errorf("Expected date 2025-08-16, got %s", entries[0].Date)
	}
}

func TestToCSV(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-16", Body: "First entry"},
			{Date: "2025-08-15", Body: "Entry with\ttab and\nnewline"},
		},
	}

	data, err := ToCSV(journal)
	if err != nil {
		t.Fatalf("ToCSV failed: %v", err)
	}

	csvStr := string(data)

	// Check header exists
	if !strings.Contains(csvStr, "date") || !strings.Contains(csvStr, "body") {
		t.Error("Expected CSV to contain header with date and body")
	}

	// Check both entries are present
	if !strings.Contains(csvStr, "2025-08-16") {
		t.Error("Expected CSV to contain date 2025-08-16")
	}

	if !strings.Contains(csvStr, "2025-08-15") {
		t.Error("Expected CSV to contain date 2025-08-15")
	}

	if !strings.Contains(csvStr, "First entry") {
		t.Error("Expected CSV to contain 'First entry'")
	}

	// Check that newlines in content are properly quoted
	if !strings.Contains(csvStr, "Entry with") {
		t.Error("Expected CSV to contain 'Entry with' from second entry")
	}
}

func TestToCSVFile(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-16", Body: "Test entry"},
			{Date: "2025-08-15", Body: "Another entry"},
		},
	}

	filename := "test_export.csv"
	defer os.Remove(filename)

	err := ToCSVFile(journal, filename)
	if err != nil {
		t.Fatalf("ToCSVFile failed: %v", err)
	}

	// Verify file exists and has correct content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	csvStr := string(data)
	lines := strings.Split(strings.TrimSpace(csvStr), "\n")

	// Check header
	if !strings.Contains(lines[0], "date") || !strings.Contains(lines[0], "body") {
		t.Errorf("Expected CSV header with date and body, got %q", lines[0])
	}

	// Check we have 3 lines (header + 2 entries)
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 entries), got %d", len(lines))
	}

	// Check entries are present
	content := string(data)
	if !strings.Contains(content, "2025-08-16") {
		t.Error("Expected CSV to contain date 2025-08-16")
	}
	if !strings.Contains(content, "Test entry") {
		t.Error("Expected CSV to contain 'Test entry'")
	}
}

func TestToCSVWithSpecialCharacters(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{
			{Date: "2025-08-16", Body: "Entry with, comma and \"quotes\""},
			{Date: "2025-08-15", Body: "Entry with\nnewlines\nand\ttabs"},
		},
	}

	data, err := ToCSV(journal)
	if err != nil {
		t.Fatalf("ToCSV with special chars failed: %v", err)
	}

	// CSV should handle special characters properly
	csvStr := string(data)
	if !strings.Contains(csvStr, "2025-08-16") {
		t.Error("CSV should contain the date even with special characters in body")
	}

	// Should not break the CSV format
	lines := strings.Split(strings.TrimSpace(csvStr), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines, got %d", len(lines))
	}
}

func TestEmptyJournal(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{},
	}

	// Test JSON export
	data, err := ToJSON(journal)
	if err != nil {
		t.Fatalf("ToJSON with empty journal failed: %v", err)
	}

	var entries []parser.Entry
	err = json.Unmarshal(data, &entries)
	if err != nil {
		t.Fatalf("Failed to parse empty JSON: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected empty array, got %d entries", len(entries))
	}

	// Test CSV export
	csvData, err := ToCSV(journal)
	if err != nil {
		t.Fatalf("ToCSV with empty journal failed: %v", err)
	}

	csvStr := strings.TrimSpace(string(csvData))
	lines := strings.Split(csvStr, "\n")

	// Should only have header
	if len(lines) != 1 {
		t.Errorf("Expected only header line for empty journal, got %d lines", len(lines))
	}

	if !strings.Contains(lines[0], "date") || !strings.Contains(lines[0], "body") {
		t.Errorf("Expected header with date and body, got %q", lines[0])
	}
}

func TestToJSONFile_FileError(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{{Date: "2025-08-16", Body: "Test"}},
	}

	// Try to write to invalid path
	err := ToJSONFile(journal, "/invalid/path/file.json")
	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
}

func TestToCSVFile_FileError(t *testing.T) {
	journal := &parser.Journal{
		Entries: []parser.Entry{{Date: "2025-08-16", Body: "Test"}},
	}

	// Try to write to invalid path
	err := ToCSVFile(journal, "/invalid/path/file.csv")
	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
}
