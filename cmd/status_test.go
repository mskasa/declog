package cmd

import (
	"strings"
	"testing"
)

func TestStatusCmd_InvalidStatus(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "use-go", "unknown")
	if err == nil {
		t.Fatal("expected error for unknown status")
	}
}

func TestStatusCmd_ValidUpdate(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Draft")
	setTestRoot(t, root)

	out, err := executeCmd(t, "status", "use-go", "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "use-go") {
		t.Errorf("expected slug in output, got: %q", out)
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

	_, err := executeCmd(t, "status", "use-go", "accepted", "--by", "use-another")
	if err == nil {
		t.Fatal("expected error when --by is used without superseded status")
	}
}
