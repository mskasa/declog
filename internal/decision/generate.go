package decision

import (
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

// isDocumentFile reports whether a filename is a document file (either format).
func isDocumentFile(name string) bool {
	return dateFilePattern.MatchString(name) || legacyFilePattern.MatchString(name)
}

// slugFromFilename extracts the semantic slug from a document filename.
// For "2026-03-23-use-go.md" returns "use-go".
// For "0001-use-go.md" returns "use-go".
func slugFromFilename(name string) string {
	if m := dateFilePattern.FindStringSubmatch(name); m != nil {
		return strings.TrimSuffix(m[2], ".md")
	}
	if m := legacyFilePattern.FindStringSubmatch(name); m != nil {
		return strings.TrimSuffix(m[2], ".md")
	}
	return ""
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
		if !isDocumentFile(d.Name()) {
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
	for i := 0; i < len(decisions)-1; i++ {
		for j := i + 1; j < len(decisions); j++ {
			less := decisions[i].Date < decisions[j].Date ||
				(decisions[i].Date == decisions[j].Date && decisions[i].File < decisions[j].File)
			if less {
				decisions[i], decisions[j] = decisions[j], decisions[i]
			}
		}
	}
	return decisions, nil
}

// FindBySlug returns the decision whose filename slug matches the given slug, or an error if not found.
// Both legacy (NNNN-slug.md) and new (YYYY-MM-DD-slug.md) formats are searched recursively.
func FindBySlug(dir, slug string) (*Decision, error) {
	var found *Decision
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !isDocumentFile(d.Name()) {
			return nil
		}
		if slugFromFilename(d.Name()) == slug {
			doc, parseErr := Parse(path)
			if parseErr != nil {
				return parseErr
			}
			found = doc
			return filepath.SkipAll
		}
		return nil
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
