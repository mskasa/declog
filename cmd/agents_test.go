package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentsSyncCmd_NoTargetFile_Errors(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	_, err := executeCmd(t, "agents", "sync")
	if err == nil {
		t.Fatal("expected error when no target file exists")
	}
}

func TestAgentsSyncCmd_WritesIntoExistingClaudeMd(t *testing.T) {
	root := newTestRepo(t)
	claudeMd := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claudeMd, []byte("# Project\n\nHand-written notes.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use Go", "Active")
	appendRelatedFile(t, path, "internal/decision/audit.go")
	setTestRoot(t, root)

	out, err := executeCmd(t, "agents", "sync")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Updated CLAUDE.md") {
		t.Errorf("expected update message, got: %q", out)
	}

	content, err := os.ReadFile(claudeMd)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Hand-written notes.") {
		t.Error("expected hand-written content preserved")
	}
	if !strings.Contains(string(content), "use-go") {
		t.Errorf("expected decision slug in synced content, got: %s", content)
	}
}

func TestAgentsSyncCmd_SecondRunIsNoop(t *testing.T) {
	root := newTestRepo(t)
	claudeMd := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claudeMd, []byte("# Project\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	if _, err := executeCmd(t, "agents", "sync"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := executeCmd(t, "agents", "sync")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "already up to date") {
		t.Errorf("expected no-op message on second run, got: %q", out)
	}
}

func TestAgentsSyncCheck_FailsWhenStale(t *testing.T) {
	t.Cleanup(func() { agentsSyncCheck = false })
	root := newTestRepo(t)
	claudeMd := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claudeMd, []byte("# Project\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setTestRoot(t, root)

	_, err := executeCmd(t, "agents", "sync", "--check")
	if err == nil {
		t.Fatal("expected error when manifest block is missing")
	}
}

func TestAgentsSyncCheck_PassesAfterSync(t *testing.T) {
	t.Cleanup(func() { agentsSyncCheck = false })
	root := newTestRepo(t)
	claudeMd := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(claudeMd, []byte("# Project\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setTestRoot(t, root)

	if _, err := executeCmd(t, "agents", "sync"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := executeCmd(t, "agents", "sync", "--check")
	if err != nil {
		t.Fatalf("unexpected error after sync: %v, output: %q", err, out)
	}
}
