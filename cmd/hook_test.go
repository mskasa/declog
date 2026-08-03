package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func stageFile(t *testing.T, root, rel string) {
	t.Helper()
	if err := exec.Command("git", "-C", root, "add", rel).Run(); err != nil {
		t.Fatalf("git add %s: %v", rel, err)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHookPreCommit_NoStagedFiles(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output for empty staging area, got: %q", out)
	}
}

func TestHookPreCommit_StagedFileMatchesRelatedFiles(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	setTestRoot(t, root)

	// Create a document with a related file, then stage the related file.
	docPath := seedDecision(t, dir, 1, "Use connection pooling", "Active")
	appendRelatedFile(t, docPath, "internal/db/db.go")

	writeFile(t, root, "internal/db/db.go", "package db\n")
	stageFile(t, root, "internal/db/db.go")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "use-connection-pooling") {
		t.Errorf("expected slug in output, got: %q", out)
	}
	if !strings.Contains(out, "⚠️") {
		t.Errorf("expected warning in output, got: %q", out)
	}
}

func TestHookPreCommit_DocumentStaged_NoWarning(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	setTestRoot(t, root)

	docPath := seedDecision(t, dir, 1, "Use connection pooling", "Active")
	appendRelatedFile(t, docPath, "internal/db/db.go")

	writeFile(t, root, "internal/db/db.go", "package db\n")
	stageFile(t, root, "internal/db/db.go")
	stageFile(t, root, filepath.ToSlash(func() string {
		rel, _ := filepath.Rel(root, docPath)
		return rel
	}()))

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning when document is staged, got: %q", out)
	}
}

func TestHookPreCommit_NoMatch_SuggestsNewDocument(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	// Stage a non-MD file with no related documents.
	writeFile(t, root, "internal/db/db.go", "package db\n")
	stageFile(t, root, "internal/db/db.go")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kizami adr") {
		t.Errorf("expected 'kizami adr' suggestion in output, got: %q", out)
	}
}

func TestHookPreCommit_OnlyMDFiles_NoWarning(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	writeFile(t, root, "README.md", "# readme\n")
	stageFile(t, root, "README.md")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning for MD-only commit, got: %q", out)
	}
}

func TestHookPreCommit_NonASCIIFilename_NoWarning(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	// Filenames with non-ASCII characters (e.g. Japanese) are quoted by git
	// when core.quotepath=true (the default). This test ensures the hook does
	// not produce a false warning for such files.
	writeFile(t, root, "docs/日本語ファイル名.md", "# test\n")
	stageFile(t, root, "docs/日本語ファイル名.md")

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning for non-ASCII MD filename, got: %q", out)
	}
}

// buildPreToolUseEvent JSON-encodes a PreToolUse event, which matters on Windows: cwd/
// file_path can contain backslashes, and naive string concatenation into a JSON literal
// (rather than proper marshaling) produces invalid JSON there even though it happens to
// look fine on POSIX paths.
func buildPreToolUseEvent(t *testing.T, cwd, filePath string) string {
	t.Helper()
	event := map[string]any{
		"cwd": cwd,
		"tool_input": map[string]any{
			"file_path": filePath,
		},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestHookPreToolUse_NoMatch_NoOutput(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	event := buildPreToolUseEvent(t, root, "internal/other/file.go")
	out, err := executeCmdWithStdin(t, event, "hook", "pre-tool-use")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output when nothing governs the file, got: %q", out)
	}
}

func TestHookPreToolUse_Match_InjectsAdditionalContext(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use connection pooling", "Active")
	appendRelatedFile(t, path, "internal/db/db.go")
	setTestRoot(t, root)

	event := buildPreToolUseEvent(t, root, "internal/db/db.go")
	out, err := executeCmdWithStdin(t, event, "hook", "pre-tool-use")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("expected valid JSON, got error %v for output: %q", err, out)
	}
	if resp.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Errorf("unexpected hookEventName: %s", resp.HookSpecificOutput.HookEventName)
	}
	if !strings.Contains(resp.HookSpecificOutput.AdditionalContext, "use-connection-pooling") {
		t.Errorf("expected slug in additionalContext, got: %q", resp.HookSpecificOutput.AdditionalContext)
	}
}

func TestHookPreToolUse_AbsoluteFilePath(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use connection pooling", "Active")
	appendRelatedFile(t, path, "internal/db/db.go")
	setTestRoot(t, root)

	absPath := filepath.Join(root, "internal/db/db.go")
	event := buildPreToolUseEvent(t, root, absPath)
	out, err := executeCmdWithStdin(t, event, "hook", "pre-tool-use")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "use-connection-pooling") {
		t.Errorf("expected slug in output for absolute file_path, got: %q", out)
	}
}

func TestHookPreToolUse_MissingFilePath_NoOutput(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	out, err := executeCmdWithStdin(t, `{"tool_input":{}}`, "hook", "pre-tool-use")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output for a missing file_path, got: %q", out)
	}
}

func TestHookPreToolUse_MalformedJSON_NoErrorNoOutput(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)

	out, err := executeCmdWithStdin(t, `not json`, "hook", "pre-tool-use")
	if err != nil {
		t.Fatalf("expected the hook to never error out, got: %v", err)
	}
	if out != "" {
		t.Errorf("expected no output for malformed input, got: %q", out)
	}
}

func TestHookPreCommit_DocInDecisionsDirStaged_NoCheck2(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	setTestRoot(t, root)

	// Stage a document (not a source file), no source changes.
	docPath := seedDecision(t, dir, 1, "Use Go", "Active")
	stageFile(t, root, filepath.ToSlash(func() string {
		rel, _ := filepath.Rel(root, docPath)
		return rel
	}()))

	out, err := executeCmd(t, "hook", "pre-commit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("expected no warning when only document is staged, got: %q", out)
	}
}
