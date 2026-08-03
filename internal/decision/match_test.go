package decision

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatch_Exact(t *testing.T) {
	kind, ok := Match("internal/decision/audit.go", "internal/decision/audit.go")
	if !ok || kind != MatchExact {
		t.Fatalf("got (%v, %v), want (%v, true)", kind, ok, MatchExact)
	}

	if _, ok := Match("internal/decision/audit.go", "internal/decision/hook.go"); ok {
		t.Fatal("expected no match for different file")
	}
}

func TestMatch_Dir(t *testing.T) {
	tests := []struct {
		file string
		want bool
		kind MatchKind
	}{
		{"internal/decision/audit.go", true, MatchDir},
		{"internal/decision", true, MatchExact}, // entry's trailing "/" is cosmetic
		{"internal/decision-other/audit.go", false, ""},
		{"internal/search/blame.go", false, ""},
	}
	for _, tt := range tests {
		kind, ok := Match("internal/decision/", tt.file)
		if ok != tt.want {
			t.Errorf("Match(%q) ok = %v, want %v", tt.file, ok, tt.want)
		}
		if ok && kind != tt.kind {
			t.Errorf("Match(%q) kind = %v, want %v", tt.file, kind, tt.kind)
		}
	}
}

// TestMatch_DirWithoutTrailingSlash locks in that a bare entry (no trailing "/") also matches as
// a directory prefix, matching CheckHook's pre-existing tested behavior (see hook_test.go),
// which this package's Match now implements as the single canonical rule.
func TestMatch_DirWithoutTrailingSlash(t *testing.T) {
	tests := []struct {
		file string
		want bool
	}{
		{"internal/db/db.go", true},
		{"internal/db", true},
		{"internal/dbc/other.go", false}, // must not match on a bare string prefix ("db" vs "dbc")
	}
	for _, tt := range tests {
		_, ok := Match("internal/db", tt.file)
		if ok != tt.want {
			t.Errorf("Match(%q) ok = %v, want %v", tt.file, ok, tt.want)
		}
	}
}

func TestMatch_Glob(t *testing.T) {
	tests := []struct {
		entry string
		file  string
		want  bool
	}{
		{"internal/**/*_test.go", "internal/decision/audit_test.go", true},
		{"internal/**/*_test.go", "internal/audit_test.go", true},
		{"internal/**/*_test.go", "internal/decision/audit.go", false},
		{"internal/**/*_test.go", "cmd/audit_test.go", false},
		{"internal/*/audit.go", "internal/decision/audit.go", true},
		{"internal/*/audit.go", "internal/decision/nested/audit.go", false},
		{"internal/decision/?udit.go", "internal/decision/audit.go", true},
	}
	for _, tt := range tests {
		kind, ok := Match(tt.entry, tt.file)
		if ok != tt.want {
			t.Errorf("Match(%q, %q) ok = %v, want %v", tt.entry, tt.file, ok, tt.want)
		}
		if ok && kind != MatchGlob {
			t.Errorf("Match(%q, %q) kind = %v, want %v", tt.entry, tt.file, kind, MatchGlob)
		}
	}
}

func TestEntryExists(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "decision"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "decision", "audit.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		entry string
		want  bool
	}{
		{"internal/decision/audit.go", true},
		{"internal/decision/missing.go", false},
		{"internal/decision/", true},
		{"internal/missing/", false},
		{"internal/**/*_test.go", true}, // glob entries are not verified
	}
	for _, tt := range tests {
		if got := EntryExists(root, tt.entry); got != tt.want {
			t.Errorf("EntryExists(%q) = %v, want %v", tt.entry, got, tt.want)
		}
	}
}
