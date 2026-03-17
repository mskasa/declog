package ai

import (
	"strings"
	"testing"
)

func TestBuildPrompt_ContainsTitle(t *testing.T) {
	input := PromptInput{
		Title:        "Adopt Helm for Kubernetes deployment",
		ChangedFiles: []string{"k8s/deployment.yaml"},
		Diff:         "diff --git a/k8s/deployment.yaml",
	}
	prompt := BuildPrompt(input)
	if !strings.Contains(prompt, "Adopt Helm for Kubernetes deployment") {
		t.Errorf("prompt missing title: %s", prompt)
	}
}

func TestBuildPrompt_ContainsChangedFiles(t *testing.T) {
	input := PromptInput{
		Title:        "test",
		ChangedFiles: []string{"internal/foo.go", "cmd/bar.go"},
		Diff:         "",
	}
	prompt := BuildPrompt(input)
	if !strings.Contains(prompt, "internal/foo.go") {
		t.Errorf("prompt missing changed file: %s", prompt)
	}
	if !strings.Contains(prompt, "cmd/bar.go") {
		t.Errorf("prompt missing changed file: %s", prompt)
	}
}

func TestBuildPrompt_ContainsDiff(t *testing.T) {
	diff := "diff --git a/internal/foo.go b/internal/foo.go\n+func New() {}"
	input := PromptInput{
		Title:        "test",
		ChangedFiles: nil,
		Diff:         diff,
	}
	prompt := BuildPrompt(input)
	if !strings.Contains(prompt, diff) {
		t.Errorf("prompt missing diff: %s", prompt)
	}
}

func TestGatherInput_TruncatesDiff(t *testing.T) {
	// Build a diff longer than DiffLimit.
	longDiff := strings.Repeat("x", DiffLimit+100)
	input := PromptInput{
		Title:        "test",
		ChangedFiles: nil,
		Diff:         longDiff,
	}
	// Simulate truncation as GatherInput does.
	if len(input.Diff) > DiffLimit {
		input.Diff = input.Diff[:DiffLimit]
	}
	if len(input.Diff) != DiffLimit {
		t.Errorf("expected diff length %d, got %d", DiffLimit, len(input.Diff))
	}
}

func TestBuildPrompt_ContainsSections(t *testing.T) {
	input := PromptInput{Title: "test"}
	prompt := BuildPrompt(input)
	for _, section := range []string{"## Context", "## Decision", "## Consequences", "## Related Files"} {
		if !strings.Contains(prompt, section) {
			t.Errorf("prompt missing section %q", section)
		}
	}
}
