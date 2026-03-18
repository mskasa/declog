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

	for _, want := range []string{"ADR Check", "pull_request", "[skip-adr]", "docs/decisions/"} {
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

	for _, want := range []string{"ADR Audit", "schedule", "cron", "kizami audit", "[ADR Audit]"} {
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
	configDir := t.TempDir()
	var out bytes.Buffer

	init_ := &Initializer{
		Root:      root,
		Input:     strings.NewReader("n\nn\nn\nn\n"), // workflow=n, hook=n, audit=n, promote=n
		Output:    &out,
		ConfigDir: configDir,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config.toml not created: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}
	for _, want := range []string{"[ai]", "claude-sonnet-4-20250514", "[decisions]", "[review]", "[editor]"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("config.toml missing %q", want)
		}
	}

	output := out.String()
	if !strings.Contains(output, "✅ Created ~/.config/kizami/config.toml") {
		t.Errorf("expected creation message, got: %s", output)
	}
}

func TestRun_ConfigAlreadyExists(t *testing.T) {
	root := t.TempDir()
	configDir := t.TempDir()

	// Pre-create the config file.
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	init_ := &Initializer{
		Root:      root,
		Input:     strings.NewReader("n\nn\nn\nn\n"),
		Output:    &out,
		ConfigDir: configDir,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "~/.config/kizami/config.toml already exists. Skipping.") {
		t.Errorf("expected skip message, got: %s", output)
	}

	// Existing file must not be overwritten.
	content, _ := os.ReadFile(configPath)
	if string(content) != "existing" {
		t.Errorf("existing config.toml was overwritten")
	}
}

func TestRun_CreatesConfigDir(t *testing.T) {
	root := t.TempDir()
	// Use a subdirectory that does not yet exist.
	configDir := filepath.Join(t.TempDir(), "kizami")
	var out bytes.Buffer

	init_ := &Initializer{
		Root:      root,
		Input:     strings.NewReader("n\nn\nn\nn\n"),
		Output:    &out,
		ConfigDir: configDir,
	}

	if err := init_.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config.toml not created when dir was missing: %v", err)
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
