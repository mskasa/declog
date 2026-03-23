package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleADR = `# Use MADR Format

- Date: 2026-03-12
- Status: Proposed
- Author: alice

## Context

Some context.
`

func writeADR(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNormalizeStatus(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"accepted", "Accepted"},
		{"ACCEPTED", "Accepted"},
		{"Proposed", "Proposed"},
		{"superseded", "Superseded"},
		{"deprecated", "Deprecated"},
	}
	for _, tt := range tests {
		got, err := NormalizeStatus(tt.input)
		if err != nil {
			t.Errorf("NormalizeStatus(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeStatus_Invalid(t *testing.T) {
	_, err := NormalizeStatus("unknown")
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestUpdateStatus_Simple(t *testing.T) {
	path := writeADR(t, t.TempDir(), "2026-03-12-use-madr.md", sampleADR)

	if err := UpdateStatus(path, "Accepted", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	body := string(data)
	if !strings.Contains(body, "- Status: Accepted") {
		t.Errorf("status not updated:\n%s", body)
	}
	if strings.Contains(body, "- Superseded by:") {
		t.Errorf("unexpected Superseded by line:\n%s", body)
	}
}

func TestUpdateStatus_Superseded(t *testing.T) {
	path := writeADR(t, t.TempDir(), "2026-03-12-use-madr.md", sampleADR)

	if err := UpdateStatus(path, "Superseded", "use-new-format"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	body := string(data)
	if !strings.Contains(body, "- Status: Superseded") {
		t.Errorf("status not updated:\n%s", body)
	}
	if !strings.Contains(body, "- Superseded by: use-new-format") {
		t.Errorf("missing Superseded by line:\n%s", body)
	}
}

func TestUpdateStatus_RemovesSupersededBy(t *testing.T) {
	content := strings.ReplaceAll(sampleADR, "- Status: Proposed", "- Status: Superseded\n- Superseded by: use-old-format")
	path := writeADR(t, t.TempDir(), "2026-03-12-use-madr.md", content)

	if err := UpdateStatus(path, "Accepted", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	body := string(data)
	if !strings.Contains(body, "- Status: Accepted") {
		t.Errorf("status not updated:\n%s", body)
	}
	if strings.Contains(body, "- Superseded by:") {
		t.Errorf("Superseded by line should have been removed:\n%s", body)
	}
}
