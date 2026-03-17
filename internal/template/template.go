package template

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ChangedFiles returns a deduplicated list of staged and unstaged changed files
// by running git in dir. Returns nil if not in a git repository or no files are changed.
func ChangedFiles(dir string) []string {
	staged := gitDiffFiles(dir, "--staged")
	unstaged := gitDiffFiles(dir, "")

	seen := make(map[string]struct{}, len(staged)+len(unstaged))
	var result []string
	for _, f := range append(staged, unstaged...) {
		if _, ok := seen[f]; !ok {
			seen[f] = struct{}{}
			result = append(result, f)
		}
	}
	return result
}

func gitDiffFiles(dir, flag string) []string {
	args := []string{"diff", "--name-only"}
	if flag != "" {
		args = []string{"diff", flag, "--name-only"}
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

// RenderHeader returns the ADR front-matter block (title + metadata lines only, no body sections).
// It is used when the body is generated externally (e.g. by an LLM).
func RenderHeader(id int, title, author string, supersededBy int) string {
	date := time.Now().Format("2006-01-02")
	supersedes := ""
	if supersededBy > 0 {
		supersedes = fmt.Sprintf("\n- Supersedes: %04d", supersededBy)
	}
	return fmt.Sprintf("# %04d: %s\n\n- Date: %s\n- Status: Active\n- Author: %s%s\n",
		id, title, date, author, supersedes)
}

// Render returns the MADR Markdown template filled with the given values.
// relatedFiles is inserted into the Related Files section; pass nil for an empty section.
// supersededBy, if > 0, adds a "- Supersedes: NNNN" line after the Author line.
func Render(id int, title, author string, relatedFiles []string, supersededBy int) string {
	date := time.Now().Format("2006-01-02")
	supersedes := ""
	if supersededBy > 0 {
		supersedes = fmt.Sprintf("\n- Supersedes: %04d", supersededBy)
	}
	return fmt.Sprintf(`# %04d: %s

- Date: %s
- Status: Active
- Author: %s`+supersedes+`

## Context

<!-- Why this decision was needed. Describe the background, constraints, and problem. -->

## Decision

<!-- What was decided. State clearly in 1–3 sentences. -->

## Consequences

<!-- Impact, benefits, and trade-offs of this decision. -->

## Alternatives Considered

<!-- Options that were considered but not adopted, and why. (Optional) -->

## Related Files

%s`, id, title, date, author, renderRelatedFiles(relatedFiles))
}

func renderRelatedFiles(files []string) string {
	if len(files) == 0 {
		return "<!-- List files related to this decision (e.g. internal/search/search.go). -->\n"
	}
	var sb strings.Builder
	for _, f := range files {
		sb.WriteString("- `")
		sb.WriteString(f)
		sb.WriteString("`\n")
	}
	return sb.String()
}
