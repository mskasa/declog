package template

import (
	"fmt"
	"time"
)

// Render returns the MADR Markdown template filled with the given values.
func Render(id int, title, author string) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`# %04d: %s

- Date: %s
- Status: Active
- Author: %s

## Context

<!-- Why this decision was needed. Describe the background, constraints, and problem. -->

## Decision

<!-- What was decided. State clearly in 1–3 sentences. -->

## Consequences

<!-- Impact, benefits, and trade-offs of this decision. -->

## Alternatives Considered

<!-- Options that were considered but not adopted, and why. (Optional) -->

## Related Files

<!-- List files related to this decision (e.g. internal/search/search.go). -->
`, id, title, date, author)
}
