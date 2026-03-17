package decision

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Use Go over Shell Script", "use-go-over-shell-script"},
		{"Use Cobra for CLI Framework", "use-cobra-for-cli-framework"},
		{"  leading and trailing spaces  ", "leading-and-trailing-spaces"},
		{"multiple   spaces", "multiple-spaces"},
		{"special!@#chars", "special-chars"},
		{"already-kebab-case", "already-kebab-case"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNextID_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	id, err := NextID(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("NextID (empty dir) = %d, want 1", id)
	}
}

func TestNextID_NonExistentDir(t *testing.T) {
	id, err := NextID(filepath.Join(t.TempDir(), "nonexistent"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("NextID (nonexistent dir) = %d, want 1", id)
	}
}

func TestNextID_WithExistingFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"0001-use-go.md",
		"0002-use-cobra.md",
		"0005-some-decision.md",
		"not-a-decision.txt",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	id, err := NextID(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 6 {
		t.Errorf("NextID = %d, want 6", id)
	}
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	path, err := Create(dir, "Use PostgreSQL", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFile := "0001-use-postgresql.md"
	if filepath.Base(path) != wantFile {
		t.Errorf("filename = %q, want %q", filepath.Base(path), wantFile)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	body := string(content)
	if !contains(body, "# 0001: Use PostgreSQL") {
		t.Errorf("file missing title header, got:\n%s", body)
	}
	if !contains(body, "Status: Draft") {
		t.Errorf("file missing status, got:\n%s", body)
	}
}

func TestCreateDesign(t *testing.T) {
	dir := t.TempDir()
	path, err := CreateDesign(dir, "Connection Pool Design", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFile := "0001-connection-pool-design.md"
	if filepath.Base(path) != wantFile {
		t.Errorf("filename = %q, want %q", filepath.Base(path), wantFile)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	body := string(content)
	for _, want := range []string{"- Type: Design", "- Status: Draft", "## Overview", "## Background"} {
		if !contains(body, want) {
			t.Errorf("design file missing %q, got:\n%s", want, body)
		}
	}
}

func TestParse(t *testing.T) {
	content := `# 0003: Use MADR Format

- Date: 2026-03-12
- Status: Accepted
- Author: masahiro.kasatani

## Context

Some context here.
`
	path := filepath.Join(t.TempDir(), "0003-use-madr-format.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != 3 {
		t.Errorf("ID = %d, want 3", d.ID)
	}
	if d.Title != "Use MADR Format" {
		t.Errorf("Title = %q, want %q", d.Title, "Use MADR Format")
	}
	if d.Date != "2026-03-12" {
		t.Errorf("Date = %q, want %q", d.Date, "2026-03-12")
	}
	if d.Status != "Accepted" {
		t.Errorf("Status = %q, want %q", d.Status, "Accepted")
	}
	if d.Author != "masahiro.kasatani" {
		t.Errorf("Author = %q, want %q", d.Author, "masahiro.kasatani")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"0001-use-go.md": "# 0001: Use Go\n\n- Date: 2026-01-01\n- Status: Accepted\n- Author: alice\n",
		"0003-use-madr.md": "# 0003: Use MADR\n\n- Date: 2026-03-01\n- Status: Proposed\n- Author: alice\n",
		"0002-use-cobra.md": "# 0002: Use Cobra\n\n- Date: 2026-02-01\n- Status: Accepted\n- Author: alice\n",
		"not-a-decision.txt": "ignored",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	decisions, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 3 {
		t.Fatalf("len = %d, want 3", len(decisions))
	}
	// Expect descending order: 3, 2, 1
	wantIDs := []int{3, 2, 1}
	for i, want := range wantIDs {
		if decisions[i].ID != want {
			t.Errorf("decisions[%d].ID = %d, want %d", i, decisions[i].ID, want)
		}
	}
}

func TestFindByID(t *testing.T) {
	dir := t.TempDir()
	content := "# 0002: Use Cobra\n\n- Date: 2026-03-12\n- Status: Accepted\n- Author: alice\n"
	if err := os.WriteFile(filepath.Join(dir, "0002-use-cobra.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := FindByID(dir, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != 2 {
		t.Errorf("ID = %d, want 2", d.ID)
	}
	if d.Title != "Use Cobra" {
		t.Errorf("Title = %q, want %q", d.Title, "Use Cobra")
	}
}

func TestFindByID_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindByID(dir, 99)
	if err == nil {
		t.Error("expected error for missing ID, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
