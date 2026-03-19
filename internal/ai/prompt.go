package ai

import (
	"os/exec"
	"strings"
)

const DiffLimit = 2000

// PromptInput holds the context used to build an ADR draft prompt.
type PromptInput struct {
	Title        string
	ChangedFiles []string
	Diff         string
}

// GatherInput collects git context for the given title.
// dir is used as the working directory for git commands.
func GatherInput(dir, title string) PromptInput {
	files := changedFiles(dir)
	diff := stagedDiff(dir)
	if len(diff) > DiffLimit {
		diff = diff[:DiffLimit]
	}
	return PromptInput{
		Title:        title,
		ChangedFiles: files,
		Diff:         diff,
	}
}

// BuildPrompt constructs the prompt string to send to the LLM.
func BuildPrompt(input PromptInput) string {
	var sb strings.Builder
	sb.WriteString("You are helping a developer write an Architecture Decision Record (ADR).\n\n")
	sb.WriteString("Title: " + input.Title + "\n\n")
	sb.WriteString("Changed files:\n")
	for _, f := range input.ChangedFiles {
		sb.WriteString("  " + f + "\n")
	}
	sb.WriteString("\nCode diff (truncated):\n")
	sb.WriteString(input.Diff)
	sb.WriteString("\n\nGenerate a draft ADR in the following Markdown format.\n")
	sb.WriteString("Output the Markdown only. No explanation or preamble.\n\n")
	sb.WriteString("## Context\n")
	sb.WriteString("(Why this decision was needed. Background, constraints, and problem.)\n\n")
	sb.WriteString("## Decision\n")
	sb.WriteString("(What was decided. 1-3 sentences.)\n\n")
	sb.WriteString("## Consequences\n")
	sb.WriteString("(Impact, benefits, and trade-offs.)\n\n")
	sb.WriteString("## Related Files\n")
	sb.WriteString("(List the related files from the changed files above.)\n")
	return sb.String()
}

// BuildDesignPrompt constructs the prompt string for generating a design document draft.
func BuildDesignPrompt(input PromptInput) string {
	var sb strings.Builder
	sb.WriteString("You are helping a developer write a software design document.\n\n")
	sb.WriteString("Title: " + input.Title + "\n\n")
	sb.WriteString("Changed files:\n")
	for _, f := range input.ChangedFiles {
		sb.WriteString("  " + f + "\n")
	}
	sb.WriteString("\nCode diff (truncated):\n")
	sb.WriteString(input.Diff)
	sb.WriteString("\n\nGenerate a draft design document in the following Markdown format.\n")
	sb.WriteString("Output the Markdown only. No explanation or preamble.\n\n")
	sb.WriteString("## Overview\n")
	sb.WriteString("(1-3 sentences summarizing what this design does and why.)\n\n")
	sb.WriteString("## Background\n")
	sb.WriteString("(Why this design was needed. Context, problem, and constraints.)\n\n")
	sb.WriteString("## Goals / Non-Goals\n")
	sb.WriteString("(Goals: what this design achieves. Non-Goals: what it explicitly does not cover.)\n\n")
	sb.WriteString("## Design\n")
	sb.WriteString("(The actual design: structure, flow, interfaces, data models, etc.)\n\n")
	sb.WriteString("## Implementation Plan\n")
	sb.WriteString("(Steps to implement this design. Omit if the scope is small.)\n\n")
	sb.WriteString("## Open Questions\n")
	sb.WriteString("(Unresolved questions at design time.)\n\n")
	sb.WriteString("## Related Files\n")
	sb.WriteString("(List the related files from the changed files above.)\n")
	return sb.String()
}

func changedFiles(dir string) []string {
	staged := gitDiffFiles(dir, "--staged")
	unstaged := gitDiffFiles(dir, "")
	seen := make(map[string]struct{})
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

func stagedDiff(dir string) string {
	cmd := exec.Command("git", "diff", "--staged")
	cmd.Dir = dir
	out, _ := cmd.Output()
	return string(out)
}
