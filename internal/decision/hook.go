package decision

import (
	"path/filepath"
	"strings"
)

// HookResult holds an Active document whose Related Files overlap with staged files
// but which is not itself staged.
type HookResult struct {
	*Decision
	MatchedFiles []string // staged file paths that matched Related Files entries
}

// CheckHook scans Active documents in dirs for Related Files entries that overlap
// with stagedFiles (paths relative to repoRoot). Documents that are themselves staged
// are skipped. Returns one HookResult per affected document.
func CheckHook(dirs []string, repoRoot string, stagedFiles []string) ([]*HookResult, error) {
	stagedSet := make(map[string]struct{}, len(stagedFiles))
	for _, f := range stagedFiles {
		stagedSet[filepath.ToSlash(f)] = struct{}{}
	}

	var results []*HookResult
	seen := make(map[string]struct{})

	for _, dir := range dirs {
		docs, err := List(dir)
		if err != nil {
			return nil, err
		}
		for _, d := range docs {
			if !strings.EqualFold(d.Status, "Active") {
				continue
			}
			if _, already := seen[d.File]; already {
				continue
			}
			seen[d.File] = struct{}{}

			// Skip if the document itself is staged.
			relDoc, err := filepath.Rel(repoRoot, d.File)
			if err != nil {
				continue
			}
			if _, docStaged := stagedSet[filepath.ToSlash(relDoc)]; docStaged {
				continue
			}

			relatedFiles, err := ParseRelatedFiles(d.File)
			if err != nil {
				return nil, err
			}

			var matched []string
			for _, staged := range stagedFiles {
				for _, rel := range relatedFiles {
					if hookPathMatches(rel, staged) {
						matched = append(matched, staged)
						break
					}
				}
			}
			if len(matched) > 0 {
				results = append(results, &HookResult{Decision: d, MatchedFiles: matched})
			}
		}
	}
	return results, nil
}

// hookPathMatches reports whether staged (relative to repo root) matches a Related Files entry.
// An entry is treated as a directory prefix when staged has it as a path component prefix.
func hookPathMatches(related, staged string) bool {
	rel := filepath.ToSlash(strings.TrimSuffix(related, "/"))
	s := filepath.ToSlash(staged)
	return s == rel || strings.HasPrefix(s, rel+"/")
}
