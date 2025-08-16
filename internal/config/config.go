package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type NotionConfig struct {
	ApiToken   string `json:"api_token,omitempty"`
	DatabaseID string `json:"database_id,omitempty"`
}

type Config struct {
	Editor           string       `json:"editor,omitempty"`
	ConflictStrategy string       `json:"conflict_strategy,omitempty"`
	Notion           NotionConfig `json:"notion,omitempty"`
}

func Load() (*Config, error) {
	config := &Config{
		ConflictStrategy: "remote", // Default from spec
	}

	// 1. Load global config first
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	globalPath := filepath.Join(homeDir, ".hfl", "config.json")

	globalConfig, err := LoadFile(globalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Merge global config into base config
	mergeConfig(config, globalConfig)

	// 2. Load and merge local config
	localConfig, err := LoadFile("./.hfl/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load local config: %w", err)
	}

	// Merge local config (overrides global)
	mergeConfig(config, localConfig)

	// 3. Apply environment variable overrides
	applyEnvOverrides(config)

	return config, nil
}

func LoadFile(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return empty config if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return cfg, nil
}

func Save(c *Config, global bool) error {
	var configPath string
	if global {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".hfl", "config.json")
	} else {
		configPath = ".hfl/config.json"
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) Set(key, value string) error {
	switch key {
	case "editor":
		c.Editor = value
	case "conflict_strategy":
		if value != "remote" && value != "local" && value != "merge" {
			return fmt.Errorf("invalid conflict strategy: %s (must be remote, local, or merge)", value)
		}
		c.ConflictStrategy = value
	case "notion.api_token":
		c.Notion.ApiToken = value
	case "notion.database_id":
		c.Notion.DatabaseID = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func (c *Config) Get(key string) (string, error) {
	switch key {
	case "editor":
		return c.GetEditor(), nil
	case "conflict_strategy":
		strategy := c.ConflictStrategy
		if strategy == "" {
			strategy = "remote" // Default
		}
		return strategy, nil
	case "notion.api_token":
		return c.Notion.ApiToken, nil
	case "notion.database_id":
		return c.Notion.DatabaseID, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func (c *Config) GetEditor() string {
	if c.Editor != "" {
		return c.Editor
	}
	if editor := os.Getenv("HFL_EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// OS-aware default
	if runtime.GOOS == "windows" {
		return "notepad"
	}

	return "vi"
}

// mergeConfig merges source config into target config (source overrides target)
func mergeConfig(target, source *Config) {
	if source.Editor != "" {
		target.Editor = source.Editor
	}
	if source.ConflictStrategy != "" {
		target.ConflictStrategy = source.ConflictStrategy
	}
	if source.Notion.ApiToken != "" {
		target.Notion.ApiToken = source.Notion.ApiToken
	}
	if source.Notion.DatabaseID != "" {
		target.Notion.DatabaseID = source.Notion.DatabaseID
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(config *Config) {
	if editor := os.Getenv("HFL_EDITOR"); editor != "" {
		config.Editor = editor
	}
	if token := os.Getenv("HFL_NOTION_TOKEN"); token != "" {
		config.Notion.ApiToken = token
	}
	if dbID := os.Getenv("HFL_NOTION_DATABASE"); dbID != "" {
		config.Notion.DatabaseID = dbID
	}
}
