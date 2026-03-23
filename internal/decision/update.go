package decision

import (
	"fmt"
	"os"
	"strings"
)

// ValidStatuses lists the allowed status values.
var ValidStatuses = []string{"Proposed", "Accepted", "Superseded", "Deprecated"}

// NormalizeStatus returns the canonical form of a status string (case-insensitive),
// or an error if the value is not recognized.
func NormalizeStatus(s string) (string, error) {
	lower := strings.ToLower(s)
	for _, v := range ValidStatuses {
		if strings.ToLower(v) == lower {
			return v, nil
		}
	}
	return "", fmt.Errorf("invalid status %q: must be one of %s", s, strings.Join(ValidStatuses, ", "))
}

// UpdateStatus rewrites the Status line in the decision file.
// If supersededBySlug is non-empty, a "- Superseded by: <slug>" line is inserted immediately
// after the Status line; any existing such line is removed beforehand.
func UpdateStatus(path, status, supersededBySlug string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines)+1)

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "- Status: "):
			out = append(out, "- Status: "+status)
			if supersededBySlug != "" {
				out = append(out, "- Superseded by: "+supersededBySlug)
			}
		case strings.HasPrefix(line, "- Superseded by: "):
			// Drop existing line; re-added above if still needed.
		default:
			out = append(out, line)
		}
	}

	if err := os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}
