package search

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mskasa/kizami/internal/decision"
)

// Blame searches for documents in dir that mention the given file path.
// It also matches documents whose Related Files section contains a directory entry
// (trailing slash convention, e.g. "internal/") that is a prefix of filePath.
// Both .md and .kizami sidecar files are searched.
// Results are deduplicated by file and sorted by decision ID.
func Blame(dir, filePath string) ([]*decision.Decision, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	var matchedFiles []string
	var err error

	if _, lookErr := exec.LookPath("rg"); lookErr == nil {
		matchedFiles, err = blameRipgrep(dir, filePath)
	} else {
		matchedFiles, err = blameStdlib(dir, filePath)
	}
	if err != nil {
		return nil, err
	}

	// Also match documents with directory entries in Related Files.
	dirMatches, err := blameDirEntries(dir, filePath)
	if err != nil {
		return nil, err
	}
	matchedFiles = append(matchedFiles, dirMatches...)

	seen := make(map[string]struct{})
	var decisions []*decision.Decision
	for _, f := range matchedFiles {
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}

		var d *decision.Decision
		if decision.IsSidecarFile(f) {
			d, err = decision.ParseSidecar(f)
		} else {
			d, err = decision.Parse(f)
		}
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f, err)
		}
		decisions = append(decisions, d)
	}

	sort.Slice(decisions, func(i, j int) bool {
		return decisions[i].ID < decisions[j].ID
	})

	return decisions, nil
}

// blameDirEntries returns document files whose Related Files section contains a directory or
// glob entry (per decision.Match) matching filePath. Exact-path entries are intentionally
// excluded here: they are already found by the full-text search above, so matching them again
// would just be redundant work. Both .md and .kizami sidecar files are checked.
// Decision to unify this against decision.CheckHook's matching logic:
// docs/decisions/2026-08-03-related-files-single-definition.md
func blameDirEntries(dir, filePath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") && !decision.IsSidecarFile(path) {
			return nil
		}
		entries, parseErr := decision.ParseRelatedFiles(path)
		if parseErr != nil {
			return nil // non-fatal: skip unreadable files
		}
		for _, entry := range entries {
			kind, ok := decision.Match(entry, filePath)
			if !ok || kind == decision.MatchExact {
				continue
			}
			files = append(files, path)
			return nil // one match per file is enough
		}
		return nil
	})
	return files, err
}

func blameRipgrep(dir, filePath string) ([]string, error) {
	out, err := exec.Command("rg", "--files-with-matches", "--glob", "*.{md,kizami}", filePath, dir).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("ripgrep: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func blameStdlib(dir, filePath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") && !decision.IsSidecarFile(path) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), filePath) {
				files = append(files, path)
				return nil
			}
		}
		return scanner.Err()
	})
	return files, err
}
