package cmd

import (
	"strings"
	"testing"
)

func TestListCmd_Empty(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	out, err := executeCmd(t, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No decisions found") {
		t.Errorf("expected 'No decisions found', got: %q", out)
	}
}

func TestListCmd_WithDecisions(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Accepted")
	seedDecision(t, dir, 2, "Use Cobra", "Deprecated")
	setTestRoot(t, root)

	out, err := executeCmd(t, "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") {
		t.Errorf("expected 'Use Go' in output, got: %q", out)
	}
	if !strings.Contains(out, "Use Cobra") {
		t.Errorf("expected 'Use Cobra' in output, got: %q", out)
	}
}

func TestListCmd_StatusFilter(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Accepted")
	seedDecision(t, dir, 2, "Use Cobra", "Deprecated")
	setTestRoot(t, root)

	out, err := executeCmd(t, "list", "--status", "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") {
		t.Errorf("expected 'Use Go' in filtered output, got: %q", out)
	}
	if strings.Contains(out, "Use Cobra") {
		t.Errorf("expected 'Use Cobra' to be filtered out, got: %q", out)
	}
}

func TestListCmd_StatusFilter_NoMatch(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Accepted")
	setTestRoot(t, root)

	out, err := executeCmd(t, "list", "--status", "deprecated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No decisions found") {
		t.Errorf("expected 'No decisions found', got: %q", out)
	}
}
