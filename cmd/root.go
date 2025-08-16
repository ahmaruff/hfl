package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:     "hfl",
	Version: "0.2.0",
	Short:   "Homework for Life - A simple journaling tool",
	Long:    "HFL manages your daily journal in a single markdown file with optional Notion sync.",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
