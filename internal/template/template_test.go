package template

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitRepo initializes a git repository in dir with a dummy initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("setup cmd %v: %v\n%s", args, err, out)
		}
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func gitAdd(t *testing.T, dir, file string) {
	t.Helper()
	c := exec.Command("git", "add", file)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func TestChangedFiles_NoChanges(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	files := ChangedFiles(dir)
	if len(files) != 0 {
		t.Errorf("expected no files, got %v", files)
	}
}

func TestChangedFiles_StagedOnly(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	writeFile(t, dir, "foo.go", "package main")
	gitAdd(t, dir, "foo.go")

	files := ChangedFiles(dir)
	if len(files) != 1 || files[0] != "foo.go" {
		t.Errorf("expected [foo.go], got %v", files)
	}
}

func TestChangedFiles_UnstagedOnly(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Commit the file first, then modify it without staging.
	writeFile(t, dir, "bar.go", "package main")
	gitAdd(t, dir, "bar.go")
	c := exec.Command("git", "commit", "-m", "add bar")
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	writeFile(t, dir, "bar.go", "package main\n// modified")

	files := ChangedFiles(dir)
	if len(files) != 1 || files[0] != "bar.go" {
		t.Errorf("expected [bar.go], got %v", files)
	}
}

func TestChangedFiles_StagedAndUnstaged(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Commit unstaged.go first, then modify it without staging.
	writeFile(t, dir, "unstaged.go", "package main")
	gitAdd(t, dir, "unstaged.go")
	c := exec.Command("git", "commit", "-m", "add unstaged")
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	writeFile(t, dir, "unstaged.go", "package main\n// modified")

	// Stage staged.go as a new file.
	writeFile(t, dir, "staged.go", "package main")
	gitAdd(t, dir, "staged.go")

	files := ChangedFiles(dir)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %v", files)
	}
	set := map[string]bool{"staged.go": true, "unstaged.go": true}
	for _, f := range files {
		if !set[f] {
			t.Errorf("unexpected file: %s", f)
		}
	}
}

func TestChangedFiles_Deduplicated(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	// Commit a file, then modify and stage it — appears in both staged and unstaged.
	writeFile(t, dir, "both.go", "package main")
	gitAdd(t, dir, "both.go")
	c := exec.Command("git", "commit", "-m", "add both")
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	writeFile(t, dir, "both.go", "package main\n// v2")
	gitAdd(t, dir, "both.go")
	writeFile(t, dir, "both.go", "package main\n// v3")

	files := ChangedFiles(dir)
	count := 0
	for _, f := range files {
		if f == "both.go" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected both.go exactly once, got %v", files)
	}
}

func TestChangedFiles_OutsideGitRepo(t *testing.T) {
	dir := t.TempDir() // not a git repo

	files := ChangedFiles(dir)
	if len(files) != 0 {
		t.Errorf("expected no files outside git repo, got %v", files)
	}
}

func TestRender_WithRelatedFiles(t *testing.T) {
	out := Render("Test Decision", "alice", []string{"internal/foo.go", "cmd/bar.go"}, "")

	if !strings.Contains(out, "- `internal/foo.go`") {
		t.Errorf("missing internal/foo.go in output:\n%s", out)
	}
	if !strings.Contains(out, "- `cmd/bar.go`") {
		t.Errorf("missing cmd/bar.go in output:\n%s", out)
	}
	if strings.Contains(out, "<!-- List files") {
		t.Errorf("placeholder comment should not appear when files are provided:\n%s", out)
	}
}

func TestRender_WithoutRelatedFiles(t *testing.T) {
	out := Render("Test Decision", "alice", nil, "")

	if !strings.Contains(out, "<!-- List files related to this decision") {
		t.Errorf("expected placeholder comment in output:\n%s", out)
	}
}

func TestRender_WithSupersedes(t *testing.T) {
	out := Render("New Decision", "alice", nil, "use-old-decision")

	if !strings.Contains(out, "- Supersedes: use-old-decision") {
		t.Errorf("missing Supersedes line in output:\n%s", out)
	}
}

func TestRender_WithoutSupersedes(t *testing.T) {
	out := Render("New Decision", "alice", nil, "")

	if strings.Contains(out, "Supersedes") {
		t.Errorf("unexpected Supersedes line in output:\n%s", out)
	}
}

func TestRender_DefaultStatusDraft(t *testing.T) {
	out := Render("Test Decision", "alice", nil, "")

	if !strings.Contains(out, "- Status: Draft") {
		t.Errorf("expected Status: Draft in ADR template:\n%s", out)
	}
	if !strings.Contains(out, "- Type: ADR") {
		t.Errorf("expected Type: ADR in ADR template:\n%s", out)
	}
}

func TestRenderDesign_Sections(t *testing.T) {
	out := RenderDesign("Test Design", "alice", nil, "")

	for _, want := range []string{
		"- Type: Design",
		"- Status: Draft",
		"## Overview",
		"## Background",
		"## Goals / Non-Goals",
		"## Design",
		"## Implementation Plan",
		"## Open Questions",
		"## Related Files",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("design template missing %q:\n%s", want, out)
		}
	}
}

func TestRenderDesign_WithSupersedes(t *testing.T) {
	out := RenderDesign("New Design", "alice", nil, "use-old-design")

	if !strings.Contains(out, "- Supersedes: use-old-design") {
		t.Errorf("missing Supersedes line in output:\n%s", out)
	}
}
