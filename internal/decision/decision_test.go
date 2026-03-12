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
	path, err := Create(dir, "Use PostgreSQL")
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
	if !contains(body, "Status: Proposed") {
		t.Errorf("file missing status, got:\n%s", body)
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
