package context

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBySlug_SingleMatch(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.", []string{"a.go"})

	docs, err := GetBySlug([]string{dir}, root, "use-go")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 match, got %d", len(docs))
	}
	if !strings.Contains(docs[0].Markdown, "Use Go.") {
		t.Errorf("expected full markdown content, got: %s", docs[0].Markdown)
	}
}

func TestGetBySlug_MultipleDirsSameSlug(t *testing.T) {
	root := t.TempDir()
	enDir := filepath.Join(root, "docs", "decisions")
	jaDir := filepath.Join(root, "docs", "decisions", "ja")
	writeDoc(t, enDir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.", []string{"a.go"})
	writeDoc(t, jaDir, "2026-01-01-use-go.md", "Active", "## Decision", "Goを使う。", []string{"a.go"})

	// Both live under the same configured decisions dir (ja/ is a subdirectory), matching
	// how this repository itself organizes EN/JA pairs.
	docs, err := GetBySlug([]string{enDir}, root, "use-go")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 matches (EN+JA), got %d", len(docs))
	}
}

func TestGetBySlug_NotFound(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-use-go.md", "Active", "## Decision", "Use Go.", []string{"a.go"})

	docs, err := GetBySlug([]string{dir}, root, "does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(docs))
	}
}
