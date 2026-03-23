package decision

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var legacyHeadingPattern = regexp.MustCompile(`^# (\d{4}): (.+)$`)
var legacySupersedesPattern = regexp.MustCompile(`^(- Supersedes: )(\d{4})$`)
var legacySupersededByPattern = regexp.MustCompile(`^(- Superseded by: )(\d{4})$`)
var legacyStatusPattern = regexp.MustCompile(`^(- Status: Superseded by )(\d{4})(.*)$`)

// MigrateLegacyFiles renames NNNN-slug.md files to YYYY-MM-DD-slug.md and updates
// internal references (headings, Supersedes, Superseded by) to use slugs.
// Subdirectories are processed recursively.
// Returns the number of files migrated.
func MigrateLegacyFiles(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("reading dir: %w", err)
	}

	// First pass: build a map from NNNN → slug and date for this directory.
	idToSlug := make(map[string]string)
	idToDate := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() || !legacyFilePattern.MatchString(e.Name()) {
			continue
		}
		m := legacyFilePattern.FindStringSubmatch(e.Name())
		idStr := m[1]
		slug := m[2]
		idToSlug[idStr] = slug

		d, err := Parse(filepath.Join(dir, e.Name()))
		if err != nil {
			return 0, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		if d.Date == "" {
			return 0, fmt.Errorf("file %s has no date; cannot migrate", e.Name())
		}
		idToDate[idStr] = d.Date
	}

	// Second pass: rewrite content and rename files in this directory.
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			// Recurse into subdirectories.
			n, err := MigrateLegacyFiles(filepath.Join(dir, e.Name()))
			if err != nil {
				return count, err
			}
			count += n
			continue
		}
		if !legacyFilePattern.MatchString(e.Name()) {
			continue
		}
		m := legacyFilePattern.FindStringSubmatch(e.Name())
		idStr := m[1]
		slug := m[2]
		date := idToDate[idStr]

		oldPath := filepath.Join(dir, e.Name())
		newName := date + "-" + slug + ".md"
		newPath := filepath.Join(dir, newName)

		data, err := os.ReadFile(oldPath)
		if err != nil {
			return count, fmt.Errorf("reading %s: %w", e.Name(), err)
		}

		updated := rewriteLegacyContent(string(data), idToSlug)

		if err := os.WriteFile(newPath, []byte(updated), 0o644); err != nil {
			return count, fmt.Errorf("writing %s: %w", newName, err)
		}
		if err := os.Remove(oldPath); err != nil {
			return count, fmt.Errorf("removing %s: %w", e.Name(), err)
		}
		count++
	}
	return count, nil
}

// rewriteLegacyContent rewrites heading, Supersedes, and Superseded by references
// from NNNN-based to slug-based format.
func rewriteLegacyContent(content string, idToSlug map[string]string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if m := legacyHeadingPattern.FindStringSubmatch(line); m != nil {
			// # NNNN: Title → # Title
			lines[i] = "# " + m[2]
			continue
		}
		if m := legacySupersedesPattern.FindStringSubmatch(line); m != nil {
			// - Supersedes: NNNN → - Supersedes: slug
			if slug, ok := idToSlug[m[2]]; ok {
				lines[i] = m[1] + slug
			}
			continue
		}
		if m := legacySupersededByPattern.FindStringSubmatch(line); m != nil {
			// - Superseded by: NNNN → - Superseded by: slug
			if slug, ok := idToSlug[m[2]]; ok {
				lines[i] = m[1] + slug
			}
			continue
		}
		if m := legacyStatusPattern.FindStringSubmatch(line); m != nil {
			// - Status: Superseded by NNNN → - Status: Superseded by slug
			if slug, ok := idToSlug[m[2]]; ok {
				lines[i] = m[1] + slug + m[3]
			}
			continue
		}
	}
	return strings.Join(lines, "\n")
}
