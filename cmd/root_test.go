package cmd

import (
	"path/filepath"
	"testing"

	"github.com/mskasa/kizami/internal/config"
)

func TestDecisionsDir(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "nil config uses default",
			cfg:  nil,
			want: filepath.Join("/repo", "docs", "decisions"),
		},
		{
			name: "empty config uses default",
			cfg:  &config.Config{},
			want: filepath.Join("/repo", "docs", "decisions"),
		},
		{
			name: "config with custom dir",
			cfg:  &config.Config{Decisions: config.DecisionsConfig{Dir: "records/adrs"}},
			want: filepath.Join("/repo", "records", "adrs"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decisionsDir("/repo", tt.cfg)
			if got != tt.want {
				t.Errorf("decisionsDir = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDesignDir(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "nil config uses default",
			cfg:  nil,
			want: filepath.Join("/repo", "docs", "design"),
		},
		{
			name: "empty config uses default",
			cfg:  &config.Config{},
			want: filepath.Join("/repo", "docs", "design"),
		},
		{
			name: "config with custom dir",
			cfg:  &config.Config{Design: config.DesignConfig{Dir: "records/design"}},
			want: filepath.Join("/repo", "records", "design"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := designDir("/repo", tt.cfg)
			if got != tt.want {
				t.Errorf("designDir = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDocumentDirs(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want []string
	}{
		{
			name: "nil config falls back to decisionsDir",
			cfg:  nil,
			want: []string{filepath.Join("/repo", "docs", "decisions")},
		},
		{
			name: "empty documents dirs falls back to decisionsDir",
			cfg:  &config.Config{},
			want: []string{filepath.Join("/repo", "docs", "decisions")},
		},
		{
			name: "config with documents dirs",
			cfg: &config.Config{
				Documents: config.DocumentsConfig{Dirs: []string{"docs/decisions", "docs/design"}},
			},
			want: []string{
				filepath.Join("/repo", "docs", "decisions"),
				filepath.Join("/repo", "docs", "design"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := documentDirs("/repo", tt.cfg)
			if len(got) != len(tt.want) {
				t.Fatalf("documentDirs len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("documentDirs[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestAuditDirs(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want []string
	}{
		{
			name: "nil config falls back to documentDirs",
			cfg:  nil,
			want: []string{filepath.Join("/repo", "docs", "decisions")},
		},
		{
			name: "empty audit dirs falls back to documentDirs",
			cfg:  &config.Config{},
			want: []string{filepath.Join("/repo", "docs", "decisions")},
		},
		{
			name: "config with audit dirs",
			cfg: &config.Config{
				Audit: config.AuditConfig{Dirs: []string{"docs/decisions", "docs/design"}},
			},
			want: []string{
				filepath.Join("/repo", "docs", "decisions"),
				filepath.Join("/repo", "docs", "design"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auditDirs("/repo", tt.cfg)
			if len(got) != len(tt.want) {
				t.Fatalf("auditDirs len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("auditDirs[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
