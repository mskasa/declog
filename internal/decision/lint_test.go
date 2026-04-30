package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLintDoc(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func validDoc(relatedFiles ...string) string {
	var b strings.Builder
	b.WriteString("# Test Decision\n\n- Date: 2026-01-01\n- Status: Active\n- Author: test\n\n## Context\n\nContext.\n\n## Related Files\n\n")
	for _, f := range relatedFiles {
		b.WriteString("- " + f + "\n")
	}
	return b.String()
}

func findIssue(issues []*LintIssue, substr string) *LintIssue {
	for _, i := range issues {
		if strings.Contains(i.Message, substr) {
			return i
		}
	}
	return nil
}

func TestLint_NoIssues(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	existing := filepath.Join(repoRoot, "internal", "foo.go")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeLintDoc(t, filepath.Join(dir, "2026-01-01-test.md"), validDoc("internal/foo.go"))

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}
}

func TestLint_MissingStatus(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	content := "# Test\n\n- Date: 2026-01-01\n- Author: test\n\n## Related Files\n\n- internal/foo.go\n"
	writeLintDoc(t, filepath.Join(dir, "2026-01-01-test.md"), content)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "Status") == nil {
		t.Error("expected missing Status issue")
	}
}

func TestLint_MalformedDate(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	existing := filepath.Join(repoRoot, "internal", "foo.go")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	content := "# Test\n\n- Date: January 1, 2026\n- Status: Active\n- Author: test\n\n## Related Files\n\n- internal/foo.go\n"
	writeLintDoc(t, filepath.Join(dir, "2026-01-01-test.md"), content)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "YYYY-MM-DD") == nil {
		t.Error("expected malformed date issue")
	}
}

func TestLint_EmptyRelatedFiles(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	content := "# Test\n\n- Date: 2026-01-01\n- Status: Active\n- Author: test\n\n## Related Files\n\n<!-- no files listed -->\n"
	writeLintDoc(t, filepath.Join(dir, "2026-01-01-test.md"), content)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "Related Files") == nil {
		t.Error("expected empty Related Files issue")
	}
}

func TestLint_UnresolvablePath(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	writeLintDoc(t, filepath.Join(dir, "2026-01-01-test.md"), validDoc("internal/missing.go"))

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "internal/missing.go") == nil {
		t.Error("expected unresolvable path issue")
	}
}

func TestLint_UnresolvablePath_AllStatuses(t *testing.T) {
	// Unlike audit (Active only), lint checks unresolvable paths for all statuses.
	dir := t.TempDir()
	repoRoot := t.TempDir()

	for _, status := range []string{"Draft", "Inactive", "Superseded by other"} {
		name := "2026-01-01-" + strings.ToLower(strings.Fields(status)[0]) + ".md"
		content := "# Test\n\n- Date: 2026-01-01\n- Status: " + status + "\n- Author: test\n\n## Related Files\n\n- internal/missing.go\n"
		writeLintDoc(t, filepath.Join(dir, name), content)
	}

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// One unresolvable path issue per document (3 docs).
	count := 0
	for _, i := range issues {
		if strings.Contains(i.Message, "internal/missing.go") {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 unresolvable path issues (one per status), got %d", count)
	}
}

func TestLint_Sidecar_NoIssues(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	existing := filepath.Join(repoRoot, "internal", "foo.go")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	sidecar := "title: My artifact\ndate: 2026-01-01\nauthor: test\nrelated:\n  - internal/foo.go\n"
	writeLintDoc(t, filepath.Join(dir, "artifact.csv.kizami"), sidecar)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got: %v", issues)
	}
}

func TestLint_Sidecar_MissingTitle(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	sidecar := "date: 2026-01-01\nauthor: test\nrelated:\n  - internal/foo.go\n"
	writeLintDoc(t, filepath.Join(dir, "artifact.csv.kizami"), sidecar)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "title") == nil {
		t.Error("expected missing title issue for sidecar")
	}
}

func TestLint_Sidecar_EmptyRelated(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	sidecar := "title: My artifact\ndate: 2026-01-01\nauthor: test\n"
	writeLintDoc(t, filepath.Join(dir, "artifact.csv.kizami"), sidecar)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "related") == nil {
		t.Error("expected empty related issue for sidecar")
	}
}

func TestLint_Sidecar_UnresolvablePath(t *testing.T) {
	dir := t.TempDir()
	repoRoot := t.TempDir()

	sidecar := "title: My artifact\ndate: 2026-01-01\nauthor: test\nrelated:\n  - internal/missing.go\n"
	writeLintDoc(t, filepath.Join(dir, "artifact.csv.kizami"), sidecar)

	issues, err := Lint(dir, repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findIssue(issues, "internal/missing.go") == nil {
		t.Error("expected unresolvable path issue for sidecar")
	}
}
