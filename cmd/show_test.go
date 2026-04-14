package cmd

import (
	"strings"
	"testing"
)

func TestShowCmd_NotFound(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	_, err := executeCmd(t, "show", "nonexistent-slug")
	if err == nil {
		t.Fatal("expected error for non-existent slug")
	}
}

func TestShowCmd_ValidSlug(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	out, err := executeCmd(t, "show", "use-go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") {
		t.Errorf("expected decision title in output, got: %q", out)
	}
}

func TestShowCmd_SlugCollision_ShowsAll(t *testing.T) {
	root := newTestRepo(t)
	seedMultiDirConfig(t, root)
	seedDecision(t, decisionsPath(root), 1, "Use Go", "Active")
	seedDecision(t, designPath(root), 1, "Use Go", "Active")
	setTestRoot(t, root)

	out, err := executeCmd(t, "show", "use-go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both entries should appear in output along with file path headers.
	if strings.Count(out, "Use Go") < 2 {
		t.Errorf("expected both docs in output, got: %q", out)
	}
	if !strings.Contains(out, "docs/decisions") || !strings.Contains(out, "docs/design") {
		t.Errorf("expected file path headers in output, got: %q", out)
	}
}
