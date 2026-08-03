package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	kizamicontext "github.com/mskasa/kizami/internal/context"
	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Commands for use in git hooks and AI agent tool hooks",
}

var hookPreCommitCmd = &cobra.Command{
	Use:   "pre-commit",
	Short: "Run pre-commit checks (called by the pre-commit git hook)",
	Args:  cobra.NoArgs,
	RunE:  runHookPreCommit,
}

var hookPreToolUseCmd = &cobra.Command{
	Use:   "pre-tool-use",
	Short: "Inject governing decisions into Claude Code's context (called by the PreToolUse hook)",
	Long: `Reads a Claude Code PreToolUse hook event from stdin and, if the file being edited is
governed by any decision, prints hookSpecificOutput.additionalContext naming them — without
blocking the edit. See docs/decisions/2026-08-03-pre-tool-use-hook-context-injection.md.

Configure in .claude/settings.json:

  {
    "hooks": {
      "PreToolUse": [
        { "matcher": "Edit|Write", "hooks": [{ "type": "command", "command": "kizami hook pre-tool-use" }] }
      ]
    }
  }`,
	Args: cobra.NoArgs,
	RunE: runHookPreToolUse,
}

func init() {
	hookCmd.AddCommand(hookPreCommitCmd)
	hookCmd.AddCommand(hookPreToolUseCmd)
	rootCmd.AddCommand(hookCmd)
}

// preToolUseEvent holds the subset of Claude Code's PreToolUse hook JSON this command needs.
// Unknown fields (session_id, transcript_path, tool_use_id, etc.) are ignored.
type preToolUseEvent struct {
	Cwd       string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

type preToolUseResponse struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

// runHookPreToolUse never blocks the tool call and never errors out to the caller — a
// broken or unexpected event, a missing git root, or a config-load failure all just result
// in no output, mirroring runHookPreCommit's existing silent-skip convention. This hook's
// only job is to add context; it must never be the reason an edit fails.
func runHookPreToolUse(cmd *cobra.Command, _ []string) error {
	var event preToolUseEvent
	if err := json.NewDecoder(cmd.InOrStdin()).Decode(&event); err != nil {
		return nil
	}
	if event.ToolInput.FilePath == "" {
		return nil
	}

	root, err := gitRepoRootFn()
	if err != nil {
		return nil
	}
	relPath := hookRelPath(root, event.Cwd, event.ToolInput.FilePath)
	if relPath == "" {
		return nil
	}

	cfg := loadCfg()
	dirs := documentDirs(root, cfg)
	result, err := kizamicontext.Resolve(dirs, root, []string{relPath}, false)
	if err != nil || len(result.Decisions) == 0 {
		return nil
	}

	additionalContext := renderPreToolUseContext(result)
	resp := preToolUseResponse{HookSpecificOutput: hookSpecificOutput{
		HookEventName:     "PreToolUse",
		AdditionalContext: additionalContext,
	}}
	enc := json.NewEncoder(cmd.OutOrStdout())
	if err := enc.Encode(resp); err != nil {
		return nil
	}
	return nil
}

// hookRelPath resolves filePath (from the hook event, possibly relative to cwd) to a
// repo-relative, slash-separated path. Returns "" if it can't be resolved under root.
func hookRelPath(root, cwd, filePath string) string {
	if !filepath.IsAbs(filePath) {
		base := cwd
		if base == "" {
			base = root
		}
		filePath = filepath.Join(base, filePath)
	}
	rel, err := filepath.Rel(root, filePath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ""
	}
	return filepath.ToSlash(rel)
}

// renderPreToolUseContext renders one line per governing decision — slug, title, and a
// drift flag if any — deliberately without the decision body text. This fires on every
// Edit/Write of a governed file, far more often than the agent manifest or an MCP tool call,
// so even the manifest's summary-only default would be too expensive here:
// docs/decisions/2026-08-03-pre-tool-use-hook-context-injection.md
func renderPreToolUseContext(result *kizamicontext.Result) string {
	var sb strings.Builder
	sb.WriteString("This file is governed by recorded decisions:\n")
	for _, d := range result.Decisions {
		sb.WriteString("- [" + d.Slug + "] " + d.Title)
		if d.Drift.State == "drift" {
			sb.WriteString(" (drift: a Related Files entry no longer exists)")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("Run `kizami show <slug>` for details before proceeding.\n")
	return sb.String()
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
