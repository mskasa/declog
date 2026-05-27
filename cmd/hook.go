package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Commands for use in git hooks",
}

var hookPreCommitCmd = &cobra.Command{
	Use:   "pre-commit",
	Short: "Run pre-commit checks (called by the pre-commit git hook)",
	Args:  cobra.NoArgs,
	RunE:  runHookPreCommit,
}

func init() {
	hookCmd.AddCommand(hookPreCommitCmd)
	rootCmd.AddCommand(hookCmd)
}

// stagedFilesFn is a variable to allow injection in tests.
var stagedFilesFn = func(root string) ([]string, error) {
	// -c core.quotepath=false prevents git from quoting non-ASCII filenames
	// (e.g. Japanese), which would break suffix and prefix matching downstream.
	out, err := exec.Command("git", "-C", root, "-c", "core.quotepath=false", "diff", "--cached", "--name-only").Output()
	if err != nil {
		return nil, fmt.Errorf("getting staged files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func runHookPreCommit(_ *cobra.Command, _ []string) error {
	root, err := gitRepoRootFn()
	if err != nil {
		return nil // not in a git repo; silently skip
	}
	cfg := loadCfg()
	dirs := auditDirs(root, cfg)

	staged, err := stagedFilesFn(root)
	if err != nil || len(staged) == 0 {
		return nil
	}

	// Check 1: Active documents whose Related Files overlap with staged files,
	// but which are not themselves staged — remind to review or update them.
	results, err := decision.CheckHook(dirs, root, staged)
	if err != nil {
		return nil // silently skip on error to avoid blocking commits
	}
	if len(results) > 0 {
		fmt.Fprintln(os.Stdout, "\n⚠️  These documents reference files you're changing but are not staged:")
		for _, r := range results {
			rel, _ := filepath.Rel(root, r.File)
			fmt.Fprintf(os.Stdout, "  [%s] %s\n", r.Slug, filepath.ToSlash(rel))
		}
		fmt.Fprintln(os.Stdout, "\n  Consider reviewing and staging them, or run kizami status to update their status.")
		fmt.Fprintln(os.Stdout)
		return nil
	}

	// Check 2: No document staged + non-doc files changed → suggest creating a new document.
	if !anyDocStaged(dirs, root, staged) && hasNonDocFiles(staged) {
		fmt.Fprintln(os.Stdout, "\n⚠️  No document found in this commit.")
		fmt.Fprintln(os.Stdout, "    If this change involves a significant design decision,")
		fmt.Fprintln(os.Stdout, "    consider running: kizami adr \"<title>\"")
		fmt.Fprintln(os.Stdout)
	}

	return nil
}

// anyDocStaged reports whether any staged file is under one of the document dirs.
func anyDocStaged(dirs []string, root string, staged []string) bool {
	for _, dir := range dirs {
		relDir, err := filepath.Rel(root, dir)
		if err != nil {
			continue
		}
		prefix := filepath.ToSlash(relDir) + "/"
		for _, f := range staged {
			if strings.HasPrefix(filepath.ToSlash(f), prefix) {
				return true
			}
		}
	}
	return false
}

// hasNonDocFiles reports whether staged contains any non-Markdown, non-sidecar file.
func hasNonDocFiles(staged []string) bool {
	for _, f := range staged {
		if !strings.HasSuffix(f, ".md") && !strings.HasSuffix(f, ".kizami") {
			return true
		}
	}
	return false
}
