package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeDoc(t *testing.T, dir, name, status, decisionOrOverviewHeading, body string, relatedFiles []string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	sb.WriteString("# Test\n\n- Date: 2026-01-01\n- Status: " + status + "\n- Author: alice\n\n")
	sb.WriteString(decisionOrOverviewHeading + "\n\n" + body + "\n\n## Related Files\n\n")
	for _, f := range relatedFiles {
		sb.WriteString("- `" + f + "`\n")
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolve_ActiveDecisionMatches(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.", []string{"internal/decision/audit.go"})

	if err := os.MkdirAll(filepath.Join(root, "internal", "decision"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "decision", "audit.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(result.Decisions))
	}
	d := result.Decisions[0]
	if d.Slug != "use-go" {
		t.Errorf("unexpected slug: %s", d.Slug)
	}
	if d.Decision != "Use Go." {
		t.Errorf("unexpected decision summary: %q", d.Decision)
	}
	if len(d.Matched) != 1 || d.Matched[0].Kind != "exact" {
		t.Errorf("unexpected matched: %+v", d.Matched)
	}
	if d.Drift.State != "ok" {
		t.Errorf("expected drift ok, got %s", d.Drift.State)
	}
	if len(result.Unmatched) != 0 {
		t.Errorf("expected no unmatched files, got %v", result.Unmatched)
	}
}

func TestResolve_DraftAndInactiveExcluded(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-draft.md", "Draft", "## Decision", "TBD.", []string{"internal/decision/audit.go"})
	writeDoc(t, dir, "2026-01-02-inactive.md", "Inactive", "## Decision", "Old.", []string{"internal/decision/audit.go"})

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 0 {
		t.Fatalf("expected 0 decisions, got %d: %+v", len(result.Decisions), result.Decisions)
	}
	if len(result.Unmatched) != 1 {
		t.Errorf("expected the file to be unmatched, got %v", result.Unmatched)
	}
}

func TestResolve_SupersededIncludedWithTarget(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-old.md", "Superseded by new-approach", "## Decision", "Old approach.", []string{"internal/decision/audit.go"})

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(result.Decisions))
	}
	if result.Decisions[0].SupersededBy != "new-approach" {
		t.Errorf("expected superseded_by = new-approach, got %q", result.Decisions[0].SupersededBy)
	}
}

func TestResolve_DriftWhenRelatedFileMissing(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.",
		[]string{"internal/decision/audit.go", "internal/decision/missing.go"})

	if err := os.MkdirAll(filepath.Join(root, "internal", "decision"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "decision", "audit.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(result.Decisions))
	}
	d := result.Decisions[0]
	if d.Drift.State != "drift" {
		t.Errorf("expected drift state, got %s", d.Drift.State)
	}
	if len(d.Drift.Missing) != 1 || d.Drift.Missing[0] != "internal/decision/missing.go" {
		t.Errorf("unexpected missing: %v", d.Drift.Missing)
	}
}

func TestResolve_FullReturnsWholeDocument(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.", []string{"internal/decision/audit.go"})

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit.go"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Decisions[0].Decision, "# Test") {
		t.Errorf("expected full document content, got %q", result.Decisions[0].Decision)
	}
}

func TestResolve_OverviewHeadingForDesignDocs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "design")
	writeDoc(t, dir, "2026-01-01-design.md", "Active", "## Overview", "A design.", []string{"internal/decision/audit.go"})

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 1 || result.Decisions[0].Decision != "A design." {
		t.Fatalf("unexpected result: %+v", result.Decisions)
	}
}

func TestResolve_GlobEntryMatches(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-tests.md", "Active", "## Decision", "Test files matter.",
		[]string{"internal/**/*_test.go"})

	result, err := Resolve([]string{dir}, root, []string{"internal/decision/audit_test.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(result.Decisions))
	}
	if result.Decisions[0].Matched[0].Kind != "glob" {
		t.Errorf("expected glob match kind, got %s", result.Decisions[0].Matched[0].Kind)
	}
	// Glob entries aren't existence-checked, so drift must stay "ok".
	if result.Decisions[0].Drift.State != "ok" {
		t.Errorf("expected drift ok for glob entry, got %s", result.Decisions[0].Drift.State)
	}
}

func TestResolve_MultipleDirsSortedBySlug(t *testing.T) {
	root := t.TempDir()
	decisionsDir := filepath.Join(root, "docs", "decisions")
	designDir := filepath.Join(root, "docs", "design")
	writeDoc(t, decisionsDir, "2026-01-01-z-decision.md", "Active", "## Decision", "Z.", []string{"file.go"})
	writeDoc(t, designDir, "2026-01-01-a-design.md", "Active", "## Overview", "A.", []string{"file.go"})

	result, err := Resolve([]string{decisionsDir, designDir}, root, []string{"file.go"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Decisions) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(result.Decisions))
	}
	if result.Decisions[0].Slug != "a-design" || result.Decisions[1].Slug != "z-decision" {
		t.Errorf("expected sorted by slug, got %s, %s", result.Decisions[0].Slug, result.Decisions[1].Slug)
	}
}
