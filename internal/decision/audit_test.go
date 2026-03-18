package decision

import (
	"os"
	"path/filepath"
	"testing"
)

func writeAuditADR(t *testing.T, dir, name, status string, relatedFiles []string) {
	t.Helper()
	content := "# 0001: Test\n\n- Date: 2026-01-01\n- Status: " + status + "\n- Author: alice\n\n## Context\n\nSome context.\n\n## Related Files\n\n"
	for _, f := range relatedFiles {
		content += "- `" + f + "`\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseRelatedFiles_WithBackticks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0001-test.md")
	content := "# 0001: Test\n\n## Related Files\n\n- `internal/foo.go`\n- `cmd/bar.go`\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := ParseRelatedFiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 || files[0] != "internal/foo.go" || files[1] != "cmd/bar.go" {
		t.Errorf("got %v", files)
	}
}

func TestParseRelatedFiles_WithoutBackticks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0001-test.md")
	content := "# 0001: Test\n\n## Related Files\n\n- internal/foo.go\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := ParseRelatedFiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "internal/foo.go" {
		t.Errorf("got %v", files)
	}
}

func TestParseRelatedFiles_SkipsComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "0001-test.md")
	content := "# 0001: Test\n\n## Related Files\n\n<!-- List files here. -->\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := ParseRelatedFiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no files, got %v", files)
	}
}

func TestAudit_DetectsMissingFiles(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	// Create a real file.
	existing := filepath.Join(repoRoot, "internal", "foo.go")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeAuditADR(t, dir, "0001-test.md", "Active", []string{"internal/foo.go", "internal/missing.go"})

	results, err := Audit(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(results[0].MissingFiles) != 1 || results[0].MissingFiles[0] != "internal/missing.go" {
		t.Errorf("unexpected missing files: %v", results[0].MissingFiles)
	}
}

func TestAudit_AllFilesExist(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	existing := filepath.Join(repoRoot, "internal", "foo.go")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeAuditADR(t, dir, "0001-test.md", "Active", []string{"internal/foo.go"})

	results, err := Audit(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestAudit_SkipsNonActive(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	writeAuditADR(t, dir, "0001-draft.md", "Draft", []string{"internal/missing.go"})
	writeAuditADR(t, dir, "0002-inactive.md", "Inactive", []string{"internal/missing.go"})
	writeAuditADR(t, dir, "0003-superseded.md", "Superseded by 0004", []string{"internal/missing.go"})

	results, err := Audit(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for non-Active documents, got %d", len(results))
	}
}

func TestAudit_SkipsEmptyRelatedFiles(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	writeAuditADR(t, dir, "0001-test.md", "Active", []string{})

	results, err := Audit(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results for ADR with no related files, got %d", len(results))
	}
}
