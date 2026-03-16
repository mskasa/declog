package decision

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// StaleADR represents an Active ADR that has not been updated recently.
type StaleADR struct {
	*Decision
	LastUpdated time.Time
}

// LastUpdated returns the date of the most recent git commit for the given file.
func LastUpdated(path string) (time.Time, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%ci", "--", path).Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("git log: %w", err)
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return time.Time{}, fmt.Errorf("no git history for %s", path)
	}
	t, err := time.Parse("2006-01-02 15:04:05 -0700", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing date %q: %w", s, err)
	}
	return t, nil
}

// FindStale returns Active ADRs whose last update is older than months before now.
// lastUpdatedFn is injectable for testing; ADRs for which it returns an error are skipped.
func FindStale(decisions []*Decision, lastUpdatedFn func(string) (time.Time, error), now time.Time, months int) ([]*StaleADR, error) {
	threshold := now.AddDate(0, -months, 0)
	var result []*StaleADR
	for _, d := range decisions {
		if !strings.EqualFold(d.Status, "Active") {
			continue
		}
		t, err := lastUpdatedFn(d.File)
		if err != nil {
			continue
		}
		if t.Before(threshold) {
			result = append(result, &StaleADR{Decision: d, LastUpdated: t})
		}
	}
	return result, nil
}

// StaleADRs returns Active ADRs in dir that have not been updated in the given number of months.
func StaleADRs(dir string, months int) ([]*StaleADR, error) {
	decisions, err := List(dir)
	if err != nil {
		return nil, err
	}
	return FindStale(decisions, LastUpdated, time.Now(), months)
}

// MonthsAgo returns the approximate number of whole months between t and now.
func MonthsAgo(t, now time.Time) int {
	years := now.Year() - t.Year()
	months := int(now.Month()) - int(t.Month())
	total := years*12 + months
	// Subtract one if the day-of-month hasn't been reached yet this month.
	if now.Day() < t.Day() {
		total--
	}
	if total < 0 {
		return 0
	}
	return total
}
