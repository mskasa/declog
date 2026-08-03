package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecord_ADR(t *testing.T) {
	root := t.TempDir()
	decisionsDir := filepath.Join(root, "docs", "decisions")
	designDir := filepath.Join(root, "docs", "design")

	path, slug, err := Record(decisionsDir, designDir, RecordInput{
		Title:        "Use PostgreSQL",
		ADRContext:   "SQLite hit write-lock limits under load.",
		ADRDecision:  "Switch to PostgreSQL.",
		RelatedFiles: []string{"internal/db/db.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug != "use-postgresql" {
		t.Errorf("unexpected slug: %s", slug)
	}
	if !strings.HasPrefix(path, decisionsDir) {
		t.Errorf("expected path under decisionsDir, got: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	for _, want := range []string{"- Status: Draft", "## Context", "SQLite hit write-lock limits", "## Decision", "Switch to PostgreSQL.", "## Related Files", "internal/db/db.go"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in generated document, got:\n%s", want, body)
		}
	}
}

func TestRecord_Design(t *testing.T) {
	root := t.TempDir()
	decisionsDir := filepath.Join(root, "docs", "decisions")
	designDir := filepath.Join(root, "docs", "design")

	path, slug, err := Record(decisionsDir, designDir, RecordInput{
		Kind:         "design",
		Title:        "Cache Layer Design",
		Overview:     "A caching layer for hot reads.",
		Background:   "DB load spiked during peak hours.",
		Design:       "An LRU cache in front of the repository layer.",
		RelatedFiles: []string{"internal/cache/cache.go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug != "cache-layer-design" {
		t.Errorf("unexpected slug: %s", slug)
	}
	if !strings.HasPrefix(path, designDir) {
		t.Errorf("expected path under designDir, got: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	for _, want := range []string{"- Type: Design", "- Status: Draft", "## Overview", "## Background", "## Design", "An LRU cache"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in generated document, got:\n%s", want, body)
		}
	}
}

func TestRecord_MissingRequiredADRFields(t *testing.T) {
	root := t.TempDir()
	_, _, err := Record(filepath.Join(root, "decisions"), filepath.Join(root, "design"), RecordInput{
		Title:        "Missing Decision",
		ADRContext:   "Some context.",
		RelatedFiles: []string{"a.go"},
	})
	if err == nil {
		t.Fatal("expected error when decision field is missing")
	}
}

func TestRecord_MissingRelatedFiles(t *testing.T) {
	root := t.TempDir()
	_, _, err := Record(filepath.Join(root, "decisions"), filepath.Join(root, "design"), RecordInput{
		Title:       "No Files",
		ADRContext:  "Context.",
		ADRDecision: "Decision.",
	})
	if err == nil {
		t.Fatal("expected error when related_files is empty")
	}
}

func TestRecord_UnknownKind(t *testing.T) {
	root := t.TempDir()
	_, _, err := Record(filepath.Join(root, "decisions"), filepath.Join(root, "design"), RecordInput{
		Kind:         "spec",
		Title:        "Bad Kind",
		ADRContext:   "x",
		ADRDecision:  "y",
		RelatedFiles: []string{"a.go"},
	})
	if err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

// TestRecord_StripsRedundantBackticksInRelatedFiles is a regression test for the same bug
// class found in kizami agents sync (see docs/decisions/2026-08-03-agent-manifest-sync-format.md):
// a caller-supplied entry that already contains backticks must not double up with the ones
// this function adds.
func TestRecord_StripsRedundantBackticksInRelatedFiles(t *testing.T) {
	root := t.TempDir()
	decisionsDir := filepath.Join(root, "docs", "decisions")
	designDir := filepath.Join(root, "docs", "design")

	path, _, err := Record(decisionsDir, designDir, RecordInput{
		Title:        "Use Go",
		ADRContext:   "x",
		ADRDecision:  "y",
		RelatedFiles: []string{"`internal/db/db.go`"},
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "``") {
		t.Errorf("expected no doubled-up backticks, got:\n%s", content)
	}
	if !strings.Contains(string(content), "- `internal/db/db.go`") {
		t.Errorf("expected a clean single-backtick entry, got:\n%s", content)
	}
}
