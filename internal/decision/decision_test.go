package decision

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSlugFromFilename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"0001-use-go-over-shell-script.md", "use-go-over-shell-script"},
		{"2026-03-23-use-go-over-shell-script.md", "use-go-over-shell-script"},
		{"0003-madr-format-compatibility.md", "madr-format-compatibility"},
		{"2026-01-15-use-postgresql.md", "use-postgresql"},
		{"not-a-doc.md", ""},
		{"README.md", ""},
	}
	for _, tt := range tests {
		got := slugFromFilename(tt.name)
		if got != tt.want {
			t.Errorf("slugFromFilename(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	path, err := Create(dir, "Use PostgreSQL", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	base := filepath.Base(path)
	if !strings.HasSuffix(base, "-use-postgresql.md") {
		t.Errorf("filename %q should end with -use-postgresql.md", base)
	}
	if !strings.HasPrefix(base, "2") { // starts with year
		t.Errorf("filename %q should start with YYYY-", base)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	body := string(content)
	if !contains(body, "# Use PostgreSQL") {
		t.Errorf("file missing title header, got:\n%s", body)
	}
	if !contains(body, "Status: Draft") {
		t.Errorf("file missing status, got:\n%s", body)
	}
}

func TestCreateDesign(t *testing.T) {
	dir := t.TempDir()
	path, err := CreateDesign(dir, "Connection Pool Design", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	base := filepath.Base(path)
	if !strings.HasSuffix(base, "-connection-pool-design.md") {
		t.Errorf("filename %q should end with -connection-pool-design.md", base)
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

func TestParse_NewFormat(t *testing.T) {
	content := `# Use MADR Format

- Date: 2026-03-12
- Status: Active
- Author: masahiro.kasatani

## Context

Some context here.
`
	path := filepath.Join(t.TempDir(), "2026-03-12-use-madr-format.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID != 0 {
		t.Errorf("ID = %d, want 0 for new format", d.ID)
	}
	if d.Slug != "use-madr-format" {
		t.Errorf("Slug = %q, want %q", d.Slug, "use-madr-format")
	}
	if d.Title != "Use MADR Format" {
		t.Errorf("Title = %q, want %q", d.Title, "Use MADR Format")
	}
	if d.Date != "2026-03-12" {
		t.Errorf("Date = %q, want %q", d.Date, "2026-03-12")
	}
	if d.Status != "Active" {
		t.Errorf("Status = %q, want %q", d.Status, "Active")
	}
	if d.Author != "masahiro.kasatani" {
		t.Errorf("Author = %q, want %q", d.Author, "masahiro.kasatani")
	}
}

func TestParse_LegacyFormat(t *testing.T) {
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
	if d.Slug != "use-madr-format" {
		t.Errorf("Slug = %q, want %q", d.Slug, "use-madr-format")
	}
	if d.Title != "Use MADR Format" {
		t.Errorf("Title = %q, want %q", d.Title, "Use MADR Format")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"2026-01-01-use-go.md":    "# Use Go\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n",
		"2026-03-01-use-madr.md":  "# Use MADR\n\n- Date: 2026-03-01\n- Status: Active\n- Author: alice\n",
		"2026-02-01-use-cobra.md": "# Use Cobra\n\n- Date: 2026-02-01\n- Status: Active\n- Author: alice\n",
		"not-a-decision.txt":      "ignored",
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
	// Expect descending order by date: 2026-03, 2026-02, 2026-01
	wantSlugs := []string{"use-madr", "use-cobra", "use-go"}
	for i, want := range wantSlugs {
		if decisions[i].Slug != want {
			t.Errorf("decisions[%d].Slug = %q, want %q", i, decisions[i].Slug, want)
		}
	}
}

func TestFindBySlug(t *testing.T) {
	dir := t.TempDir()
	content := "# Use Cobra\n\n- Date: 2026-03-12\n- Status: Active\n- Author: alice\n"
	if err := os.WriteFile(filepath.Join(dir, "2026-03-12-use-cobra.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := FindBySlug(dir, "use-cobra")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Slug != "use-cobra" {
		t.Errorf("Slug = %q, want %q", d.Slug, "use-cobra")
	}
	if d.Title != "Use Cobra" {
		t.Errorf("Title = %q, want %q", d.Title, "Use Cobra")
	}
}

func TestFindBySlug_Legacy(t *testing.T) {
	dir := t.TempDir()
	content := "# 0002: Use Cobra\n\n- Date: 2026-03-12\n- Status: Accepted\n- Author: alice\n"
	if err := os.WriteFile(filepath.Join(dir, "0002-use-cobra.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := FindBySlug(dir, "use-cobra")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Slug != "use-cobra" {
		t.Errorf("Slug = %q, want %q", d.Slug, "use-cobra")
	}
	if d.ID != 2 {
		t.Errorf("ID = %d, want 2", d.ID)
	}
}

func TestFindBySlug_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindBySlug(dir, "nonexistent-slug")
	if err == nil {
		t.Error("expected error for missing slug, got nil")
	}
}

func TestList_Recursive(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "ja")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		filepath.Join(dir, "2026-01-01-use-go.md"):     "# Use Go\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n",
		filepath.Join(subdir, "2026-01-01-use-go.md"):  "# Use Go (ja)\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n",
		filepath.Join(subdir, "2026-03-01-use-madr.md"): "# Use MADR (ja)\n\n- Date: 2026-03-01\n- Status: Active\n- Author: alice\n",
	}
	for path, body := range files {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	decisions, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 3 {
		t.Fatalf("len = %d, want 3 (including subdirectory files)", len(decisions))
	}
}

func TestFindBySlug_Recursive(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "ja")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "# Use Cobra (ja)\n\n- Date: 2026-03-12\n- Status: Active\n- Author: alice\n"
	if err := os.WriteFile(filepath.Join(subdir, "2026-03-12-use-cobra.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := FindBySlug(dir, "use-cobra")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Slug != "use-cobra" {
		t.Errorf("Slug = %q, want %q", d.Slug, "use-cobra")
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
