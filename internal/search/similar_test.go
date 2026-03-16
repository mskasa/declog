package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeywords_Basic(t *testing.T) {
	got := Keywords("Use PostgreSQL for the database")
	want := []string{"use", "postgresql", "database"}
	if len(got) != len(want) {
		t.Fatalf("Keywords = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keywords[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKeywords_AllStopWords(t *testing.T) {
	got := Keywords("a the for in")
	if len(got) != 0 {
		t.Errorf("expected no keywords, got %v", got)
	}
}

func TestKeywords_Empty(t *testing.T) {
	got := Keywords("")
	if len(got) != 0 {
		t.Errorf("expected no keywords for empty title, got %v", got)
	}
}

func writeSimilarFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSimilar_MatchFound(t *testing.T) {
	dir := t.TempDir()
	writeSimilarFile(t, dir, "0001-use-postgresql.md",
		"# 0001: Use PostgreSQL\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n\n## Context\n\nPostgreSQL selected for relational data.\n")
	writeSimilarFile(t, dir, "0002-use-redis.md",
		"# 0002: Use Redis\n\n- Date: 2026-01-02\n- Status: Active\n- Author: alice\n\n## Context\n\nRedis selected for caching.\n")

	decisions, err := Similar(dir, "postgresql database")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(decisions), decisions)
	}
	if decisions[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", decisions[0].ID)
	}
}

func TestSimilar_NoMatch(t *testing.T) {
	dir := t.TempDir()
	writeSimilarFile(t, dir, "0001-use-go.md",
		"# 0001: Use Go\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n\n## Context\n\nGo chosen for its simplicity.\n")

	decisions, err := Similar(dir, "postgresql database")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected no results, got %v", decisions)
	}
}

func TestSimilar_Deduplicated(t *testing.T) {
	dir := t.TempDir()
	// File matches multiple keywords — should appear only once.
	writeSimilarFile(t, dir, "0001-connection-pool.md",
		"# 0001: Connection Pool\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n\n## Context\n\nconnection pool size set to 10.\n")

	decisions, err := Similar(dir, "connection pool size")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 1 {
		t.Errorf("expected 1 result (deduplicated), got %d", len(decisions))
	}
}

func TestSimilar_StopWordsExcluded(t *testing.T) {
	dir := t.TempDir()
	// File contains "the" but that's a stop word — should not match.
	writeSimilarFile(t, dir, "0001-use-go.md",
		"# 0001: Use Go\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n")

	// "a the for" are all stop words — no keywords remain, so no search is run.
	decisions, err := Similar(dir, "a the for")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 0 {
		t.Errorf("expected no results for all-stopword title, got %v", decisions)
	}
}

func TestSimilar_SortedByID(t *testing.T) {
	dir := t.TempDir()
	writeSimilarFile(t, dir, "0003-cache-policy.md",
		"# 0003: Cache Policy\n\n- Date: 2026-01-03\n- Status: Active\n- Author: alice\n\n## Context\n\ncache strategy decision.\n")
	writeSimilarFile(t, dir, "0001-cache-config.md",
		"# 0001: Cache Config\n\n- Date: 2026-01-01\n- Status: Active\n- Author: alice\n\n## Context\n\ncache configuration approach.\n")

	decisions, err := Similar(dir, "cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decisions) != 2 {
		t.Fatalf("expected 2 results, got %d", len(decisions))
	}
	if decisions[0].ID != 1 || decisions[1].ID != 3 {
		t.Errorf("expected IDs [1, 3], got [%d, %d]", decisions[0].ID, decisions[1].ID)
	}
}
