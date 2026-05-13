package decision

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Decision represents a parsed architectural decision record or design document.
type Decision struct {
	ID     int    // non-zero only for legacy NNNN- filenames
	Slug   string // semantic filename part without date/ID prefix
	Title  string
	Date   string
	Status string
	Author string
	File   string
}

// Parse reads a decision file and returns a Decision with its metadata.
func Parse(path string) (*Decision, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	d := &Decision{
		File: path,
		Slug: slugFromFilename(filepath.Base(path)),
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "# "):
			body := strings.TrimPrefix(line, "# ")
			// Legacy format: # NNNN: Title
			if idStr, rest, ok := strings.Cut(body, ": "); ok {
				n, err := strconv.Atoi(idStr)
				if err == nil {
					d.ID = n
					d.Title = rest
					continue
				}
			}
			// New format: # Title
			d.Title = body
		case strings.HasPrefix(line, "- Date: "):
			d.Date = strings.TrimPrefix(line, "- Date: ")
		case strings.HasPrefix(line, "- Status: "):
			d.Status = strings.TrimPrefix(line, "- Status: ")
		case strings.HasPrefix(line, "- Author: "):
			d.Author = strings.TrimPrefix(line, "- Author: ")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return d, nil
}

const excerptMaxLen = 100

// ParseDecisionExcerpt returns the first non-blank line of the ## Decision section,
// truncated to 100 characters. Returns "" for sidecar files or documents without
// a ## Decision section.
func ParseDecisionExcerpt(path string) (string, error) {
	if IsSidecarFile(path) {
		return "", nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

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
		if line == "## Decision" {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(line, "## ") {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			if len(trimmed) > excerptMaxLen {
				return trimmed[:excerptMaxLen] + "...", nil
			}
			return trimmed, nil
		}
	}
	return "", scanner.Err()
}
