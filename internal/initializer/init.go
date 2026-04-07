package initializer

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed templates/adr-check.yml
var adrCheckWorkflow string

//go:embed templates/adr-audit.yml
var adrAuditWorkflow string

//go:embed templates/kizami-promote.yml
var promoteWorkflow string

const defaultConfigContent = `[ai]
model = "claude-sonnet-4-20250514"

[documents]
dirs = ["docs/decisions", "docs/design"]

[decisions]
dir = "docs/decisions"

[design]
dir = "docs/design"

[audit]
dirs = ["docs/decisions", "docs/design"]

[review]
months_threshold = 6

[editor]
command = "code --wait"
`

// Initializer handles the kizami initialization process.
type Initializer struct {
	Root   string
	Input  io.Reader
	Output io.Writer
	YesAll bool // when true, all prompts are auto-accepted
}

// prompt prints a y/n question and returns true if the user answers "y".
// When YesAll is set, it auto-answers "y" without reading input.
func (i *Initializer) prompt(scanner *bufio.Scanner, question string) bool {
	if i.YesAll {
		fmt.Fprintf(i.Output, "%s (y/n): y\n", question)
		return true
	}
	fmt.Fprintf(i.Output, "%s (y/n): ", question)
	if !scanner.Scan() {
		return false
	}
	return strings.TrimSpace(strings.ToLower(scanner.Text())) == "y"
}

// detectDefaultBranch returns the default branch name by inspecting the remote HEAD ref.
// Falls back to "main" if detection fails.
func detectDefaultBranch(root string) string {
	out, err := exec.Command("git", "-C", root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		// output is like "origin/main"
		parts := strings.SplitN(strings.TrimSpace(string(out)), "/", 2)
		if len(parts) == 2 && parts[1] != "" {
			return parts[1]
		}
	}
	// Fallback: use the current branch name
	out, err = exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		if b := strings.TrimSpace(string(out)); b != "" && b != "HEAD" {
			return b
		}
	}
	return "main"
}

// Run performs the initialization steps sequentially.
func (i *Initializer) Run() error {
	fmt.Fprintln(i.Output, "Initializing kizami...")

	if err := i.createDecisionsDir(); err != nil {
		return err
	}

	if err := i.createDesignDir(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(i.Input)

	if err := i.setupWorkflow(scanner); err != nil {
		return err
	}

	if err := i.setupHook(scanner); err != nil {
		return err
	}

	if err := i.setupAuditWorkflow(scanner); err != nil {
		return err
	}

	if err := i.setupPromoteWorkflow(scanner); err != nil {
		return err
	}

	if err := i.setupConfig(); err != nil {
		return err
	}

	fmt.Fprintln(i.Output)
	fmt.Fprintln(i.Output, `Done! Run `+"`"+`kizami adr "<title>"`+"`"+` to create your first decision.`)
	return nil
}

func (i *Initializer) createDecisionsDir() error {
	dir := filepath.Join(i.Root, "docs", "decisions")
	if _, err := os.Stat(dir); err == nil {
		fmt.Fprintf(i.Output, "  ⚠️  docs/decisions/ already exists. Skipping.\n")
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating docs/decisions/: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created docs/decisions/\n")
	return nil
}

func (i *Initializer) createDesignDir() error {
	dir := filepath.Join(i.Root, "docs", "design")
	if _, err := os.Stat(dir); err == nil {
		fmt.Fprintf(i.Output, "  ⚠️  docs/design/ already exists. Skipping.\n")
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating docs/design/: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created docs/design/\n")
	return nil
}

func (i *Initializer) setupWorkflow(scanner *bufio.Scanner) error {
	if !i.prompt(scanner, "Add GitHub Actions ADR check workflow?") {
		return nil
	}

	workflowDir := filepath.Join(i.Root, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return fmt.Errorf("creating .github/workflows/: %w", err)
	}

	workflowPath := filepath.Join(workflowDir, "adr-check.yml")
	if err := os.WriteFile(workflowPath, []byte(adrCheckWorkflow), 0o644); err != nil {
		return fmt.Errorf("writing adr-check.yml: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created .github/workflows/adr-check.yml\n")
	return nil
}

func (i *Initializer) setupHook(scanner *bufio.Scanner) error {
	if !i.prompt(scanner, "Add pre-commit hook to prompt for a decision record?") {
		return nil
	}

	return InstallHook(i.Root, i.Output)
}

func (i *Initializer) setupConfig() error {
	configPath := filepath.Join(i.Root, "kizami.toml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Fprintf(i.Output, "  ⚠️  kizami.toml already exists. Skipping.\n")
		return nil
	}

	if err := os.WriteFile(configPath, []byte(defaultConfigContent), 0o644); err != nil {
		return fmt.Errorf("writing kizami.toml: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created kizami.toml\n")
	return nil
}

func (i *Initializer) setupPromoteWorkflow(scanner *bufio.Scanner) error {
	if !i.prompt(scanner, "Add auto-promote workflow (Draft → Active on push to default branch)?") {
		return nil
	}

	workflowDir := filepath.Join(i.Root, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return fmt.Errorf("creating .github/workflows/: %w", err)
	}

	workflowPath := filepath.Join(workflowDir, "kizami-promote.yml")
	if _, err := os.Stat(workflowPath); err == nil {
		fmt.Fprintf(i.Output, "  ⚠️  .github/workflows/kizami-promote.yml already exists. Skipping.\n")
		return nil
	}

	branch := detectDefaultBranch(i.Root)
	content := strings.ReplaceAll(promoteWorkflow, "{{DEFAULT_BRANCH}}", branch)
	if err := os.WriteFile(workflowPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing kizami-promote.yml: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created .github/workflows/kizami-promote.yml (branch: %s)\n", branch)
	return nil
}

func (i *Initializer) setupAuditWorkflow(scanner *bufio.Scanner) error {
	if !i.prompt(scanner, "Add weekly audit workflow?") {
		return nil
	}

	workflowDir := filepath.Join(i.Root, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return fmt.Errorf("creating .github/workflows/: %w", err)
	}

	workflowPath := filepath.Join(workflowDir, "adr-audit.yml")
	if _, err := os.Stat(workflowPath); err == nil {
		fmt.Fprintf(i.Output, "  ⚠️  .github/workflows/adr-audit.yml already exists. Skipping.\n")
		return nil
	}

	if err := os.WriteFile(workflowPath, []byte(adrAuditWorkflow), 0o644); err != nil {
		return fmt.Errorf("writing adr-audit.yml: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created .github/workflows/adr-audit.yml\n")
	return nil
}
