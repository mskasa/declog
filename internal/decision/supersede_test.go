package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSupersedable_Active(t *testing.T) {
	d := &Decision{Slug: "use-sqlite", Status: "Active"}
	if err := CheckSupersedable(d); err != nil {
		t.Errorf("expected no error for Active, got: %v", err)
	}
}

func TestCheckSupersedable_AlreadySuperseded(t *testing.T) {
	d := &Decision{Slug: "use-sqlite", Status: "Superseded by use-postgresql"}
	err := CheckSupersedable(d)
	if err == nil {
		t.Fatal("expected error for already-superseded document, got nil")
	}
	if !strings.Contains(err.Error(), "already Superseded by use-postgresql") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCheckSupersedable_Inactive(t *testing.T) {
	d := &Decision{Slug: "use-sqlite", Status: "Inactive"}
	err := CheckSupersedable(d)
	if err == nil {
		t.Fatal("expected error for Inactive document, got nil")
	}
	if !strings.Contains(err.Error(), "already Inactive") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSupersede_UpdatesOldAndCreatesNew(t *testing.T) {
	dir := t.TempDir()

	// Create an existing document to supersede.
	oldContent := "# Use SQLite\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n\n## Context\n\nSQLite chosen initially.\n"
	oldPath := filepath.Join(dir, "2026-01-01-use-sqlite.md")
	if err := os.WriteFile(oldPath, []byte(oldContent), 0o644); err != nil {
		t.Fatal(err)
	}

	old, err := FindBySlug(dir, "use-sqlite")
	if err != nil {
		t.Fatalf("FindBySlug: %v", err)
	}
	if err := CheckSupersedable(old); err != nil {
		t.Fatalf("CheckSupersedable: %v", err)
	}

	// Update old document status.
	status := "Superseded by use-postgresql"
	if err := UpdateStatus(oldPath, status, ""); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	// Verify old document is updated.
	updated, err := Parse(oldPath)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if updated.Status != status {
		t.Errorf("old document status = %q, want %q", updated.Status, status)
	}
}

func TestSupersede_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindBySlug(dir, "nonexistent-slug")
	if err == nil {
		t.Fatal("expected error for missing document, got nil")
	}
}
