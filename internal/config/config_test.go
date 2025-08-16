package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoad_NoConfigFiles(t *testing.T) {
	// Test in temporary directory with no config files
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Clear environment variables
	os.Unsetenv("HFL_EDITOR")
	os.Unsetenv("HFL_NOTION_TOKEN")
	os.Unsetenv("HFL_NOTION_DATABASE")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() should not fail when no config files exist: %v", err)
	}

	// Should have default values
	if config.ConflictStrategy != "remote" {
		t.Errorf("Expected default conflict strategy 'remote', got %q", config.ConflictStrategy)
	}

	if config.Editor != "" {
		t.Errorf("Expected empty editor with no config, got %q", config.Editor)
	}
}

func TestLoad_LocalConfigOnly(t *testing.T) {
	// Test with only local config
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create .hfl directory and config
	err := os.MkdirAll(".hfl", 0755)
	if err != nil {
		t.Fatal(err)
	}

	localConfig := Config{
		Editor:           "vim",
		ConflictStrategy: "local",
		Notion: NotionConfig{
			ApiToken:   "local-token",
			DatabaseID: "local-db",
		},
	}

	data, _ := json.Marshal(localConfig)
	err = os.WriteFile(".hfl/config.json", data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.Editor != "vim" {
		t.Errorf("Expected editor 'vim', got %q", config.Editor)
	}

	if config.ConflictStrategy != "local" {
		t.Errorf("Expected conflict strategy 'local', got %q", config.ConflictStrategy)
	}

	if config.Notion.ApiToken != "local-token" {
		t.Errorf("Expected notion token 'local-token', got %q", config.Notion.ApiToken)
	}
}

func TestLoad_GlobalConfigOnly(t *testing.T) {
	// Test with only global config
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create fake global config in temp directory
	globalDir := filepath.Join(tempDir, "fake-home", ".hfl")
	err := os.MkdirAll(globalDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	globalConfig := Config{
		Editor:           "emacs",
		ConflictStrategy: "merge",
		Notion: NotionConfig{
			ApiToken:   "global-token",
			DatabaseID: "global-db",
		},
	}

	data, _ := json.Marshal(globalConfig)
	err = os.WriteFile(filepath.Join(globalDir, "config.json"), data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Mock os.UserHomeDir() by temporarily setting HOME (Unix) or USERPROFILE (Windows)
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", filepath.Join(tempDir, "fake-home"))
	os.Setenv("USERPROFILE", filepath.Join(tempDir, "fake-home"))

	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalUserProfile != "" {
			os.Setenv("USERPROFILE", originalUserProfile)
		} else {
			os.Unsetenv("USERPROFILE")
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.Editor != "emacs" {
		t.Errorf("Expected editor 'emacs', got %q", config.Editor)
	}

	if config.ConflictStrategy != "merge" {
		t.Errorf("Expected conflict strategy 'merge', got %q", config.ConflictStrategy)
	}
}

func TestLoad_LocalOverridesGlobal(t *testing.T) {
	// Test that local config overrides global
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create fake global config
	globalDir := filepath.Join(tempDir, "fake-home", ".hfl")
	err := os.MkdirAll(globalDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	globalConfig := Config{
		Editor:           "emacs",
		ConflictStrategy: "merge",
		Notion: NotionConfig{
			ApiToken:   "global-token",
			DatabaseID: "global-db",
		},
	}

	data, _ := json.Marshal(globalConfig)
	err = os.WriteFile(filepath.Join(globalDir, "config.json"), data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create local config that partially overrides
	err = os.MkdirAll(".hfl", 0755)
	if err != nil {
		t.Fatal(err)
	}

	localConfig := Config{
		Editor: "vim", // Override editor
		Notion: NotionConfig{
			ApiToken: "local-token", // Override token, keep database from global
		},
	}

	data, _ = json.Marshal(localConfig)
	err = os.WriteFile(".hfl/config.json", data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Mock home directory
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", filepath.Join(tempDir, "fake-home"))
	os.Setenv("USERPROFILE", filepath.Join(tempDir, "fake-home"))

	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalUserProfile != "" {
			os.Setenv("USERPROFILE", originalUserProfile)
		} else {
			os.Unsetenv("USERPROFILE")
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Local should override
	if config.Editor != "vim" {
		t.Errorf("Expected local editor 'vim', got %q", config.Editor)
	}

	if config.Notion.ApiToken != "local-token" {
		t.Errorf("Expected local token 'local-token', got %q", config.Notion.ApiToken)
	}

	// Global should be preserved where not overridden
	if config.ConflictStrategy != "merge" {
		t.Errorf("Expected global conflict strategy 'merge', got %q", config.ConflictStrategy)
	}

	if config.Notion.DatabaseID != "global-db" {
		t.Errorf("Expected global database 'global-db', got %q", config.Notion.DatabaseID)
	}
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	// Test that environment variables override config files
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create local config
	err := os.MkdirAll(".hfl", 0755)
	if err != nil {
		t.Fatal(err)
	}

	localConfig := Config{
		Editor: "vim",
		Notion: NotionConfig{
			ApiToken:   "file-token",
			DatabaseID: "file-db",
		},
	}

	data, _ := json.Marshal(localConfig)
	err = os.WriteFile(".hfl/config.json", data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set environment variables
	os.Setenv("HFL_EDITOR", "code")
	os.Setenv("HFL_NOTION_TOKEN", "env-token")
	os.Setenv("HFL_NOTION_DATABASE", "env-db")
	defer func() {
		os.Unsetenv("HFL_EDITOR")
		os.Unsetenv("HFL_NOTION_TOKEN")
		os.Unsetenv("HFL_NOTION_DATABASE")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Environment should override file
	if config.Editor != "code" {
		t.Errorf("Expected env editor 'code', got %q", config.Editor)
	}

	if config.Notion.ApiToken != "env-token" {
		t.Errorf("Expected env token 'env-token', got %q", config.Notion.ApiToken)
	}

	if config.Notion.DatabaseID != "env-db" {
		t.Errorf("Expected env database 'env-db', got %q", config.Notion.DatabaseID)
	}
}

func TestGetEditor_Precedence(t *testing.T) {
	// Determine expected default editor based on OS
	expectedDefault := "vi"
	if runtime.GOOS == "windows" {
		expectedDefault = "notepad"
	}

	tests := []struct {
		name           string
		configEditor   string
		hflEditor      string
		systemEditor   string
		expectedEditor string
	}{
		{
			name:           "config editor takes precedence",
			configEditor:   "vim",
			hflEditor:      "code",
			systemEditor:   "nano",
			expectedEditor: "vim",
		},
		{
			name:           "HFL_EDITOR when no config",
			configEditor:   "",
			hflEditor:      "code",
			systemEditor:   "nano",
			expectedEditor: "code",
		},
		{
			name:           "EDITOR when no config or HFL_EDITOR",
			configEditor:   "",
			hflEditor:      "",
			systemEditor:   "nano",
			expectedEditor: "nano",
		},
		{
			name:           "default editor when nothing set",
			configEditor:   "",
			hflEditor:      "",
			systemEditor:   "",
			expectedEditor: expectedDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("HFL_EDITOR")
			os.Unsetenv("EDITOR")

			// Set environment variables if specified
			if tt.hflEditor != "" {
				os.Setenv("HFL_EDITOR", tt.hflEditor)
			}
			if tt.systemEditor != "" {
				os.Setenv("EDITOR", tt.systemEditor)
			}

			config := &Config{
				Editor: tt.configEditor,
			}

			result := config.GetEditor()
			if result != tt.expectedEditor {
				t.Errorf("Expected editor %q, got %q", tt.expectedEditor, result)
			}

			// Cleanup
			os.Unsetenv("HFL_EDITOR")
			os.Unsetenv("EDITOR")
		})
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	// Test with invalid JSON in config file
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	err := os.MkdirAll(".hfl", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	err = os.WriteFile(".hfl/config.json", []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Error("Expected error when loading invalid JSON config")
	}
}
