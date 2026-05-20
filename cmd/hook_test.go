package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func stageFile(t *testing.T, root, rel string) {
	t.Helper()
	if err := exec.Command("git", "-C", root, "add", rel).Run(); err != nil {
		t.Fatalf("git add %s: %v", rel, err)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHookPreCommit_NoStagedFiles(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output for empty staging area, got: %q", out)
	}
}

func TestHookPreCommit_StagedFileMatchesRelatedFiles(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	setTestRoot(t, root)

	// Create a document with a related file, then stage the related file.
	docPath := seedDecision(t, dir, 1, "Use connection pooling", "Active")
	appendRelatedFile(t, docPath, "internal/db/db.go")

	writeFile(t, root, "internal/db/db.go", "package db\n")
	stageFile(t, root, "internal/db/db.go")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "use-connection-pooling") {
		t.Errorf("expected slug in output, got: %q", out)
	}
	if !strings.Contains(out, "⚠️") {
		t.Errorf("expected warning in output, got: %q", out)
	}
}

func TestHookPreCommit_DocumentStaged_NoWarning(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	setTestRoot(t, root)

	docPath := seedDecision(t, dir, 1, "Use connection pooling", "Active")
	appendRelatedFile(t, docPath, "internal/db/db.go")

	writeFile(t, root, "internal/db/db.go", "package db\n")
	stageFile(t, root, "internal/db/db.go")
	stageFile(t, root, filepath.ToSlash(func() string {
		rel, _ := filepath.Rel(root, docPath)
		return rel
	}()))

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning when document is staged, got: %q", out)
	}
}

func TestHookPreCommit_NoMatch_SuggestsNewDocument(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	// Stage a non-MD file with no related documents.
	writeFile(t, root, "internal/db/db.go", "package db\n")
	stageFile(t, root, "internal/db/db.go")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kizami adr") {
		t.Errorf("expected 'kizami adr' suggestion in output, got: %q", out)
	}
}

func TestHookPreCommit_OnlyMDFiles_NoWarning(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	writeFile(t, root, "README.md", "# readme\n")
	stageFile(t, root, "README.md")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning for MD-only commit, got: %q", out)
	}
}

func TestHookPreCommit_DocInDecisionsDirStaged_NoCheck2(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	setTestRoot(t, root)

	// Stage a document (not a source file), no source changes.
	docPath := seedDecision(t, dir, 1, "Use Go", "Active")
	stageFile(t, root, filepath.ToSlash(func() string {
		rel, _ := filepath.Rel(root, docPath)
		return rel
	}()))

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning when only document is staged, got: %q", out)
	}
}
