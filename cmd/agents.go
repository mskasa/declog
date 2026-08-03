package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	kizamicontext "github.com/mskasa/kizami/internal/context"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Commands for syncing decisions into agent-read files",
}

var agentsSyncCheck bool

var agentsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a decisions pointer table into CLAUDE.md / AGENTS.md",
	Long: `Write a "path -> decision" pointer table into a marker-delimited section of
CLAUDE.md and/or AGENTS.md (whichever exist at the repo root, or the files configured
via kizami.toml's [agents] targets), so agents see governing decisions without needing
to search for them.

Use --check in CI to fail the build when a target's block is missing or stale.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := documentDirs(root, cfg)
		targets := agentsTargets(root, cfg)

		entries, err := kizamicontext.Manifest(dirs)
		if err != nil {
			return err
		}
		block := kizamicontext.RenderManifest(entries)

		if agentsSyncCheck {
			return runAgentsSyncCheck(root, targets, block)
		}
		return runAgentsSync(root, targets, block)
	},
}

func runAgentsSync(root string, targets []string, block string) error {
	var existing int
	for _, target := range targets {
		content, err := os.ReadFile(target)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading %s: %w", target, err)
		}
		existing++

		updated := kizamicontext.SyncBlock(string(content), block)
		rel, _ := filepath.Rel(root, target)
		if updated == string(content) {
			fmt.Fprintf(os.Stdout, "%s already up to date.\n", rel)
			continue
		}
		if err := os.WriteFile(target, []byte(updated), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", target, err)
		}
		fmt.Fprintf(os.Stdout, "Updated %s.\n", rel)
	}

	if existing == 0 {
		return fmt.Errorf("no target file found (looked for: %v) — create one first, kizami won't create it for you", targets)
	}
	return nil
}

func runAgentsSyncCheck(root string, targets []string, block string) error {
	var stale []string
	for _, target := range targets {
		content, err := os.ReadFile(target)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading %s: %w", target, err)
		}
		if !kizamicontext.HasCurrentBlock(string(content), block) {
			rel, _ := filepath.Rel(root, target)
			stale = append(stale, rel)
		}
	}

	if len(stale) == 0 {
		fmt.Fprintln(os.Stdout, "Agent manifest is up to date. ✅")
		return nil
	}

	fmt.Fprintln(os.Stdout, "Agent manifest is out of date in:")
	for _, f := range stale {
		fmt.Fprintf(os.Stdout, "  %s\n", f)
	}
	fmt.Fprintln(os.Stdout, "\nRun `kizami agents sync` and commit the result.")
	return fmt.Errorf("agent manifest out of date in %d file(s)", len(stale))
}

func init() {
	agentsSyncCmd.Flags().BoolVar(&agentsSyncCheck, "check", false, "Fail if any target's manifest block is missing or stale, without writing")
	agentsCmd.AddCommand(agentsSyncCmd)
	rootCmd.AddCommand(agentsCmd)
}
