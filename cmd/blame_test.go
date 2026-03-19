package cmd

import (
	"strings"
	"testing"
)

func TestBlameCmd_NoResults(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	out, err := executeCmd(t, "blame", "internal/some/file.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No decisions found") {
		t.Errorf("expected 'No decisions found', got: %q", out)
	}
}

func TestBlameCmd_WithResults(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use Go", "Active")

	// Append a Related Files section referencing a specific file.
	appendRelatedFile(t, path, "internal/search/search.go")
	setTestRoot(t, root)

	out, err := executeCmd(t, "blame", "internal/search/search.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") {
		t.Errorf("expected decision title in output, got: %q", out)
	}
}
