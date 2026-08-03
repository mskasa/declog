// Package context resolves which decisions govern a given set of files, unifying the
// previously-duplicated matching logic in internal/decision and internal/search:
// docs/decisions/2026-08-03-related-files-single-definition.md
// docs/design/2026-08-03-agent-context-layer-design.md
package context

// Version is the schema version of Result, bumped on breaking JSON shape changes.
const Version = 1

// MatchedFile records that file matched a decision's Related Files entry.
type MatchedFile struct {
	File string `json:"file"`
	Rule string `json:"rule"` // the Related Files entry text that matched
	Kind string `json:"kind"` // exact | dir | glob
}

// Drift holds the existence-based drift state of a decision's Related Files entries.
// This only checks existence (what `kizami audit` already does), not whether the code
// still matches the decision's content — semantic drift detection is a separate,
// not-yet-built phase (see the design doc's Non-Goals).
type Drift struct {
	State   string   `json:"state"` // ok | drift
	Missing []string `json:"missing,omitempty"`
}

// DecisionResult is a single decision that governs one or more of the queried files.
type DecisionResult struct {
	Slug   string `json:"slug"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Date   string `json:"date"`
	// Path is repo-root-relative, not absolute — meaningful outside the local filesystem
	// (e.g. to an agent that only knows the repo tree).
	Path string `json:"path"`
	// Decision is the "## Decision"/"## Overview" section only unless Full was requested,
	// to keep the default response's token cost bounded.
	Decision string `json:"decision"`
	// SupersededBy is set when Status is "Superseded by <slug>". The decision is still
	// returned (not dropped) — see the design doc's Design section for why.
	SupersededBy string        `json:"superseded_by,omitempty"`
	Matched      []MatchedFile `json:"matched"`
	Drift        Drift         `json:"drift"`
	// LastUpdated is the last git commit date/time (RFC3339) touching this document's file,
	// omitted if the file has no git history (e.g. uncommitted).
	LastUpdated string `json:"last_updated,omitempty"`
}

// Result is the top-level response of internal/context.Resolve.
type Result struct {
	Version   int               `json:"version"`
	Query     []string          `json:"query"`
	Decisions []*DecisionResult `json:"decisions"`
	Unmatched []string          `json:"unmatched,omitempty"`
	// Truncated is reserved for future use once a per-response decision-count cap is added
	// (see the design doc's Open Questions); Resolve always sets it to false today.
	Truncated bool `json:"truncated"`
}
