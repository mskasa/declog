package decision

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// AuditResult holds an ADR and the Related Files entries that no longer exist.
type AuditResult struct {
	*Decision
	MissingFiles []string
}

// ParseRelatedFiles reads the Related Files section of a kizami document and returns
// the listed file paths. For .kizami sidecar files the related: YAML list is read;
// for Markdown files the ## Related Files section is parsed.
// Backtick wrappers and comment lines are stripped.
func ParseRelatedFiles(path string) ([]string, error) {
	if IsSidecarFile(path) {
		return ParseSidecarRelatedFiles(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var files []string
	inSection := false
	inFence := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if line == "## Related Files" {
			inSection = true
			continue
		}
		if inSection {
			// A new ## heading ends the section.
			if strings.HasPrefix(line, "## ") {
				break
			}
			// Skip blank lines and HTML comments.
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			// List item: "- `path`" or "- path"
			if entry, ok := strings.CutPrefix(trimmed, "- "); ok {
				entry = strings.Trim(entry, "`")
				entry = strings.TrimSpace(entry)
				if entry != "" {
					files = append(files, entry)
				}
			}
		}
	}
	return files, scanner.Err()
}

// Audit checks Related Files entries for all Active ADRs in dir against repoRoot.
// Returns results only for ADRs that have at least one missing file.
func Audit(dir, repoRoot string) ([]*AuditResult, error) {
	decisions, err := List(dir)
	if err != nil {
		return nil, err
	}

	var results []*AuditResult
	for _, d := range decisions {
		if !strings.EqualFold(d.Status, "Active") {
			continue
		}

		relatedFiles, err := ParseRelatedFiles(d.File)
		if err != nil {
			return nil, err
		}
		if len(relatedFiles) == 0 {
			continue
		}

		var missing []string
		for _, rel := range relatedFiles {
			if _, err := os.Stat(filepath.Join(repoRoot, rel)); os.IsNotExist(err) {
				missing = append(missing, rel)
			}
		}
		if len(missing) > 0 {
			results = append(results, &AuditResult{Decision: d, MissingFiles: missing})
		}
	}
	return results, nil
}
