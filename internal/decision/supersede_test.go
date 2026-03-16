package decision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSupersedable_Active(t *testing.T) {
	d := &Decision{ID: 1, Status: "Active"}
	if err := CheckSupersedable(d); err != nil {
		t.Errorf("expected no error for Active, got: %v", err)
	}
}

func TestCheckSupersedable_AlreadySuperseded(t *testing.T) {
	d := &Decision{ID: 3, Status: "Superseded by 0009"}
	err := CheckSupersedable(d)
	if err == nil {
		t.Fatal("expected error for already-superseded ADR, got nil")
	}
	if !strings.Contains(err.Error(), "already Superseded by 0009") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCheckSupersedable_Inactive(t *testing.T) {
	d := &Decision{ID: 2, Status: "Inactive"}
	err := CheckSupersedable(d)
	if err == nil {
		t.Fatal("expected error for Inactive ADR, got nil")
	}
	if !strings.Contains(err.Error(), "already Inactive") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSupersede_UpdatesOldAndCreatesNew(t *testing.T) {
	dir := t.TempDir()

	// Create an existing ADR to supersede.
	oldContent := "# 0001: Use SQLite\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n\n## Context\n\nSQLite chosen initially.\n"
	oldPath := filepath.Join(dir, "0001-use-sqlite.md")
	if err := os.WriteFile(oldPath, []byte(oldContent), 0o644); err != nil {
		t.Fatal(err)
	}

	old, err := FindByID(dir, 1)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if err := CheckSupersedable(old); err != nil {
		t.Fatalf("CheckSupersedable: %v", err)
	}

	newID, err := NextID(dir)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if newID != 2 {
		t.Fatalf("expected newID 2, got %d", newID)
	}

	// Update old ADR status.
	status := "Superseded by 0002"
	if err := UpdateStatus(oldPath, status, 0); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	// Verify old ADR is updated.
	updated, err := Parse(oldPath)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if updated.Status != status {
		t.Errorf("old ADR status = %q, want %q", updated.Status, status)
	}
}

func TestSupersede_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindByID(dir, 99)
	if err == nil {
		t.Fatal("expected error for missing ADR, got nil")
	}
}
