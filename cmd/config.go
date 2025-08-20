package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ahmaruff/hfl/internal/config"
	"github.com/ahmaruff/hfl/internal/gitignore"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage HFL configuration",
	Long:  "Get and set configuration values for HFL",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  "Set a configuration value. Use --global to set in global config.",
	Args:  cobra.ExactArgs(2),
	Run:   runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value(s)",
	Long:  "Get a configuration value, or show all config if no key specified.",
	Args:  cobra.MaximumNArgs(1),
	Run:   runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available configuration keys",
	Long:  "Show all available configuration keys and their descriptions.",
	Run:   runConfigList,
}

var globalFlag bool

func runConfigSet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	// Ensure .hfl/ is gitignored (before creating any .hfl files)
	if err := gitignore.EnsureHFLIgnored(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update .gitignore: %v\n", err)
	}

	// Determine which config file to load/modify
	var configPath string
	if globalFlag {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		configPath = filepath.Join(homeDir, ".hfl", "config.json")
	} else {
		configPath = ".hfl/config.json"
	}

	cfg, err := config.LoadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Set(key, value); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := config.Save(cfg, globalFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	scope := "local"
	if globalFlag {
		scope = "global"
	}
	fmt.Printf("Set %s config: %s = %s\n", scope, key, value)
}

func runConfigGet(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	key := args[0]
	value, err := cfg.Get(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s = %s\n", key, value)
}

func runConfigList(cmd *cobra.Command, args []string) {
	fmt.Println("Available configuration keys:")
	fmt.Println()
	fmt.Println("  editor                - Your preferred text editor")
	fmt.Println("  conflict_strategy     - How to handle sync conflicts (remote, local, merge)")
	fmt.Println("  notion.api_token      - Notion API token for sync")
	fmt.Println("  notion.database_id    - Notion database ID for sync")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  hfl config set editor \"code\"")
	fmt.Println("  hfl config set editor \"vim\" --global")
	fmt.Println("  hfl config set notion.api_token \"secret_xyz\"")
	fmt.Println("  hfl config get editor")
	fmt.Println("  hfl config get")
	fmt.Println()
	fmt.Println("Configuration precedence (highest to lowest):")
	fmt.Println("  1. Environment variables (HFL_EDITOR, HFL_NOTION_TOKEN, etc.)")
	fmt.Println("  2. Local config (./.hfl/config.json)")
	fmt.Println("  3. Global config (~/.hfl/config.json)")
	fmt.Println("  4. Built-in defaults")
}

func init() {
	// Add flags
	configSetCmd.Flags().BoolVar(&globalFlag, "global", false, "Set in global config instead of local")

	// Add subcommands
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)

	// Add to root command
	RootCmd.AddCommand(configCmd)
}
