package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mskasa/kizami/internal/config"
	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "kizami",
	Version: Version,
	Short:   "Maintain living documentation alongside code, with automatic drift detection",
	Long: `kizami is a CLI tool to maintain living documentation (ADRs, design docs, API specs)
alongside your code, with automatic detection of document-code drift.

Documents are saved as Markdown files and managed with Git.`,
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

// designDir returns the design documents directory path.
func designDir(root string) string {
	return filepath.Join(root, "docs", "design")
}

// documentDirs returns the list of directories for read/write commands (list, search, show, etc.).
// Uses cfg.Documents.Dirs if set, otherwise falls back to decisionsDir.
func documentDirs(root string, cfg *config.Config) []string {
	if cfg != nil && len(cfg.Documents.Dirs) > 0 {
		dirs := make([]string, len(cfg.Documents.Dirs))
		for i, d := range cfg.Documents.Dirs {
			dirs[i] = filepath.Join(root, d)
		}
		return dirs
	}
	return []string{decisionsDir(root, cfg)}
}

// auditDirs returns the list of directories to audit.
// Uses cfg.Audit.Dirs if set, otherwise falls back to documentDirs.
func auditDirs(root string, cfg *config.Config) []string {
	if cfg != nil && len(cfg.Audit.Dirs) > 0 {
		dirs := make([]string, len(cfg.Audit.Dirs))
		for i, d := range cfg.Audit.Dirs {
			dirs[i] = filepath.Join(root, d)
		}
		return dirs
	}
	return documentDirs(root, cfg)
}

// findByID searches for a decision by ID across all document directories.
func findByID(root string, cfg *config.Config, id int) (*decision.Decision, error) {
	for _, dir := range documentDirs(root, cfg) {
		d, err := decision.FindByID(dir, id)
		if err == nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("decision %04d not found", id)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
