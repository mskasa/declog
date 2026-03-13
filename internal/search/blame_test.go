package search

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const adr0001 = "# 0001: Use Go\n\n- Date: 2025-01-10\n- Status: Active\n- Author: Alice\n\n## Related Files\n\ncmd/root.go\n"
const adr0003 = "# 0003: Connection Pool Size\n\n- Date: 2025-01-15\n- Status: Active\n- Author: Bob\n\n## Related Files\n\ndatabase/connection.go\n"
const adr0007 = "# 0007: Switch to sqlx\n\n- Date: 2025-03-02\n- Status: Active\n- Author: Bob\n\n## Related Files\n\ndatabase/connection.go\ninternal/db/query.go\n"

func TestBlameStdlib_Match(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", adr0001)
	writeFile(t, dir, "0003-connection-pool-size.md", adr0003)
	writeFile(t, dir, "0007-switch-to-sqlx.md", adr0007)

	decisions, err := blameStdlib(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 2 {
		t.Fatalf("len = %d, want 2", len(decisions))
	}

	wantFiles := []string{
		filepath.Join(dir, "0003-connection-pool-size.md"),
		filepath.Join(dir, "0007-switch-to-sqlx.md"),
	}
	for i, f := range decisions {
		if f != wantFiles[i] {
			t.Errorf("decisions[%d] = %q, want %q", i, f, wantFiles[i])
		}
	}
}

func TestBlameStdlib_NoMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", adr0001)

	decisions, err := blameStdlib(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected no results, got %d", len(decisions))
	}
}

func TestBlame_Match(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0003-connection-pool-size.md", adr0003)
	writeFile(t, dir, "0007-switch-to-sqlx.md", adr0007)

	decisions, err := Blame(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 2 {
		t.Fatalf("len = %d, want 2", len(decisions))
	}
	if decisions[0].ID != 3 {
		t.Errorf("decisions[0].ID = %d, want 3", decisions[0].ID)
	}
	if decisions[1].ID != 7 {
		t.Errorf("decisions[1].ID = %d, want 7", decisions[1].ID)
	}
}

func TestBlame_NoMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", adr0001)

	decisions, err := Blame(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected no results, got %d", len(decisions))
	}
}

func TestBlame_DeduplicatesFiles(t *testing.T) {
	// A file that mentions the target path twice should appear only once.
	dir := t.TempDir()
	content := "# 0005: Duplicate Mention\n\n- Date: 2025-02-01\n- Status: Active\n- Author: Alice\n\ndatabase/connection.go is used here.\nAlso database/connection.go appears again.\n"
	writeFile(t, dir, "0005-duplicate-mention.md", content)

	decisions, err := Blame(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Errorf("expected 1 result (deduplicated), got %d", len(decisions))
	}
}

func TestBlameRipgrep(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep not installed")
	}

	dir := t.TempDir()
	writeFile(t, dir, "0003-connection-pool-size.md", adr0003)
	writeFile(t, dir, "0001-use-go.md", adr0001)

	files, err := blameRipgrep(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only 0003 mentions database/connection.go
	got := make(map[string]bool)
	for _, f := range files {
		got[filepath.Base(f)] = true
	}
	if !got["0003-connection-pool-size.md"] {
		t.Errorf("expected 0003-connection-pool-size.md in results")
	}
	if got["0001-use-go.md"] {
		t.Errorf("0001-use-go.md should not appear in results")
	}
}

func TestBlame_NonMarkdownSkipped(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "0001-use-go.md", adr0001)
	// Write a non-markdown file that also mentions the keyword
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("database/connection.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	decisions, err := Blame(dir, "database/connection.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected 0 results (non-md file should be skipped), got %d", len(decisions))
	}
}
