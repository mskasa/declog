package initializer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_FreshInit_WithWorkflow(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("y\nn\n"), // workflow=y, hook=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	decisionsDir := filepath.Join(root, "docs", "decisions")
	if _, err := os.Stat(decisionsDir); err != nil {
		t.Errorf("docs/decisions/ not created: %v", err)
	}

	workflowPath := filepath.Join(root, ".github", "workflows", "adr-check.yml")
	if _, err := os.Stat(workflowPath); err != nil {
		t.Errorf("adr-check.yml not created: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created docs/decisions/") {
		t.Errorf("expected creation message, got: %s", output)
	}
	if !strings.Contains(output, "✅ Created .github/workflows/adr-check.yml") {
		t.Errorf("expected workflow creation message, got: %s", output)
	}
	if !strings.Contains(output, "Done!") {
		t.Errorf("expected done message, got: %s", output)
	}
}

func TestRun_FreshInit_SkipWorkflow(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\n"), // workflow=n, hook=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	workflowPath := filepath.Join(root, ".github", "workflows", "adr-check.yml")
	if _, err := os.Stat(workflowPath); err == nil {
		t.Errorf("adr-check.yml should not be created when answering n")
	}
}

func TestRun_DecisionsDirAlreadyExists(t *testing.T) {
	root := t.TempDir()

	decisionsDir := filepath.Join(root, "docs", "decisions")
	if err := os.MkdirAll(decisionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\n"), // workflow=n, hook=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "already exists. Skipping.") {
		t.Errorf("expected skip message, got: %s", output)
	}
	if strings.Contains(output, "✅ Created docs/decisions/") {
		t.Errorf("should not print creation message when dir already exists")
	}
}

func TestRun_WorkflowContent(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("y\nn\n"), // workflow=y, hook=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	workflowPath := filepath.Join(root, ".github", "workflows", "adr-check.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("reading adr-check.yml: %v", err)
	}

	for _, want := range []string{"ADR Check", "pull_request", "[skip-adr]", "docs/decisions/"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("adr-check.yml missing %q", want)
		}
	}
}

func TestRun_WithHook(t *testing.T) {
	root := t.TempDir()
	// Create .git/hooks dir so InstallHook can write there.
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\ny\n"), // workflow=n, hook=y
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err != nil {
		t.Errorf("pre-commit hook not created: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created .git/hooks/pre-commit") {
		t.Errorf("expected hook creation message, got: %s", output)
	}
}
