package search

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunStdlib_Match(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 0001: Use Go\n\n- Status: Accepted\n\nWe chose Go for single binary distribution.\n")
	writeFile(t, dir, "0002-use-cobra.md", "# 0002: Use Cobra\n\n- Status: Accepted\n\nCobra is the de facto standard.\n")

	results, err := runStdlib(dir, "single binary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].Line != 5 {
		t.Errorf("Line = %d, want 5", results[0].Line)
	}
	if results[0].Text != "We chose Go for single binary distribution." {
		t.Errorf("Text = %q", results[0].Text)
	}
}

func TestRunStdlib_NoMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 0001: Use Go\n\n- Status: Accepted\n")

	results, err := runStdlib(dir, "postgresql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestRunStdlib_SkipsNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 0001: Use Go\n\nGo is great.\n")
	writeFile(t, dir, "notes.txt", "Go is great.\n")

	results, err := runStdlib(dir, "Go is great")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if filepath.Ext(r.File) != ".md" {
			t.Errorf("non-markdown file included: %s", r.File)
		}
	}
}

func TestRunRipgrep(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep not installed")
	}

	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", "# 0001: Use Go\n\n- Status: Accepted\n\nWe chose Go for single binary distribution.\n")

	results, err := runRipgrep(dir, "single binary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].Text != "We chose Go for single binary distribution." {
		t.Errorf("Text = %q", results[0].Text)
	}
}

func TestRun_NonExistentDir(t *testing.T) {
	results, err := Run("/nonexistent/path/that/does/not/exist", "keyword")
	if err != nil {
		t.Fatalf("Run() on non-existent dir should not error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got: %v", results)
	}
}

func TestRunCaseInsensitive_NonExistentDir(t *testing.T) {
	results, err := RunCaseInsensitive("/nonexistent/path/that/does/not/exist", "keyword")
	if err != nil {
		t.Fatalf("RunCaseInsensitive() on non-existent dir should not error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got: %v", results)
	}
}
