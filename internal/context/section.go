package context

import (
	"os"
	"strings"
)

// summaryHeadings are tried, in file order, when extracting a decision's one-line summary.
// ADRs use "## Decision"; design documents use "## Overview". A document has at most one
// of the two, so whichever is present in the file is the one returned.
var summaryHeadings = []string{"## Decision", "## Overview"}

// summaryText returns a decision's summary for the JSON/CLI output: the full document
// content when full is true, otherwise the body of whichever heading in summaryHeadings
// is present. Returns "" if full is false and neither heading is present (e.g. a .kizami
// sidecar, which has no Markdown sections).
func summaryText(path string, full bool) (string, error) {
	if full {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}
	return extractSection(path, summaryHeadings)
}

// extractSection returns the trimmed body text of the first heading (in file order) that
// appears in headings, stopping at the next "## " heading. Unlike ParseRelatedFiles, code
// fences are not treated specially — the section's prose may legitimately contain them.
func extractSection(path string, headings []string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	headingSet := make(map[string]bool, len(headings))
	for _, h := range headings {
		headingSet[h] = true
	}

	var body []string
	inSection := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "## ") {
			if inSection {
				break
			}
			inSection = headingSet[strings.TrimSpace(line)]
			continue
		}
		if inSection {
			body = append(body, line)
		}
	}
	return strings.TrimSpace(strings.Join(body, "\n")), nil
}
