package decision

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsSidecarFile reports whether path ends with the .kizami extension.
func IsSidecarFile(path string) bool {
	return strings.HasSuffix(path, ".kizami")
}

// slugFromSidecar extracts the slug from a .kizami filename.
// For "test_matrix.csv.kizami" returns "test_matrix.csv".
func slugFromSidecar(name string) string {
	return strings.TrimSuffix(name, ".kizami")
}

// ParseSidecar reads a .kizami YAML file and returns a Decision.
// Sidecars have no status field and are always treated as Active.
func ParseSidecar(path string) (*Decision, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening sidecar: %w", err)
	}
	defer f.Close()

	d := &Decision{
		File:   path,
		Slug:   slugFromSidecar(filepath.Base(path)),
		Status: "Active",
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(trimmed, "title:"):
			d.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "title:"))
		case strings.HasPrefix(trimmed, "date:"):
			d.Date = strings.TrimSpace(strings.TrimPrefix(trimmed, "date:"))
		case strings.HasPrefix(trimmed, "author:"):
			d.Author = strings.TrimSpace(strings.TrimPrefix(trimmed, "author:"))
		}
	}
	return d, scanner.Err()
}

// ParseSidecarRelatedFiles reads the related: list from a .kizami file.
func ParseSidecarRelatedFiles(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening sidecar: %w", err)
	}
	defer f.Close()

	var files []string
	inRelated := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "related:" {
			inRelated = true
			continue
		}
		if inRelated {
			if strings.HasPrefix(trimmed, "- ") {
				entry := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				entry = strings.Trim(entry, "`")
				if entry != "" {
					files = append(files, entry)
				}
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				break
			}
		}
	}
	return files, scanner.Err()
}
