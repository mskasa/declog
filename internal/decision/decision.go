package decision

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Decision represents a parsed architectural decision record.
type Decision struct {
	ID     int
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

	d := &Decision{File: path}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "# "):
			// Format: # NNNN: Title
			body := strings.TrimPrefix(line, "# ")
			if idx := strings.Index(body, ": "); idx != -1 {
				idStr := body[:idx]
				n, err := strconv.Atoi(idStr)
				if err == nil {
					d.ID = n
				}
				d.Title = body[idx+2:]
			}
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
