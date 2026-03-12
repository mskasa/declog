package decision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	tmpl "github.com/mskasa/declog/internal/template"
)

var filePattern = regexp.MustCompile(`^(\d{4})-.*\.md$`)

// NextID returns the next available 4-digit ID by scanning the decisions directory.
func NextID(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, fmt.Errorf("reading decisions dir: %w", err)
	}

	max := 0
	for _, e := range entries {
		m := filePattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}

// Slugify converts a title to kebab-case.
func Slugify(title string) string {
	s := strings.ToLower(title)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// AuthorFromGit returns the git user.name, or "unknown" if unavailable.
func AuthorFromGit() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// List returns all decisions in dir sorted by ID descending.
func List(dir string) ([]*Decision, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading decisions dir: %w", err)
	}

	var decisions []*Decision
	for _, e := range entries {
		if !filePattern.MatchString(e.Name()) {
			continue
		}
		d, err := Parse(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}

	// Sort by ID descending.
	for i := 0; i < len(decisions)-1; i++ {
		for j := i + 1; j < len(decisions); j++ {
			if decisions[i].ID < decisions[j].ID {
				decisions[i], decisions[j] = decisions[j], decisions[i]
			}
		}
	}
	return decisions, nil
}

// FindByID returns the decision with the given ID, or an error if not found.
func FindByID(dir string, id int) (*Decision, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading decisions dir: %w", err)
	}

	prefix := fmt.Sprintf("%04d-", id)
	for _, e := range entries {
		if !filePattern.MatchString(e.Name()) {
			continue
		}
		if strings.HasPrefix(e.Name(), prefix) {
			return Parse(filepath.Join(dir, e.Name()))
		}
	}
	return nil, fmt.Errorf("decision %04d not found", id)
}

// Create generates a new ADR file and returns its path.
func Create(dir, title string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating decisions dir: %w", err)
	}

	id, err := NextID(dir)
	if err != nil {
		return "", err
	}

	slug := Slugify(title)
	filename := fmt.Sprintf("%04d-%s.md", id, slug)
	path := filepath.Join(dir, filename)

	author := AuthorFromGit()
	content := tmpl.Render(id, title, author)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return path, nil
}
