package initializer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitHooksDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestInstallHook_CreatesFile(t *testing.T) {
	root := setupGitHooksDir(t)
	var out bytes.Buffer

	if err := InstallHook(root, &out); err != nil {
		t.Fatalf("InstallHook() error: %v", err)
	}

	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err != nil {
		t.Errorf("pre-commit not created: %v", err)
	}

	if !strings.Contains(out.String(), "✅ Created .git/hooks/pre-commit") {
		t.Errorf("expected creation message, got: %s", out.String())
	}
}

func TestInstallHook_FileIsExecutable(t *testing.T) {
	root := setupGitHooksDir(t)
	var out bytes.Buffer

	if err := InstallHook(root, &out); err != nil {
		t.Fatalf("InstallHook() error: %v", err)
	}

	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("stat pre-commit: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("pre-commit is not executable, mode: %v", info.Mode())
	}
}

func TestInstallHook_HookContent(t *testing.T) {
	root := setupGitHooksDir(t)
	var out bytes.Buffer

	if err := InstallHook(root, &out); err != nil {
		t.Fatalf("InstallHook() error: %v", err)
	}

	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("reading pre-commit: %v", err)
	}

	for _, want := range []string{"#!/bin/sh", "docs/decisions/", "why log"} {
		if !strings.Contains(string(content), want) {
			t.Errorf("hook missing %q", want)
		}
	}
}

func TestInstallHook_AlreadyExists(t *testing.T) {
	root := setupGitHooksDir(t)

	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho existing\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := InstallHook(root, &out); err != nil {
		t.Fatalf("InstallHook() error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "⚠️  pre-commit hook already exists") {
		t.Errorf("expected warning about existing hook, got: %s", output)
	}
	// Existing hook must not be overwritten.
	content, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(content), "existing") {
		t.Errorf("existing hook was overwritten")
	}
}

func TestInstallHook_CreatesMissingHooksDir(t *testing.T) {
	root := t.TempDir()
	// Only create .git dir, not .git/hooks.
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := InstallHook(root, &out); err != nil {
		t.Fatalf("InstallHook() error: %v", err)
	}

	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); err != nil {
		t.Errorf("pre-commit not created: %v", err)
	}
}
