package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	kizamicontext "github.com/mskasa/kizami/internal/context"
)

func TestContextCmd_NoResults(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	out, err := executeCmd(t, "context", "internal/some/file.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No decisions govern") {
		t.Errorf("expected 'No decisions govern', got: %q", out)
	}
	if !strings.Contains(out, "internal/some/file.go") {
		t.Errorf("expected the file listed as unmatched, got: %q", out)
	}
}

func TestContextCmd_WithResults(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use Go", "Active")
	appendRelatedFile(t, path, "internal/search/search.go")
	writeRelatedFile(t, root, "internal/search/search.go")
	setTestRoot(t, root)

	out, err := executeCmd(t, "context", "internal/search/search.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") {
		t.Errorf("expected decision title in output, got: %q", out)
	}
	if !strings.Contains(out, "Drift: ok") {
		t.Errorf("expected 'Drift: ok', got: %q", out)
	}
}

func TestContextCmd_JSON(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	path := seedDecision(t, dir, 1, "Use Go", "Active")
	appendRelatedFile(t, path, "internal/search/search.go")
	setTestRoot(t, root)

	out, err := executeCmd(t, "context", "internal/search/search.go", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result kizamicontext.Result
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON, got error %v for output: %q", err, out)
	}
	if len(result.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(result.Decisions))
	}
	if result.Decisions[0].Slug != "use-go" {
		t.Errorf("unexpected slug: %s", result.Decisions[0].Slug)
	}
}

func TestContextCmd_MultipleDirs(t *testing.T) {
	root := newTestRepo(t)
	seedMultiDirConfig(t, root)
	decPath := seedDecision(t, decisionsPath(root), 1, "Use Go", "Active")
	appendRelatedFile(t, decPath, "cmd/root.go")
	designDocPath := seedDecision(t, designPath(root), 1, "Cache Design", "Active")
	appendRelatedFile(t, designDocPath, "internal/search/search.go")
	setTestRoot(t, root)

	out, err := executeCmd(t, "context", "cmd/root.go", "internal/search/search.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Use Go") || !strings.Contains(out, "Cache Design") {
		t.Errorf("expected both decisions in output, got: %q", out)
	}
}
