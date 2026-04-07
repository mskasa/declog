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
		Input:  strings.NewReader("y\nn\nn\nn\n"), // workflow=y, hook=n, audit=n, promote=n
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
		Input:  strings.NewReader("n\nn\nn\nn\n"), // workflow=n, hook=n, audit=n, promote=n
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
		Input:  strings.NewReader("n\nn\nn\nn\n"), // workflow=n, hook=n, audit=n, promote=n
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
		Input:  strings.NewReader("y\nn\nn\nn\n"), // workflow=y, hook=n, audit=n, promote=n
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

	for _, want := range []string{"Document Check", "pull_request", "[skip-doc]", "docs/decisions/", "docs/design/"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("adr-check.yml missing %q", want)
		}
	}
}

func TestRun_WithAuditWorkflow(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\ny\nn\n"), // workflow=n, hook=n, audit=y, promote=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	auditPath := filepath.Join(root, ".github", "workflows", "adr-audit.yml")
	if _, err := os.Stat(auditPath); err != nil {
		t.Errorf("adr-audit.yml not created: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created .github/workflows/adr-audit.yml") {
		t.Errorf("expected audit workflow creation message, got: %s", output)
	}
}

func TestRun_AuditWorkflowAlreadyExists(t *testing.T) {
	root := t.TempDir()

	workflowDir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	auditPath := filepath.Join(workflowDir, "adr-audit.yml")
	if err := os.WriteFile(auditPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\ny\nn\n"), // workflow=n, hook=n, audit=y, promote=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "adr-audit.yml already exists. Skipping.") {
		t.Errorf("expected skip message, got: %s", output)
	}

	// Existing file must not be overwritten.
	content, _ := os.ReadFile(auditPath)
	if string(content) != "existing" {
		t.Errorf("existing adr-audit.yml was overwritten")
	}
}

func TestRun_AuditWorkflowContent(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\ny\nn\n"), // workflow=n, hook=n, audit=y, promote=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	auditPath := filepath.Join(root, ".github", "workflows", "adr-audit.yml")
	content, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("reading adr-audit.yml: %v", err)
	}

	for _, want := range []string{"kizami Audit", "schedule", "cron", "kizami audit", "[kizami audit]"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("adr-audit.yml missing %q", want)
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
		Input:  strings.NewReader("n\ny\nn\nn\n"), // workflow=n, hook=y, audit=n, promote=n
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

func TestRun_CreatesConfig(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\nn\n"), // workflow=n, hook=n, audit=n, promote=n
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	configPath := filepath.Join(root, "kizami.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("kizami.toml not created: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading kizami.toml: %v", err)
	}
	for _, want := range []string{"[ai]", "claude-sonnet-4-20250514", "[decisions]", "[audit]", "[review]", "[editor]"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("kizami.toml missing %q", want)
		}
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created kizami.toml") {
		t.Errorf("expected creation message, got: %s", output)
	}
}

func TestRun_ConfigAlreadyExists(t *testing.T) {
	root := t.TempDir()

	// Pre-create the config file.
	configPath := filepath.Join(root, "kizami.toml")
	if err := os.WriteFile(configPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\nn\n"),
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "kizami.toml already exists. Skipping.") {
		t.Errorf("expected skip message, got: %s", output)
	}

	// Existing file must not be overwritten.
	content, _ := os.ReadFile(configPath)
	if string(content) != "existing" {
		t.Errorf("existing kizami.toml was overwritten")
	}
}

func TestRun_WithPromoteWorkflow(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\ny\n"), // workflow=n, hook=n, audit=n, promote=y
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	promotePath := filepath.Join(root, ".github", "workflows", "kizami-promote.yml")
	if _, err := os.Stat(promotePath); err != nil {
		t.Errorf("kizami-promote.yml not created: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created .github/workflows/kizami-promote.yml") {
		t.Errorf("expected promote workflow creation message, got: %s", output)
	}
}

func TestRun_PromoteWorkflowContent(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\ny\n"), // workflow=n, hook=n, audit=n, promote=y
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	promotePath := filepath.Join(root, ".github", "workflows", "kizami-promote.yml")
	content, err := os.ReadFile(promotePath)
	if err != nil {
		t.Fatalf("reading kizami-promote.yml: %v", err)
	}

	for _, want := range []string{"Promote Draft Documents", "Status: Draft", "Status: Active", "docs/decisions", "docs/design", "[skip ci]"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("kizami-promote.yml missing %q", want)
		}
	}
}

func TestRun_PromoteWorkflowAlreadyExists(t *testing.T) {
	root := t.TempDir()

	workflowDir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	promotePath := filepath.Join(workflowDir, "kizami-promote.yml")
	if err := os.WriteFile(promotePath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\ny\n"), // workflow=n, hook=n, audit=n, promote=y
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "kizami-promote.yml already exists. Skipping.") {
		t.Errorf("expected skip message, got: %s", output)
	}

	content, _ := os.ReadFile(promotePath)
	if string(content) != "existing" {
		t.Errorf("existing kizami-promote.yml was overwritten")
	}
}

func TestRun_YesAll(t *testing.T) {
	root := t.TempDir()
	// Create .git/hooks so InstallHook works.
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader(""), // no input — should not be needed
		Output: &out,
		YesAll: true,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// All workflows and hook must be created.
	for _, path := range []string{
		filepath.Join(root, ".github", "workflows", "adr-check.yml"),
		filepath.Join(root, ".github", "workflows", "adr-audit.yml"),
		filepath.Join(root, ".github", "workflows", "kizami-promote.yml"),
		filepath.Join(root, ".git", "hooks", "pre-commit"),
		filepath.Join(root, "kizami.toml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to be created, got: %v", path, err)
		}
	}
}

func TestRun_YesAll_NoInputRead(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}

	// errReader simulates a closed stdin that returns an error on read.
	type errReader struct{}
	_ = errReader{}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader(""), // EOF immediately
		Output: &out,
		YesAll: true,
	}

	// Must succeed even though stdin is empty.
	if err := init_.Run(); err != nil {
		t.Fatalf("Run() with YesAll and empty stdin should not fail: %v", err)
	}
}

func TestRun_CreatesDesignDir(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer

	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\nn\n"),
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	designDir := filepath.Join(root, "docs", "design")
	if _, err := os.Stat(designDir); err != nil {
		t.Errorf("docs/design/ not created: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created docs/design/") {
		t.Errorf("expected design dir creation message, got: %s", output)
	}
}

func TestRun_DesignDirAlreadyExists(t *testing.T) {
	root := t.TempDir()

	designDir := filepath.Join(root, "docs", "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\nn\n"),
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "docs/design/ already exists. Skipping.") {
		t.Errorf("expected skip message, got: %s", output)
	}
}

func TestRun_PromoteWorkflowContainsBranch(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	init_ := &Initializer{
		Root:   root,
		Input:  strings.NewReader("n\nn\nn\ny\n"), // workflow=n, hook=n, audit=n, promote=y
		Output: &out,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	promotePath := filepath.Join(root, ".github", "workflows", "kizami-promote.yml")
	content, err := os.ReadFile(promotePath)
	if err != nil {
		t.Fatalf("reading kizami-promote.yml: %v", err)
	}

	// The placeholder must be replaced; the literal string must not remain.
	if strings.Contains(string(content), "{{DEFAULT_BRANCH}}") {
		t.Errorf("kizami-promote.yml still contains unreplaced placeholder {{DEFAULT_BRANCH}}")
	}
	// Must have a concrete branch name in the push trigger.
	if !strings.Contains(string(content), "branches:") {
		t.Errorf("kizami-promote.yml missing branches: section")
	}
}
