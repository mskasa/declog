package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusCmd_InvalidStatus(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Active")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "use-go", "unknown")
	if err == nil {
		t.Fatal("expected error for unknown status")
	}
}

func TestStatusCmd_ValidUpdate(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Draft")
	setTestRoot(t, root)

	out, err := executeCmd(t, "status", "use-go", "accepted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "use-go") {
		t.Errorf("expected slug in output, got: %q", out)
	}
	if !strings.Contains(out, "Accepted") {
		t.Errorf("expected 'Accepted' in output, got: %q", out)
	}
}

func TestStatusCmd_ByFlagWithoutSuperseded(t *testing.T) {
	t.Cleanup(func() { supersededBySlug = "" })
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Draft")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "use-go", "accepted", "--by", "use-another")
	if err == nil {
		t.Fatal("expected error when --by is used without superseded status")
	}
}

func TestStatusCmd_SlugCollision_Errors(t *testing.T) {
	root := newTestRepo(t)
	seedMultiDirConfig(t, root)
	seedDecision(t, decisionsPath(root), 1, "Use Go", "Draft")
	seedDecision(t, designPath(root), 1, "Use Go", "Draft")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "use-go", "accepted")
	if err == nil {
		t.Fatal("expected error when slug is ambiguous across directories")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Errorf("expected 'multiple' in error message, got: %v", err)
	}
}

// TestStatusCmd_NestedLanguageVariant_ErrorsAmbiguous is a regression test: before
// decision.FindAllBySlug, this ambiguity went undetected for EN/JA pairs nested under the
// same configured directory (docs/decisions/ja/), and `kizami status` would silently update
// only the EN document, leaving the JA one to quietly diverge:
// docs/decisions/2026-08-03-findbyslug-recursive-language-variants.md
func TestStatusCmd_NestedLanguageVariant_ErrorsAmbiguous(t *testing.T) {
	root := newTestRepo(t)
	dir := decisionsPath(root)
	seedDecision(t, dir, 1, "Use Go", "Draft")
	seedDecision(t, filepath.Join(dir, "ja"), 1, "Use Go", "Draft")
	setTestRoot(t, root)

	_, err := executeCmd(t, "status", "use-go", "accepted")
	if err == nil {
		t.Fatal("expected error when slug is ambiguous within a nested language variant")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Errorf("expected 'multiple' in error message, got: %v", err)
	}
}
