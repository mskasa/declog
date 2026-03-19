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

// Blame searches for ADRs in dir that mention the given file path.
// It also matches ADRs whose Related Files section contains a directory entry
// (trailing slash convention, e.g. "internal/") that is a prefix of filePath.
// Results are deduplicated by ADR file and sorted by decision ID.
func Blame(dir, filePath string) ([]*decision.Decision, error) {
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

	// Also match ADRs with directory entries in Related Files.
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

		d, err := decision.Parse(f)
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

// blameDirEntries returns ADR files whose Related Files section contains a
// directory entry (ending with "/") that is a prefix of filePath.
func blameDirEntries(dir, filePath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		entries, parseErr := decision.ParseRelatedFiles(path)
		if parseErr != nil {
			return nil // non-fatal: skip unreadable files
		}
		for _, entry := range entries {
			if !strings.HasSuffix(entry, "/") {
				continue // not a directory entry
			}
			if strings.HasPrefix(filePath, entry) {
				files = append(files, path)
				return nil // one match per file is enough
			}
		}
		return nil
	})
	return files, err
}

func blameRipgrep(dir, filePath string) ([]string, error) {
	out, err := exec.Command("rg", "--files-with-matches", "--glob", "*.md", filePath, dir).Output()
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
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
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
