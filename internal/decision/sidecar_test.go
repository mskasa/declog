package decision

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSidecar(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestParseSidecar(t *testing.T) {
	dir := t.TempDir()
	path := writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix for user flow
date: 2026-04-17
author: alice
related:
  - tests/user_flow_test.go
  - data/schema.json
`)

	d, err := ParseSidecar(path)
	if err != nil {
		t.Fatalf("ParseSidecar: %v", err)
	}
	if d.Title != "Test matrix for user flow" {
		t.Errorf("Title = %q", d.Title)
	}
	if d.Date != "2026-04-17" {
		t.Errorf("Date = %q", d.Date)
	}
	if d.Author != "alice" {
		t.Errorf("Author = %q", d.Author)
	}
	if d.Status != "Active" {
		t.Errorf("Status = %q, want Active", d.Status)
	}
	if d.Slug != "test_matrix.csv" {
		t.Errorf("Slug = %q, want test_matrix.csv", d.Slug)
	}
}

func TestParseSidecarRelatedFiles(t *testing.T) {
	dir := t.TempDir()
	path := writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix
date: 2026-04-17
author: alice
related:
  - tests/user_flow_test.go
  - data/schema.json
`)

	files, err := ParseSidecarRelatedFiles(path)
	if err != nil {
		t.Fatalf("ParseSidecarRelatedFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	if files[0] != "tests/user_flow_test.go" {
		t.Errorf("files[0] = %q", files[0])
	}
	if files[1] != "data/schema.json" {
		t.Errorf("files[1] = %q", files[1])
	}
}

func TestParseSidecarRelatedFiles_WithBackticks(t *testing.T) {
	dir := t.TempDir()
	path := writeSidecar(t, dir, "spec.yaml.kizami",
		"title: API spec\ndate: 2026-04-17\nauthor: alice\nrelated:\n  - `internal/handler.go`\n")

	files, err := ParseSidecarRelatedFiles(path)
	if err != nil {
		t.Fatalf("ParseSidecarRelatedFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "internal/handler.go" {
		t.Errorf("got %v", files)
	}
}

func TestIsSidecarFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"test_matrix.csv.kizami", true},
		{"spec.yaml.kizami", true},
		{"README.md", false},
		{"doc.kizami.md", false},
	}
	for _, c := range cases {
		if got := IsSidecarFile(c.path); got != c.want {
			t.Errorf("IsSidecarFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestList_IncludesSidecar(t *testing.T) {
	dir := t.TempDir()
	writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix
date: 2026-04-17
author: alice
related:
  - tests/user_flow_test.go
`)

	decisions, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("got %d decisions, want 1", len(decisions))
	}
	if decisions[0].Slug != "test_matrix.csv" {
		t.Errorf("Slug = %q", decisions[0].Slug)
	}
	if decisions[0].Status != "Active" {
		t.Errorf("Status = %q", decisions[0].Status)
	}
}

func TestFindBySlug_Sidecar(t *testing.T) {
	dir := t.TempDir()
	writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix
date: 2026-04-17
author: alice
related:
  - tests/user_flow_test.go
`)

	d, err := FindBySlug(dir, "test_matrix.csv")
	if err != nil {
		t.Fatalf("FindBySlug: %v", err)
	}
	if d.Title != "Test matrix" {
		t.Errorf("Title = %q", d.Title)
	}
}

func TestParseRelatedFiles_Sidecar(t *testing.T) {
	dir := t.TempDir()
	path := writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix
date: 2026-04-17
author: alice
related:
  - tests/user_flow_test.go
`)

	files, err := ParseRelatedFiles(path)
	if err != nil {
		t.Fatalf("ParseRelatedFiles: %v", err)
	}
	if len(files) != 1 || files[0] != "tests/user_flow_test.go" {
		t.Errorf("got %v", files)
	}
}

func TestAudit_Sidecar_MissingFile(t *testing.T) {
	dir := t.TempDir()
	writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix
date: 2026-04-17
author: alice
related:
  - tests/missing_test.go
`)

	results, err := Audit(dir, dir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if len(results[0].MissingFiles) != 1 || results[0].MissingFiles[0] != "tests/missing_test.go" {
		t.Errorf("MissingFiles = %v", results[0].MissingFiles)
	}
}

func TestAudit_Sidecar_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "test_matrix.csv")
	if err := os.WriteFile(existing, []byte("a,b\n1,2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSidecar(t, dir, "test_matrix.csv.kizami", `title: Test matrix
date: 2026-04-17
author: alice
related:
  - test_matrix.csv
`)

	results, err := Audit(dir, dir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d audit results, want 0", len(results))
	}
}
