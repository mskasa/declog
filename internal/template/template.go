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
func RenderHeader(title, author, supersededSlug string) string {
	date := time.Now().Format("2006-01-02")
	supersedes := ""
	if supersededSlug != "" {
		supersedes = fmt.Sprintf("\n- Supersedes: %s", supersededSlug)
	}
	return fmt.Sprintf("# %s\n\n- Date: %s\n- Type: ADR\n- Status: Draft\n- Author: %s%s\n",
		title, date, author, supersedes)
}

// Render returns the ADR Markdown template filled with the given values.
// relatedFiles is inserted into the Related Files section; pass nil for an empty section.
// supersededSlug, if non-empty, adds a "- Supersedes: <slug>" line after the Author line.
func Render(title, author string, relatedFiles []string, supersededSlug string) string {
	date := time.Now().Format("2006-01-02")
	supersedes := ""
	if supersededSlug != "" {
		supersedes = fmt.Sprintf("\n- Supersedes: %s", supersededSlug)
	}
	return fmt.Sprintf(`# %s

- Date: %s
- Type: ADR
- Status: Draft
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

%s`, title, date, author, renderRelatedFiles(relatedFiles))
}

// RenderDesignHeader returns the Design document front-matter block (no body sections).
// It is used when the body is generated externally (e.g. by an LLM).
func RenderDesignHeader(title, author, supersededSlug string) string {
	date := time.Now().Format("2006-01-02")
	supersedes := ""
	if supersededSlug != "" {
		supersedes = fmt.Sprintf("\n- Supersedes: %s", supersededSlug)
	}
	return fmt.Sprintf("# %s\n\n- Date: %s\n- Type: Design\n- Status: Draft\n- Author: %s%s\n",
		title, date, author, supersedes)
}

// RenderDesign returns the Design document Markdown template filled with the given values.
// relatedFiles is inserted into the Related Files section; pass nil for an empty section.
// supersededSlug, if non-empty, adds a "- Supersedes: <slug>" line after the Author line.
func RenderDesign(title, author string, relatedFiles []string, supersededSlug string) string {
	date := time.Now().Format("2006-01-02")
	supersedes := ""
	if supersededSlug != "" {
		supersedes = fmt.Sprintf("\n- Supersedes: %s", supersededSlug)
	}
	return fmt.Sprintf(`# %s

- Date: %s
- Type: Design
- Status: Draft
- Author: %s`+supersedes+`

## Overview

<!-- 1–3 sentences summarizing what this design does and why. -->

## Background

<!-- Why this design was needed. Describe the context, problem, and constraints. -->

## Goals / Non-Goals

<!--
Goals:
- ...

Non-Goals:
- ...
-->

## Design

<!-- The actual design: structure, flow, interfaces, data models, diagrams, etc. -->

## Implementation Plan

<!-- Steps to implement this design. Omit if the scope is small. -->

## Open Questions

<!-- Unresolved questions at design time. Update as they are answered. -->

## Related Files

%s`, title, date, author, renderRelatedFiles(relatedFiles))
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
