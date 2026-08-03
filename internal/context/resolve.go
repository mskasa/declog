package context

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mskasa/kizami/internal/decision"
)

// Resolve returns the decisions (across dirs, expected to be absolute paths already joined
// with the repo root) that govern the given files, plus which of the queried files matched
// no decision at all. full controls whether each decision's summary is its full document
// content or just the "## Decision"/"## Overview" section.
func Resolve(dirs []string, repoRoot string, files []string, full bool) (*Result, error) {
	result := &Result{
		Version: Version,
		Query:   files,
	}

	docs, err := collectGoverningDocs(dirs)
	if err != nil {
		return nil, err
	}

	matched := make(map[string]struct{})
	for _, d := range docs {
		dr, ok, err := resolveDecision(d, repoRoot, files, full)
		if err != nil {
			return nil, fmt.Errorf("resolving %s: %w", d.File, err)
		}
		if !ok {
			continue
		}
		for _, m := range dr.Matched {
			matched[m.File] = struct{}{}
		}
		result.Decisions = append(result.Decisions, dr)
	}

	sort.Slice(result.Decisions, func(i, j int) bool {
		return result.Decisions[i].Slug < result.Decisions[j].Slug
	})

	for _, f := range files {
		if _, ok := matched[f]; !ok {
			result.Unmatched = append(result.Unmatched, f)
		}
	}

	return result, nil
}

// governs reports whether a decision's status is one the resolver surfaces: Active decisions,
// and Superseded decisions (kept, not dropped — see the design doc's Design section).
// Draft and Inactive decisions are excluded.
func governs(d *decision.Decision) bool {
	return strings.EqualFold(d.Status, "Active") || d.SupersededBy() != ""
}

// collectGoverningDocs returns every decision across dirs that governs (see governs),
// deduplicated by file. Shared by Resolve and Manifest — both start from the same set of
// documents, they just do different things with them afterward.
func collectGoverningDocs(dirs []string) ([]*decision.Decision, error) {
	seen := make(map[string]struct{})
	var docs []*decision.Decision
	for _, dir := range dirs {
		list, err := decision.List(dir)
		if err != nil {
			return nil, err
		}
		for _, d := range list {
			if _, already := seen[d.File]; already {
				continue
			}
			seen[d.File] = struct{}{}
			if governs(d) {
				docs = append(docs, d)
			}
		}
	}
	return docs, nil
}

// collectAllDocs returns every decision across dirs, deduplicated by file, regardless of
// status. Used where status filtering doesn't apply — e.g. Search, which should surface a
// Draft decision too if it matches, unlike Resolve/Manifest which only surface what governs.
func collectAllDocs(dirs []string) ([]*decision.Decision, error) {
	seen := make(map[string]struct{})
	var docs []*decision.Decision
	for _, dir := range dirs {
		list, err := decision.List(dir)
		if err != nil {
			return nil, err
		}
		for _, d := range list {
			if _, already := seen[d.File]; already {
				continue
			}
			seen[d.File] = struct{}{}
			docs = append(docs, d)
		}
	}
	return docs, nil
}

// relPath returns file relative to repoRoot with forward slashes, or file itself if it
// isn't under repoRoot. Repo-relative paths are meaningful outside the local filesystem
// (e.g. to an agent that only knows the repo tree), unlike an absolute path.
func relPath(file, repoRoot string) string {
	rel, err := filepath.Rel(repoRoot, file)
	if err != nil {
		return file
	}
	return filepath.ToSlash(rel)
}

// resolveDecision builds a DecisionResult for d if at least one of files matches its
// Related Files. ok is false if none matched (d should be skipped, not included empty).
func resolveDecision(d *decision.Decision, repoRoot string, files []string, full bool) (*DecisionResult, bool, error) {
	related, err := decision.ParseRelatedFiles(d.File)
	if err != nil {
		return nil, false, err
	}
	if len(related) == 0 {
		return nil, false, nil
	}

	var matchedFiles []MatchedFile
	for _, f := range files {
		for _, entry := range related {
			kind, ok := decision.Match(entry, f)
			if !ok {
				continue
			}
			matchedFiles = append(matchedFiles, MatchedFile{File: f, Rule: entry, Kind: string(kind)})
			break
		}
	}
	if len(matchedFiles) == 0 {
		return nil, false, nil
	}

	summary, err := summaryText(d.File, full)
	if err != nil {
		return nil, false, err
	}

	var missing []string
	for _, entry := range related {
		if !decision.EntryExists(repoRoot, entry) {
			missing = append(missing, entry)
		}
	}
	state := "ok"
	if len(missing) > 0 {
		state = "drift"
	}

	var lastUpdated string
	if t, err := decision.LastUpdated(d.File); err == nil {
		lastUpdated = t.Format(time.RFC3339)
	}

	return &DecisionResult{
		Slug:         d.Slug,
		Title:        d.Title,
		Status:       d.Status,
		Date:         d.Date,
		Path:         relPath(d.File, repoRoot),
		Decision:     summary,
		SupersededBy: d.SupersededBy(),
		Matched:      matchedFiles,
		Drift:        Drift{State: state, Missing: missing},
		LastUpdated:  lastUpdated,
	}, true, nil
}
