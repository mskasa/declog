package context

import (
	"os"
	"strings"
)

// defaultSearchLimit is applied when limit <= 0.
const defaultSearchLimit = 10

// SearchResult is one document matching a keyword search.
type SearchResult struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Path         string `json:"path"`
	Decision     string `json:"decision"`
	SupersededBy string `json:"superseded_by,omitempty"`
	Excerpt      string `json:"excerpt"`
}

// Search returns one SearchResult per kizami document across dirs whose content contains
// keyword (case-insensitive), up to limit results (<=0 uses defaultSearchLimit). Unlike
// internal/search.Run, which greps every ".md" file, this enumerates recognized kizami
// documents via decision.List first, so a stray non-kizami Markdown file never appears in
// results: docs/decisions/2026-08-03-mcp-tools-as-questions-not-verbs.md
//
// Superseded decisions are excluded unless includeSuperseded is true. Draft and Inactive
// decisions are not excluded — unlike Resolve/Manifest, Search isn't about what currently
// governs a file, it's about whether something has been discussed at all.
func Search(dirs []string, repoRoot, keyword string, limit int, includeSuperseded bool) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	docs, err := collectAllDocs(dirs)
	if err != nil {
		return nil, err
	}

	var results []*SearchResult
	for _, d := range docs {
		if !includeSuperseded && d.SupersededBy() != "" {
			continue
		}

		excerpt, ok, err := firstMatchingLine(d.File, keyword)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		summary, err := summaryText(d.File, false)
		if err != nil {
			return nil, err
		}

		results = append(results, &SearchResult{
			Slug:         d.Slug,
			Title:        d.Title,
			Status:       d.Status,
			Path:         relPath(d.File, repoRoot),
			Decision:     summary,
			SupersededBy: d.SupersededBy(),
			Excerpt:      excerpt,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

// firstMatchingLine returns the first line of path containing keyword (case-insensitive).
func firstMatchingLine(path, keyword string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	lower := strings.ToLower(keyword)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(strings.ToLower(line), lower) {
			return strings.TrimSpace(line), true, nil
		}
	}
	return "", false, nil
}
