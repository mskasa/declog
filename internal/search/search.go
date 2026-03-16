package search

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Result represents a single search match.
type Result struct {
	File string
	Line int
	Text string
}

// Run searches for keyword in dir using ripgrep if available, otherwise stdlib.
func Run(dir, keyword string) ([]Result, error) {
	if _, err := exec.LookPath("rg"); err == nil {
		return runRipgrep(dir, keyword)
	}
	return runStdlib(dir, keyword)
}

// RunCaseInsensitive searches for keyword case-insensitively.
func RunCaseInsensitive(dir, keyword string) ([]Result, error) {
	if _, err := exec.LookPath("rg"); err == nil {
		return runRipgrepCI(dir, keyword)
	}
	return runStdlibCI(dir, keyword)
}

func runRipgrepCI(dir, keyword string) ([]Result, error) {
	out, err := exec.Command("rg", "--line-number", "--no-heading", "--with-filename", "-i", keyword, dir).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("ripgrep: %w", err)
	}

	var results []Result
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		n, _ := strconv.Atoi(parts[1])
		results = append(results, Result{
			File: parts[0],
			Line: n,
			Text: strings.TrimSpace(parts[2]),
		})
	}
	return results, nil
}

func runStdlibCI(dir, keyword string) ([]Result, error) {
	lower := strings.ToLower(keyword)
	var results []Result
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

		lineNum := 0
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lineNum++
			text := scanner.Text()
			if strings.Contains(strings.ToLower(text), lower) {
				results = append(results, Result{
					File: path,
					Line: lineNum,
					Text: strings.TrimSpace(text),
				})
			}
		}
		return scanner.Err()
	})
	return results, err
}

func runRipgrep(dir, keyword string) ([]Result, error) {
	out, err := exec.Command("rg", "--line-number", "--no-heading", "--with-filename", keyword, dir).Output()
	if err != nil {
		// exit code 1 means no matches — not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("ripgrep: %w", err)
	}

	var results []Result
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		// Format: file:linenum:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		n, _ := strconv.Atoi(parts[1])
		results = append(results, Result{
			File: parts[0],
			Line: n,
			Text: strings.TrimSpace(parts[2]),
		})
	}
	return results, nil
}

func runStdlib(dir, keyword string) ([]Result, error) {
	var results []Result
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

		lineNum := 0
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lineNum++
			text := scanner.Text()
			if strings.Contains(text, keyword) {
				results = append(results, Result{
					File: path,
					Line: lineNum,
					Text: strings.TrimSpace(text),
				})
			}
		}
		return scanner.Err()
	})
	return results, err
}
