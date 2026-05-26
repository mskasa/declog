package decision

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// LintIssue represents a structural problem found in a kizami document.
type LintIssue struct {
	File     string
	Message  string
	Severity string // "error" or "warning"
}

func errIssue(file, message string) *LintIssue {
	return &LintIssue{File: file, Message: message, Severity: "error"}
}

func warnIssue(file, message string) *LintIssue {
	return &LintIssue{File: file, Message: message, Severity: "warning"}
}

var lintDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Lint checks all kizami documents in dir for structural issues and returns all findings.
func Lint(dir, repoRoot string) ([]*LintIssue, error) {
	docs, err := List(dir)
	if err != nil {
		return nil, err
	}

	var issues []*LintIssue
	for _, d := range docs {
		var docIssues []*LintIssue
		if IsSidecarFile(filepath.Base(d.File)) {
			docIssues = lintSidecar(d, repoRoot)
		} else {
			docIssues = lintMarkdown(d, repoRoot)
		}
		issues = append(issues, docIssues...)
	}
	return issues, nil
}

func lintMarkdown(d *Decision, repoRoot string) []*LintIssue {
	var issues []*LintIssue
	rel := lintRelPath(d.File, repoRoot)

	if d.Status == "" {
		issues = append(issues, errIssue(rel, `missing "- Status:" field`))
	}
	if d.Date != "" && !lintDatePattern.MatchString(d.Date) {
		issues = append(issues, errIssue(rel, fmt.Sprintf(`"- Date:" value %q is not in YYYY-MM-DD format`, d.Date)))
	}

	related, err := ParseRelatedFiles(d.File)
	if err != nil {
		issues = append(issues, errIssue(rel, fmt.Sprintf("error reading Related Files: %v", err)))
		return issues
	}
	if len(related) == 0 {
		// Empty Related Files is a warning: the document may not yet have related code.
		// kizami audit and kizami blame simply skip documents with no entries.
		issues = append(issues, warnIssue(rel, `"## Related Files" section is missing or empty`))
	} else {
		for _, path := range related {
			if _, statErr := os.Stat(filepath.Join(repoRoot, path)); os.IsNotExist(statErr) {
				issues = append(issues, errIssue(rel, fmt.Sprintf("Related Files: path does not exist: %s", path)))
			}
		}
	}
	return issues
}

func lintSidecar(d *Decision, repoRoot string) []*LintIssue {
	var issues []*LintIssue
	rel := lintRelPath(d.File, repoRoot)

	if d.Title == "" {
		issues = append(issues, errIssue(rel, `missing "title:" field`))
	}
	if d.Date != "" && !lintDatePattern.MatchString(d.Date) {
		issues = append(issues, errIssue(rel, fmt.Sprintf(`"date:" value %q is not in YYYY-MM-DD format`, d.Date)))
	}

	related, err := ParseSidecarRelatedFiles(d.File)
	if err != nil {
		issues = append(issues, errIssue(rel, fmt.Sprintf("error reading related: %v", err)))
		return issues
	}
	if len(related) == 0 {
		// Sidecar files exist solely to annotate other files, so an empty related: list is always an error.
		issues = append(issues, errIssue(rel, `"related:" list is missing or empty`))
	} else {
		for _, path := range related {
			if _, statErr := os.Stat(filepath.Join(repoRoot, path)); os.IsNotExist(statErr) {
				issues = append(issues, errIssue(rel, fmt.Sprintf("related: path does not exist: %s", path)))
			}
		}
	}
	return issues
}

func lintRelPath(file, repoRoot string) string {
	rel, err := filepath.Rel(repoRoot, file)
	if err != nil {
		return file
	}
	return rel
}
