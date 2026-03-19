package cmd

import (
	"strings"
	"testing"
)

func TestStatusCmd_InvalidID(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "abc", "active")
	if err == nil {
		t.Fatal("expected error for non-integer ID")
	}
}

func TestStatusCmd_InvalidStatus(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "1", "unknown")
	if err == nil {
		t.Fatal("expected error for unknown status")
	}
}

func TestStatusCmd_ValidUpdate(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Draft")
	setTestRoot(t, root)

	out, err := executeCmd(t, "status", "1", "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "0001") {
		t.Errorf("expected ID in output, got: %q", out)
	}
	if !strings.Contains(out, "Accepted") {
		t.Errorf("expected 'Accepted' in output, got: %q", out)
	}
}

func TestStatusCmd_ByFlagWithoutSuperseded(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Draft")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "1", "accepted", "--by", "2")
	if err == nil {
		t.Fatal("expected error when --by is used without superseded status")
	}
}
