package context

import (
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestManifest_OneRowPerDecision(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-a.md", "Active", "## Decision", "Decide A.", []string{"a.go", "b.go"})
	writeDoc(t, dir, "2026-01-02-b.md", "Draft", "## Decision", "Not final.", []string{"c.go"})

	entries, err := Manifest([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (Draft excluded), got %d", len(entries))
	}
	if entries[0].Slug != "a" {
		t.Errorf("unexpected slug: %s", entries[0].Slug)
	}
	if len(entries[0].RelatedFiles) != 2 {
		t.Errorf("expected 2 related files joined into one entry, got %v", entries[0].RelatedFiles)
	}
}

func TestManifest_SkipsEmptyRelatedFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "decisions")
	writeDoc(t, dir, "2026-01-01-a.md", "Active", "## Decision", "No files.", nil)

	entries, err := Manifest([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestRenderManifest_TruncatesLongDecision(t *testing.T) {
	long := strings.Repeat("word ", 100)
	entries := []*ManifestEntry{{Slug: "s", RelatedFiles: []string{"a.go"}, Decision: long}}
	out := RenderManifest(entries)
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation marker in output: %s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "| s |") && len(line) > summaryCellMaxLen+100 {
			t.Errorf("row line too long: %d chars", len(line))
		}
	}
}

func TestRenderManifest_EmptyEntries(t *testing.T) {
	out := RenderManifest(nil)
	if !strings.Contains(out, "No governing decisions found") {
		t.Errorf("expected empty-state message, got: %s", out)
	}
	if !strings.Contains(out, manifestStartMarker) || !strings.Contains(out, manifestEndMarker) {
		t.Errorf("expected markers present even when empty: %s", out)
	}
}

func TestRenderManifest_TruncatesMultiByteTextSafely(t *testing.T) {
	// Regression: truncating by byte index instead of rune index can split a multi-byte
	// UTF-8 rune (e.g. Japanese) and produce invalid UTF-8 in the generated file.
	long := strings.Repeat("決定を下した理由と背景をここに記す。", 20)
	entries := []*ManifestEntry{{Slug: "s", RelatedFiles: []string{"a.go"}, Decision: long}}
	out := RenderManifest(entries)
	if !utf8.ValidString(out) {
		t.Fatal("expected valid UTF-8 output")
	}
}

func TestRenderManifest_EscapesPipeInDecision(t *testing.T) {
	entries := []*ManifestEntry{{Slug: "s", RelatedFiles: []string{"a.go"}, Decision: "a | b"}}
	out := RenderManifest(entries)
	if !strings.Contains(out, `a \| b`) {
		t.Errorf("expected escaped pipe, got: %s", out)
	}
}
