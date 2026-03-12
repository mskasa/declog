package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "why",
	Short: "Record and search architectural design decisions",
	Long: `why is a CLI tool to record and search architectural design decisions (ADRs).

Decisions are saved as Markdown files under docs/decisions/ and managed with Git.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
