package cmd

import (
	"os"
	"path/filepath"

	"github.com/mskasa/declog/internal/config"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "why",
	Version: Version,
	Short:   "Record and search architectural design decisions",
	Long: `why is a CLI tool to record and search architectural design decisions (ADRs).

Decisions are saved as Markdown files under docs/decisions/ and managed with Git.`,
}

// loadCfg loads the user config, returning an empty Config on error.
func loadCfg() *config.Config {
	cfg, _ := config.Load()
	if cfg == nil {
		return &config.Config{}
	}
	return cfg
}

// decisionsDir returns the decisions directory path.
// Uses cfg.Decisions.Dir if set, otherwise defaults to "docs/decisions".
func decisionsDir(root string, cfg *config.Config) string {
	if cfg != nil && cfg.Decisions.Dir != "" {
		return filepath.Join(root, cfg.Decisions.Dir)
	}
	return filepath.Join(root, "docs", "decisions")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
