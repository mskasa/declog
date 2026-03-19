package cmd

import (
	"strings"
	"testing"
)

func TestShowCmd_InvalidID(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	_, err := executeCmd(t, "show", "abc")
	if err == nil {
		t.Fatal("expected error for non-integer ID")
	}
}

func TestShowCmd_ZeroID(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	_, err := executeCmd(t, "show", "0")
	if err == nil {
		t.Fatal("expected error for ID=0")
	}
}

func TestShowCmd_NotFound(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	_, err := executeCmd(t, "show", "999")
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func TestShowCmd_ValidID(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	out, err := executeCmd(t, "show", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") {
		t.Errorf("expected decision title in output, got: %q", out)
	}
}
