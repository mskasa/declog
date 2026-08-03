package context

import (
	"fmt"
	"strings"

	"github.com/mskasa/kizami/internal/decision"
)

// RecordInput holds the fields for creating a new agent-authored decision document.
// Kind selects the template: "adr" (default) or "design". Section fields are grouped by
// which kind they apply to; unused fields for the other kind are ignored.
type RecordInput struct {
	Kind         string
	Title        string
	RelatedFiles []string

	// ADR fields.
	ADRContext   string
	ADRDecision  string
	Consequences string
	Alternatives string

	// Design fields.
	Overview           string
	Background         string
	Goals              string
	NonGoals           string
	Design             string
	ImplementationPlan string
	OpenQuestions      string
}

// Record creates a new Status: Draft document (ADR under decisionsDir, or design doc under
// designDir, per in.Kind) and returns its path and slug. It never edits or deletes an
// existing file: docs/decisions/2026-08-03-agent-authored-decision-write-path.md
func Record(decisionsDir, designDir string, in RecordInput) (path, slug string, err error) {
	if strings.TrimSpace(in.Title) == "" {
		return "", "", fmt.Errorf("title must not be empty")
	}
	if len(in.RelatedFiles) == 0 {
		return "", "", fmt.Errorf("related_files must not be empty")
	}

	kind := in.Kind
	if kind == "" {
		kind = "adr"
	}

	switch kind {
	case "adr":
		if strings.TrimSpace(in.ADRContext) == "" || strings.TrimSpace(in.ADRDecision) == "" {
			return "", "", fmt.Errorf("kind \"adr\" requires context and decision")
		}
		path, err = decision.CreateFromDraft(decisionsDir, in.Title, renderADRBody(in), "")
	case "design":
		if strings.TrimSpace(in.Overview) == "" || strings.TrimSpace(in.Background) == "" || strings.TrimSpace(in.Design) == "" {
			return "", "", fmt.Errorf("kind \"design\" requires overview, background, and design")
		}
		path, err = decision.CreateDesignFromDraft(designDir, in.Title, renderDesignBody(in), "")
	default:
		return "", "", fmt.Errorf("unknown kind %q: must be \"adr\" or \"design\"", kind)
	}
	if err != nil {
		return "", "", err
	}
	return path, decision.Slugify(in.Title), nil
}

// relatedFilesSection renders the "## Related Files" list, stripping any backticks the
// caller might have redundantly included — otherwise they'd double up with the ones this
// function adds itself, producing a malformed entry that breaks downstream parsing (the
// same class of bug found while building kizami agents sync: [[agent-manifest-sync-format]]).
func relatedFilesSection(files []string) string {
	var sb strings.Builder
	sb.WriteString("## Related Files\n\n")
	for _, f := range files {
		f = strings.Trim(strings.TrimSpace(f), "`")
		if f == "" {
			continue
		}
		sb.WriteString("- `" + f + "`\n")
	}
	return sb.String()
}

func renderADRBody(in RecordInput) string {
	var sb strings.Builder
	sb.WriteString("## Context\n\n" + strings.TrimSpace(in.ADRContext) + "\n\n")
	sb.WriteString("## Decision\n\n" + strings.TrimSpace(in.ADRDecision) + "\n\n")
	if strings.TrimSpace(in.Consequences) != "" {
		sb.WriteString("## Consequences\n\n" + strings.TrimSpace(in.Consequences) + "\n\n")
	}
	if strings.TrimSpace(in.Alternatives) != "" {
		sb.WriteString("## Alternatives Considered\n\n" + strings.TrimSpace(in.Alternatives) + "\n\n")
	}
	sb.WriteString(relatedFilesSection(in.RelatedFiles))
	return sb.String()
}

func renderDesignBody(in RecordInput) string {
	var sb strings.Builder
	sb.WriteString("## Overview\n\n" + strings.TrimSpace(in.Overview) + "\n\n")
	sb.WriteString("## Background\n\n" + strings.TrimSpace(in.Background) + "\n\n")
	if strings.TrimSpace(in.Goals) != "" || strings.TrimSpace(in.NonGoals) != "" {
		sb.WriteString("## Goals / Non-Goals\n\n")
		if strings.TrimSpace(in.Goals) != "" {
			sb.WriteString("Goals:\n" + strings.TrimSpace(in.Goals) + "\n\n")
		}
		if strings.TrimSpace(in.NonGoals) != "" {
			sb.WriteString("Non-Goals:\n" + strings.TrimSpace(in.NonGoals) + "\n\n")
		}
	}
	sb.WriteString("## Design\n\n" + strings.TrimSpace(in.Design) + "\n\n")
	if strings.TrimSpace(in.ImplementationPlan) != "" {
		sb.WriteString("## Implementation Plan\n\n" + strings.TrimSpace(in.ImplementationPlan) + "\n\n")
	}
	if strings.TrimSpace(in.OpenQuestions) != "" {
		sb.WriteString("## Open Questions\n\n" + strings.TrimSpace(in.OpenQuestions) + "\n\n")
	}
	sb.WriteString(relatedFilesSection(in.RelatedFiles))
	return sb.String()
}
