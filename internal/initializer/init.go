package initializer

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
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
}

// Run performs the initialization steps sequentially.
func (i *Initializer) Run() error {
	fmt.Fprintln(i.Output, "Initializing kizami...")

	if err := i.createDecisionsDir(); err != nil {
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

func (i *Initializer) setupWorkflow(scanner *bufio.Scanner) error {
	fmt.Fprintf(i.Output, "Add GitHub Actions ADR check workflow? (y/n): ")

	if !scanner.Scan() {
		return nil
	}
	if strings.TrimSpace(strings.ToLower(scanner.Text())) != "y" {
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
	fmt.Fprintf(i.Output, "Add pre-commit hook to prompt for a decision record? (y/n): ")

	if !scanner.Scan() {
		return nil
	}
	if strings.TrimSpace(strings.ToLower(scanner.Text())) != "y" {
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
	fmt.Fprintf(i.Output, "Add auto-promote workflow (Draft → Active on push to main)? (y/n): ")

	if !scanner.Scan() {
		return nil
	}
	if strings.TrimSpace(strings.ToLower(scanner.Text())) != "y" {
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

	if err := os.WriteFile(workflowPath, []byte(promoteWorkflow), 0o644); err != nil {
		return fmt.Errorf("writing kizami-promote.yml: %w", err)
	}
	fmt.Fprintf(i.Output, "  ✅ Created .github/workflows/kizami-promote.yml\n")
	return nil
}

func (i *Initializer) setupAuditWorkflow(scanner *bufio.Scanner) error {
	fmt.Fprintf(i.Output, "Add weekly audit workflow? (y/n): ")

	if !scanner.Scan() {
		return nil
	}
	if strings.TrimSpace(strings.ToLower(scanner.Text())) != "y" {
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
