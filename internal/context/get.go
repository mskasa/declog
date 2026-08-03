package context

import (
	"fmt"
	"os"
)

// Document is a decision's full content, as returned by GetBySlug.
type Document struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Date         string `json:"date"`
	Path         string `json:"path"`
	SupersededBy string `json:"superseded_by,omitempty"`
	Markdown     string `json:"markdown"`
}

// GetBySlug returns every document across dirs whose filename slug matches slug. Usually
// one, but this repository's own EN/JA document pairs share a slug and are both legitimate
// matches, so callers must handle more than one result.
//
// This deliberately does not use decision.FindBySlug: that function stops at the first match
// within a directory tree (filepath.SkipAll), so it would miss a slug's ja/ counterpart
// nested inside the same configured directory — discovered while building this function,
// see docs/decisions/2026-08-03-mcp-tools-as-questions-not-verbs.md. Scanning every
// recognized document via collectAllDocs avoids that early-exit entirely.
func GetBySlug(dirs []string, repoRoot, slug string) ([]*Document, error) {
	all, err := collectAllDocs(dirs)
	if err != nil {
		return nil, err
	}

	var docs []*Document
	for _, d := range all {
		if d.Slug != slug {
			continue
		}
		content, err := os.ReadFile(d.File)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", d.File, err)
		}
		docs = append(docs, &Document{
			Slug:         d.Slug,
			Title:        d.Title,
			Status:       d.Status,
			Date:         d.Date,
			Path:         relPath(d.File, repoRoot),
			SupersededBy: d.SupersededBy(),
			Markdown:     string(content),
		})
	}
	return docs, nil
}
