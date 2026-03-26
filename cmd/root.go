package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/mskasa/kizami/internal/config"
	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
// Falls back to the module version from build info (e.g. when installed via go install).
var Version = "dev"

func init() {
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
	rootCmd.Version = Version
}

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
	root, _ := gitRepoRootFn()
	cfg, _ := config.Load(root)
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
// Uses cfg.Design.Dir if set, otherwise defaults to "docs/design".
func designDir(root string, cfg *config.Config) string {
	if cfg != nil && cfg.Design.Dir != "" {
		return filepath.Join(root, cfg.Design.Dir)
	}
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

// findBySlug searches for a decision by slug across all document directories.
func findBySlug(root string, cfg *config.Config, slug string) (*decision.Decision, error) {
	for _, dir := range documentDirs(root, cfg) {
		d, err := decision.FindBySlug(dir, slug)
		if err == nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("document %q not found", slug)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
