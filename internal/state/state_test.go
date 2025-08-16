package state

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestLoad_NoStateFile(t *testing.T) {
	// Test in temporary directory with no state file
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	state, err := Load()
	if err != nil {
		t.Fatalf("Load() should not fail when no state file exists: %v", err)
	}

	if state == nil {
		t.Fatal("Expected non-nil state")
	}

	if state.Entries == nil {
		t.Fatal("Expected initialized entries map")
	}

	if len(state.Entries) != 0 {
		t.Errorf("Expected empty entries map, got %d entries", len(state.Entries))
	}

	if state.LastSynced != "" {
		t.Errorf("Expected empty last_synced, got %q", state.LastSynced)
	}
}

func TestLoad_ExistingStateFile(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create .hfl directory and state file
	err := os.MkdirAll(".hfl", 0755)
	if err != nil {
		t.Fatal(err)
	}

	testState := State{
		LastSynced: "2025-08-17T10:30:00Z",
		Entries: map[string]EntryState{
			"2025-08-16": {
				NotionID:       "abc123",
				Hash:           "test-hash",
				LastRemoteEdit: "2025-08-16T15:30:00Z",
				LastLocalSync:  "2025-08-16T16:00:00Z",
			},
		},
	}

	data, _ := json.Marshal(testState)
	err = os.WriteFile(".hfl/state.json", data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load the state
	state, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if state.LastSynced != "2025-08-17T10:30:00Z" {
		t.Errorf("Expected last_synced '2025-08-17T10:30:00Z', got %q", state.LastSynced)
	}

	entry, exists := state.Entries["2025-08-16"]
	if !exists {
		t.Fatal("Expected entry for 2025-08-16")
	}

	if entry.NotionID != "abc123" {
		t.Errorf("Expected notion_id 'abc123', got %q", entry.NotionID)
	}

	if entry.Hash != "test-hash" {
		t.Errorf("Expected hash 'test-hash', got %q", entry.Hash)
	}
}

func TestSave(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	state := &State{
		LastSynced: "2025-08-17T10:30:00Z",
		Entries: map[string]EntryState{
			"2025-08-16": {
				NotionID:      "abc123",
				Hash:          "test-hash",
				LastLocalSync: "2025-08-16T16:00:00Z",
			},
		},
	}

	err := state.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(".hfl/state.json"); os.IsNotExist(err) {
		t.Fatal("Expected state.json file to be created")
	}

	// Verify file contents
	data, err := os.ReadFile(".hfl/state.json")
	if err != nil {
		t.Fatal(err)
	}

	var savedState State
	err = json.Unmarshal(data, &savedState)
	if err != nil {
		t.Fatalf("Failed to parse saved state: %v", err)
	}

	if savedState.LastSynced != "2025-08-17T10:30:00Z" {
		t.Errorf("Expected saved last_synced '2025-08-17T10:30:00Z', got %q", savedState.LastSynced)
	}

	if len(savedState.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(savedState.Entries))
	}
}

func TestUpdateEntry(t *testing.T) {
	state := &State{
		Entries: make(map[string]EntryState),
	}

	// Update new entry
	state.UpdateEntry("2025-08-16", "Test content")

	entry, exists := state.Entries["2025-08-16"]
	if !exists {
		t.Fatal("Expected entry to be created")
	}

	if entry.Hash == "" {
		t.Error("Expected hash to be set")
	}

	if entry.LastLocalSync == "" {
		t.Error("Expected last_local_sync to be set")
	}

	// Parse the timestamp to ensure it's valid
	_, err := time.Parse(time.RFC3339, entry.LastLocalSync)
	if err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got %q: %v", entry.LastLocalSync, err)
	}

	// Update existing entry
	originalHash := entry.Hash
	state.UpdateEntry("2025-08-16", "Modified content")

	updatedEntry := state.Entries["2025-08-16"]
	if updatedEntry.Hash == originalHash {
		t.Error("Expected hash to change when content changes")
	}
}

func TestGetEntry(t *testing.T) {
	state := &State{
		Entries: map[string]EntryState{
			"2025-08-16": {
				NotionID: "abc123",
				Hash:     "test-hash",
			},
		},
	}

	// Test existing entry
	entry, exists := state.GetEntry("2025-08-16")
	if !exists {
		t.Error("Expected entry to exist")
	}

	if entry.NotionID != "abc123" {
		t.Errorf("Expected notion_id 'abc123', got %q", entry.NotionID)
	}

	// Test non-existing entry
	_, exists = state.GetEntry("2025-08-15")
	if exists {
		t.Error("Expected entry to not exist")
	}
}

func TestHasChanged(t *testing.T) {
	state := &State{
		Entries: make(map[string]EntryState),
	}

	// Test new entry (should be considered changed)
	if !state.HasChanged("2025-08-16", "New content") {
		t.Error("Expected new entry to be considered changed")
	}

	// Add entry to state
	state.UpdateEntry("2025-08-16", "Original content")

	// Test same content (should not be changed)
	if state.HasChanged("2025-08-16", "Original content") {
		t.Error("Expected same content to not be considered changed")
	}

	// Test different content (should be changed)
	if !state.HasChanged("2025-08-16", "Modified content") {
		t.Error("Expected different content to be considered changed")
	}
}

func TestCalculateHash(t *testing.T) {
	// Test consistent hashing
	content := "Test content"
	hash1 := calculateHash(content)
	hash2 := calculateHash(content)

	if hash1 != hash2 {
		t.Error("Expected consistent hashing for same content")
	}

	// Test different content produces different hashes
	hash3 := calculateHash("Different content")
	if hash1 == hash3 {
		t.Error("Expected different hashes for different content")
	}

	// Test empty content
	emptyHash := calculateHash("")
	if emptyHash == "" {
		t.Error("Expected non-empty hash for empty content")
	}

	// Verify hash format (should be hex string)
	if len(hash1) != 64 { // SHA-256 produces 64 character hex string
		t.Errorf("Expected 64 character hash, got %d characters", len(hash1))
	}
}

func TestSetLastSynced(t *testing.T) {
	state := &State{
		Entries: make(map[string]EntryState),
	}

	timestamp := "2025-08-17T10:30:00Z"
	state.SetLastSynced(timestamp)

	if state.LastSynced != timestamp {
		t.Errorf("Expected last_synced %q, got %q", timestamp, state.LastSynced)
	}
}

func TestSetNotionID(t *testing.T) {
	state := &State{
		Entries: make(map[string]EntryState),
	}

	// Set notion ID for non-existing entry
	state.SetNotionID("2025-08-16", "abc123")

	entry, exists := state.Entries["2025-08-16"]
	if !exists {
		t.Fatal("Expected entry to be created")
	}

	if entry.NotionID != "abc123" {
		t.Errorf("Expected notion_id 'abc123', got %q", entry.NotionID)
	}

	// Update notion ID for existing entry
	state.Entries["2025-08-16"] = EntryState{
		Hash:          "existing-hash",
		LastLocalSync: "2025-08-16T10:00:00Z",
	}

	state.SetNotionID("2025-08-16", "xyz789")

	updatedEntry := state.Entries["2025-08-16"]
	if updatedEntry.NotionID != "xyz789" {
		t.Errorf("Expected updated notion_id 'xyz789', got %q", updatedEntry.NotionID)
	}

	// Verify other fields are preserved
	if updatedEntry.Hash != "existing-hash" {
		t.Error("Expected existing hash to be preserved")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	err := os.MkdirAll(".hfl", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	err = os.WriteFile(".hfl/state.json", []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Error("Expected error when loading invalid JSON state")
	}
}

func TestSave_CreateDirectory(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Don't create .hfl directory manually
	state := &State{
		Entries: make(map[string]EntryState),
	}

	err := state.Save()
	if err != nil {
		t.Fatalf("Save() should create directory automatically: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(".hfl"); os.IsNotExist(err) {
		t.Error("Expected .hfl directory to be created")
	}

	// Verify file was created
	if _, err := os.Stat(".hfl/state.json"); os.IsNotExist(err) {
		t.Error("Expected state.json file to be created")
	}
}
