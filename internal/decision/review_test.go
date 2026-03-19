package decision

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	testNow    = time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	staleTime  = time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC) // ~8 months before testNow
	recentTime = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) // ~1 month before testNow
)

func makeDecisions(dir string, t *testing.T, specs []struct {
	name, status string
}) []*Decision {
	t.Helper()
	var decisions []*Decision
	for _, s := range specs {
		content := "# 0001: Title\n\n- Date: 2026-01-01\n- Status: " + s.status + "\n- Author: alice\n"
		path := filepath.Join(dir, s.name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		d, err := Parse(path)
		if err != nil {
			t.Fatal(err)
		}
		decisions = append(decisions, d)
	}
	return decisions
}

func TestFindStale_DetectsStaleActiveADR(t *testing.T) {
	dir := t.TempDir()
	decisions := makeDecisions(dir, t, []struct{ name, status string }{
		{"0001-old.md", "Active"},
		{"0002-recent.md", "Active"},
	})

	lastUpdatedFn := func(path string) (time.Time, error) {
		if filepath.Base(path) == "0001-old.md" {
			return staleTime, nil
		}
		return recentTime, nil
	}

	stale, err := FindStale(decisions, lastUpdatedFn, testNow, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale ADR, got %d", len(stale))
	}
	if filepath.Base(stale[0].File) != "0001-old.md" {
		t.Errorf("unexpected stale file: %s", stale[0].File)
	}
}

func TestFindStale_RecentADRNotDetected(t *testing.T) {
	dir := t.TempDir()
	decisions := makeDecisions(dir, t, []struct{ name, status string }{
		{"0001-recent.md", "Active"},
	})

	lastUpdatedFn := func(path string) (time.Time, error) { return recentTime, nil }

	stale, err := FindStale(decisions, lastUpdatedFn, testNow, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected no stale ADRs, got %d", len(stale))
	}
}

func TestFindStale_SkipsNonActive(t *testing.T) {
	dir := t.TempDir()
	decisions := makeDecisions(dir, t, []struct{ name, status string }{
		{"0001-inactive.md", "Inactive"},
		{"0002-superseded.md", "Superseded by 0003"},
	})

	lastUpdatedFn := func(path string) (time.Time, error) { return staleTime, nil }

	stale, err := FindStale(decisions, lastUpdatedFn, testNow, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected no stale ADRs for non-Active, got %d", len(stale))
	}
}

func TestFindStale_NoStaleADRs(t *testing.T) {
	stale, err := FindStale([]*Decision{}, func(string) (time.Time, error) { return recentTime, nil }, testNow, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected empty result, got %d", len(stale))
	}
}

func TestFindStale_MonthsThreshold(t *testing.T) {
	dir := t.TempDir()
	decisions := makeDecisions(dir, t, []struct{ name, status string }{
		{"0001-old.md", "Active"},
	})
	// staleTime is ~8 months ago; with threshold=3 it should be stale, with threshold=12 it should not.
	lastUpdatedFn := func(path string) (time.Time, error) { return staleTime, nil }

	stale3, _ := FindStale(decisions, lastUpdatedFn, testNow, 3)
	if len(stale3) != 1 {
		t.Errorf("threshold=3: expected 1 stale ADR, got %d", len(stale3))
	}

	stale12, _ := FindStale(decisions, lastUpdatedFn, testNow, 12)
	if len(stale12) != 0 {
		t.Errorf("threshold=12: expected 0 stale ADRs, got %d", len(stale12))
	}
}

func TestMonthsAgo(t *testing.T) {
	tests := []struct {
		t    time.Time
		now  time.Time
		want int
	}{
		{time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), 8},
		{time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), 5},
	}
	for _, tt := range tests {
		got := MonthsAgo(tt.t, tt.now)
		if got != tt.want {
			t.Errorf("MonthsAgo(%v, %v) = %d, want %d", tt.t, tt.now, got, tt.want)
		}
	}
}
