package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type EntryState struct {
	NotionID       string `json:"notion_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
	LastRemoteEdit string `json:"last_remote_edit,omitempty"`
	LastLocalSync  string `json:"last_local_sync"`
}

type State struct {
	LastSynced string                `json:"last_synced,omitempty"`
	Entries    map[string]EntryState `json:"entries"`
}

func Load() (*State, error) {
	statePath := ".hfl/state.json"

	// Return empty state if file doesn't exist
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &State{
			Entries: make(map[string]EntryState),
		}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	// Ensure entries map is initialized
	if state.Entries == nil {
		state.Entries = make(map[string]EntryState)
	}

	return &state, nil
}

func (s *State) Save() error {
	statePath := ".hfl/state.json"

	// Ensure .hfl directory exists
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (s *State) UpdateEntry(date, body string) {
	hash := calculateHash(body)
	now := time.Now().Format(time.RFC3339)

	entry := s.Entries[date]
	entry.Hash = hash
	entry.LastLocalSync = now

	s.Entries[date] = entry
}

func (s *State) GetEntry(date string) (EntryState, bool) {
	entry, exists := s.Entries[date]
	return entry, exists
}

func (s *State) HasChanged(date, body string) bool {
	entry, exists := s.Entries[date]
	if !exists {
		return true // New entry
	}

	currentHash := calculateHash(body)
	return entry.Hash != currentHash
}

func calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (s *State) SetLastSynced(timestamp string) {
	s.LastSynced = timestamp
}

func (s *State) SetNotionID(date, notionID string) {
	entry := s.Entries[date]
	entry.NotionID = notionID
	s.Entries[date] = entry
}
