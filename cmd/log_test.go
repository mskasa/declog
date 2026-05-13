package cmd

import (
	"strings"
	"testing"
)

func TestAdrCmd_AIFlag_MissingAPIKey(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)
	t.Setenv("ANTHROPIC_API_KEY", "")

	_, err := executeCmd(t, "adr", "--ai", "test title")
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is not set")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("expected ANTHROPIC_API_KEY error, got: %v", err)
	}
}

func TestDesignCmd_AIFlag_MissingAPIKey(t *testing.T) {
	root := newTestRepo(t)
	setTestRoot(t, root)
	t.Setenv("ANTHROPIC_API_KEY", "")

	_, err := executeCmd(t, "design", "--ai", "test title")
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is not set")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("expected ANTHROPIC_API_KEY error, got: %v", err)
	}
}
