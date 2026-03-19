package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mskasa/kizami/internal/decision"
)

// newTestRepo creates a temporary directory initialized as a git repository.
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if err := exec.Command("git", append([]string{"-C", dir}, args...)...).Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	return dir
}

// setTestRoot overrides gitRepoRootFn to return the given root for the duration of the test.
func setTestRoot(t *testing.T, root string) {
	t.Helper()
	orig := gitRepoRootFn
	gitRepoRootFn = func() (string, error) { return root, nil }
	t.Cleanup(func() { gitRepoRootFn = orig })
}

// executeCmd runs a cobra command with the given args, capturing os.Stdout and returning it.
func executeCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Capture os.Stdout via a pipe (commands write directly to os.Stdout).
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	rootCmd.SetArgs(args)
	_, execErr := rootCmd.ExecuteC()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	r.Close()

	// Reset args to avoid leaking state between tests.
	rootCmd.SetArgs(nil)
	return buf.String(), execErr
}

// seedDecision writes a minimal decision file into dir and returns its path.
func seedDecision(t *testing.T, dir string, id int, title, status string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path, err := decision.Create(dir, title, 0)
	if err != nil {
		t.Fatalf("decision.Create: %v", err)
	}
	if status != "Draft" {
		if err := decision.UpdateStatus(path, status, 0); err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
	}
	_ = id
	return path
}

// decisionsPath returns the default decisions directory for a given repo root.
func decisionsPath(root string) string {
	return filepath.Join(root, "docs", "decisions")
}

// appendRelatedFile appends a file entry to the Related Files section of a decision.
func appendRelatedFile(t *testing.T, decisionPath, file string) {
	t.Helper()
	content, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	updated := string(content) + "\n- " + file + "\n"
	if err := os.WriteFile(decisionPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
