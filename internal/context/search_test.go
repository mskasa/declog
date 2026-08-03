package context

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestSearch_MatchesKeywordCaseInsensitive(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use PostgreSQL for storage.", []string{"a.go"})

	results, err := Search([]string{dir}, root, "postgresql", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "use-go" {
		t.Errorf("unexpected slug: %s", results[0].Slug)
	}
	if results[0].Excerpt == "" {
		t.Error("expected a non-empty excerpt")
	}
}

func TestSearch_NoMatch(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.", []string{"a.go"})

	results, err := Search([]string{dir}, root, "rust", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_ExcludesSupersededByDefault(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-old.md", "Superseded by new", "## Decision", "Old PostgreSQL choice.", []string{"a.go"})

	results, err := Search([]string{dir}, root, "postgresql", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected superseded decision excluded by default, got %d", len(results))
	}

	results, err = Search([]string{dir}, root, "postgresql", 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected superseded decision included, got %d", len(results))
	}
	if results[0].SupersededBy != "new" {
		t.Errorf("unexpected superseded_by: %s", results[0].SupersededBy)
	}
}

func TestSearch_IncludesDraftAndInactive(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-draft.md", "Draft", "## Decision", "Draft PostgreSQL idea.", []string{"a.go"})
	writeDoc(t, dir, "2026-01-02-inactive.md", "Inactive", "## Decision", "Inactive PostgreSQL idea.", []string{"b.go"})

	results, err := Search([]string{dir}, root, "postgresql", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected Draft and Inactive both included in search, got %d", len(results))
	}
}

func TestSearch_RespectsLimit(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	for i := 1; i <= 5; i++ {
		name := fmt.Sprintf("2026-01-0%d-doc.md", i)
		writeDoc(t, dir, name, "Active", "## Decision", "Uses PostgreSQL.", []string{"a.go"})
	}

	results, err := Search([]string{dir}, root, "postgresql", 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected limit=2 to cap results, got %d", len(results))
	}
}
