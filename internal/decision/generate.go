package decision

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tmpl "github.com/mskasa/kizami/internal/template"
)

// legacyFilePattern matches legacy NNNN-slug.md filenames.
var legacyFilePattern = regexp.MustCompile(`^(\d{4})-([a-z].*)\.md$`)

// dateFilePattern matches new YYYY-MM-DD-slug.md filenames.
var dateFilePattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)\.md$`)

// isDocumentFile reports whether a filename matches the kizami naming convention.
func isDocumentFile(name string) bool {
	return dateFilePattern.MatchString(name) || legacyFilePattern.MatchString(name)
}

// isKizamiDocument reads a .md file and returns true if it contains both
// "- Status:" and "## Related Files", the two required markers for kizami documents.
// This allows arbitrary .md filenames (e.g. ARCHITECTURE.md) to be recognised.
func isKizamiDocument(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	hasStatus := false
	hasRelatedFiles := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "- Status:") {
			hasStatus = true
		}
		if line == "## Related Files" {
			hasRelatedFiles = true
		}
		if hasStatus && hasRelatedFiles {
			return true
		}
	}
	return false
}

// slugFromFilename extracts the semantic slug from a document filename.
// For "2026-03-23-use-go.md" returns "use-go".
// For "0001-use-go.md" returns "use-go".
// For arbitrary filenames like "ARCHITECTURE.md" returns "ARCHITECTURE".
func slugFromFilename(name string) string {
	if m := dateFilePattern.FindStringSubmatch(name); m != nil {
		return strings.TrimSuffix(m[2], ".md")
	}
	if m := legacyFilePattern.FindStringSubmatch(name); m != nil {
		return strings.TrimSuffix(m[2], ".md")
	}
	return strings.TrimSuffix(name, ".md")
}

// sortKey returns a date string used for descending sort in List.
// Uses the Date field from front-matter if present, falls back to the file's
// modification time, and finally to "0000-00-00" if stat fails.
func sortKey(d *Decision) string {
	if d.Date != "" {
		return d.Date
	}
	info, err := os.Stat(d.File)
	if err == nil {
		return info.ModTime().Format("2006-01-02")
	}
	return "0000-00-00"
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
// Decision to use git config instead of an environment variable: docs/decisions/0009-author-source.md
func AuthorFromGit() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// List returns all decisions in dir (recursively) sorted by date descending (newest first).
// Recognises both filename-patterned files (YYYY-MM-DD-*.md, NNNN-*.md) and arbitrary
// .md files that contain both "- Status:" and "## Related Files" markers.
func List(dir string) ([]*Decision, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	var decisions []*Decision
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		// Fast path: filename pattern match (no I/O).
		// Slow path: content check for non-pattern filenames.
		if !isDocumentFile(d.Name()) && !isKizamiDocument(path) {
			return nil
		}
		doc, parseErr := Parse(path)
		if parseErr != nil {
			return parseErr
		}
		decisions = append(decisions, doc)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("reading decisions dir: %w", err)
	}

	// Sort by date descending, then by filename descending as tiebreaker.
	// Files without a Date field are sorted by mtime (see sortKey).
	for i := 0; i < len(decisions)-1; i++ {
		for j := i + 1; j < len(decisions); j++ {
			ki, kj := sortKey(decisions[i]), sortKey(decisions[j])
			less := ki < kj || (ki == kj && decisions[i].File < decisions[j].File)
			if less {
				decisions[i], decisions[j] = decisions[j], decisions[i]
			}
		}
	}
	return decisions, nil
}

// FindBySlug returns the decision whose filename slug matches the given slug, or an error if not found.
// Both legacy (NNNN-slug.md), date-prefixed (YYYY-MM-DD-slug.md), and arbitrary .md filenames
// are searched recursively. Arbitrary filenames are only considered if they contain both
// "- Status:" and "## Related Files" markers.
// Returns "not found" without error if dir does not exist.
func FindBySlug(dir, slug string) (*Decision, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("document %q not found", slug)
	}
	var found *Decision
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		if slugFromFilename(d.Name()) != slug {
			return nil
		}
		// Verify it is a kizami document.
		if !isDocumentFile(d.Name()) && !isKizamiDocument(path) {
			return nil
		}
		doc, parseErr := Parse(path)
		if parseErr != nil {
			return parseErr
		}
		found = doc
		return filepath.SkipAll
	})
	if err != nil {
		return nil, fmt.Errorf("reading decisions dir: %w", err)
	}
	if found == nil {
		return nil, fmt.Errorf("document %q not found", slug)
	}
	return found, nil
}

// CreateFromDraft creates an ADR file using AI-generated draft sections.
// The standard header (Date, Status, Author, Supersedes) is prepended to the draft.
func CreateFromDraft(dir, title, draft, supersededSlug string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating decisions dir: %w", err)
	}

	slug := Slugify(title)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)
	path := filepath.Join(dir, filename)

	author := AuthorFromGit()
	header := tmpl.RenderHeader(title, author, supersededSlug)
	content := header + "\n" + draft

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return path, nil
}

// CreateDesignFromDraft creates a design document file using AI-generated draft sections.
// The standard header (Date, Type, Status, Author, Supersedes) is prepended to the draft.
func CreateDesignFromDraft(dir, title, draft, supersededSlug string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating design dir: %w", err)
	}

	slug := Slugify(title)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)
	path := filepath.Join(dir, filename)

	author := AuthorFromGit()
	header := tmpl.RenderDesignHeader(title, author, supersededSlug)
	content := header + "\n" + draft

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}
	return path, nil
}

// Create generates a new ADR file and returns its path.
// supersededSlug, if non-empty, adds a "- Supersedes: <slug>" line to the template.
func Create(dir, title, supersededSlug string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating decisions dir: %w", err)
	}

	slug := Slugify(title)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)
	path := filepath.Join(dir, filename)

	author := AuthorFromGit()
	relatedFiles := tmpl.ChangedFiles(dir)
	content := tmpl.Render(title, author, relatedFiles, supersededSlug)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return path, nil
}

// CreateDesign generates a new design document file and returns its path.
// supersededSlug, if non-empty, adds a "- Supersedes: <slug>" line to the template.
func CreateDesign(dir, title, supersededSlug string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating design dir: %w", err)
	}

	slug := Slugify(title)
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)
	path := filepath.Join(dir, filename)

	author := AuthorFromGit()
	relatedFiles := tmpl.ChangedFiles(dir)
	content := tmpl.RenderDesign(title, author, relatedFiles, supersededSlug)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return path, nil
}
