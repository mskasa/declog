package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeHookADR(t *testing.T, dir, name, status string, relatedFiles []string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	sb.WriteString("# Test\n\n- Date: 2026-01-01\n- Status: " + status + "\n- Author: alice\n\n## Context\n\n## Decision\n\n## Consequences\n\n## Related Files\n\n")
	for _, f := range relatedFiles {
		sb.WriteString("- " + f + "\n")
	}
	content := sb.String()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckHook_MatchedFileNotStaged(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Active", []string{"internal/db/db.go"})

	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "use-go" {
		t.Errorf("unexpected slug: %s", results[0].Slug)
	}
	if len(results[0].MatchedFiles) != 1 || results[0].MatchedFiles[0] != "internal/db/db.go" {
		t.Errorf("unexpected matched files: %v", results[0].MatchedFiles)
	}
}

func TestCheckHook_DocumentStaged_Skipped(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	docPath := writeHookADR(t, dir, "2026-01-01-use-go.md", "Active", []string{"internal/db/db.go"})

	relDoc, _ := filepath.Rel(root, docPath)
	staged := []string{"internal/db/db.go", filepath.ToSlash(relDoc)}
	results, err := CheckHook([]string{dir}, root, staged)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results when doc is staged, got %d", len(results))
	}
}

func TestCheckHook_InactiveSkipped(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Inactive", []string{"internal/db/db.go"})

	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for Inactive doc, got %d", len(results))
	}
}

func TestCheckHook_DraftSkipped(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Draft", []string{"internal/db/db.go"})

	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for Draft doc, got %d", len(results))
	}
}

func TestCheckHook_NoMatchingRelatedFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Active", []string{"cmd/root.go"})

	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching files, got %d", len(results))
	}
}

func TestCheckHook_DirectoryMatch(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Active", []string{"internal/db"})

	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for directory match, got %d", len(results))
	}
}

func TestCheckHook_DirectoryWithTrailingSlash(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Active", []string{"internal/db/"})

	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for directory with trailing slash, got %d", len(results))
	}
}

func TestCheckHook_NoStagedFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-go.md", "Active", []string{"internal/db/db.go"})

	results, err := CheckHook([]string{dir}, root, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty staged files, got %d", len(results))
	}
}

func TestCheckHook_MultipleDocsOnlyAffectedReturned(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeHookADR(t, dir, "2026-01-01-use-db.md", "Active", []string{"internal/db/db.go"})
	writeHookADR(t, dir, "2026-01-02-use-cache.md", "Active", []string{"internal/cache/cache.go"})

	// Only staging a file referenced by the first doc.
	results, err := CheckHook([]string{dir}, root, []string{"internal/db/db.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "use-db" {
		t.Errorf("unexpected slug: %s", results[0].Slug)
	}
}

func TestHookPathMatches(t *testing.T) {
	cases := []struct {
		related string
		staged  string
		want    bool
	}{
		{"internal/db/db.go", "internal/db/db.go", true},
		{"internal/db", "internal/db/db.go", true},
		{"internal/db/", "internal/db/db.go", true},
		{"internal/db", "internal/dbc/other.go", false},
		{"internal/db", "internal/db", true},
		{"cmd/root.go", "internal/db/db.go", false},
	}
	for _, tc := range cases {
		got := hookPathMatches(tc.related, tc.staged)
		if got != tc.want {
			t.Errorf("hookPathMatches(%q, %q) = %v, want %v", tc.related, tc.staged, got, tc.want)
		}
	}
}
