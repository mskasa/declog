package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_FileNotExist(t *testing.T) {
	cfg, err := parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AI.Model != "" {
		t.Errorf("expected empty model, got %q", cfg.AI.Model)
	}
}

func TestLoad_ParsesModel(t *testing.T) {
	content := "[ai]\nmodel = \"claude-opus-4-20250514\"\n"
	cfg, err := parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AI.Model != "claude-opus-4-20250514" {
		t.Errorf("expected claude-opus-4-20250514, got %q", cfg.AI.Model)
	}
}

func TestLoad_IgnoresOtherSections(t *testing.T) {
	content := "[other]\nmodel = \"ignore-me\"\n[ai]\nmodel = \"claude-opus-4-20250514\"\n"
	cfg, err := parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AI.Model != "claude-opus-4-20250514" {
		t.Errorf("got %q", cfg.AI.Model)
	}
}

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[ai]\nmodel = \"claude-haiku-4-5-20251001\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, err := parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AI.Model != "claude-haiku-4-5-20251001" {
		t.Errorf("got %q", cfg.AI.Model)
	}
}

func TestResolveModel_FlagTakesPriority(t *testing.T) {
	cfg := &Config{AI: AIConfig{Model: "from-config"}}
	got := ResolveModel("from-flag", cfg)
	if got != "from-flag" {
		t.Errorf("expected from-flag, got %q", got)
	}
}

func TestResolveModel_ConfigOverDefault(t *testing.T) {
	cfg := &Config{AI: AIConfig{Model: "from-config"}}
	got := ResolveModel("", cfg)
	if got != "from-config" {
		t.Errorf("expected from-config, got %q", got)
	}
}

func TestResolveModel_DefaultWhenEmpty(t *testing.T) {
	cfg := &Config{}
	got := ResolveModel("", cfg)
	if got != DefaultModel {
		t.Errorf("expected %q, got %q", DefaultModel, got)
	}
}

func TestResolveModel_NilConfig(t *testing.T) {
	got := ResolveModel("", nil)
	if got != DefaultModel {
		t.Errorf("expected %q, got %q", DefaultModel, got)
	}
}

func TestLoad_ParsesAllSections(t *testing.T) {
	content := `
[ai]
model = "claude-opus-4-20250514"

[decisions]
dir = "records/decisions"

[audit]
dirs = ["docs/decisions", "docs/design"]

[review]
months_threshold = 12

[editor]
command = "code --wait"
`
	cfg, err := parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AI.Model != "claude-opus-4-20250514" {
		t.Errorf("AI.Model: got %q", cfg.AI.Model)
	}
	if cfg.Decisions.Dir != "records/decisions" {
		t.Errorf("Decisions.Dir: got %q", cfg.Decisions.Dir)
	}
	if len(cfg.Audit.Dirs) != 2 || cfg.Audit.Dirs[0] != "docs/decisions" || cfg.Audit.Dirs[1] != "docs/design" {
		t.Errorf("Audit.Dirs: got %v", cfg.Audit.Dirs)
	}
	if cfg.Review.MonthsThreshold != 12 {
		t.Errorf("Review.MonthsThreshold: got %d", cfg.Review.MonthsThreshold)
	}
	if cfg.Editor.Command != "code --wait" {
		t.Errorf("Editor.Command: got %q", cfg.Editor.Command)
	}
}

func TestLoad_ParsesAuditDirs_Single(t *testing.T) {
	content := "[audit]\ndirs = [\"docs/decisions\"]\n"
	cfg, err := parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Audit.Dirs) != 1 || cfg.Audit.Dirs[0] != "docs/decisions" {
		t.Errorf("Audit.Dirs: got %v", cfg.Audit.Dirs)
	}
}

func TestLoad_DefaultsWhenEmpty(t *testing.T) {
	cfg, err := parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Decisions.Dir != "" {
		t.Errorf("expected empty Decisions.Dir, got %q", cfg.Decisions.Dir)
	}
	if cfg.Review.MonthsThreshold != 0 {
		t.Errorf("expected 0 MonthsThreshold, got %d", cfg.Review.MonthsThreshold)
	}
	if cfg.Editor.Command != "" {
		t.Errorf("expected empty Editor.Command, got %q", cfg.Editor.Command)
	}
}

func TestResolveModel_Priority(t *testing.T) {
	tests := []struct {
		name      string
		flagModel string
		cfgModel  string
		want      string
	}{
		{"flag wins over config and default", "flag-model", "config-model", "flag-model"},
		{"config wins over default", "", "config-model", "config-model"},
		{"default when nothing set", "", "", DefaultModel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AI: AIConfig{Model: tt.cfgModel}}
			got := ResolveModel(tt.flagModel, cfg)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
