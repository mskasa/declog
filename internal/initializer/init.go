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

// Initializer handles the declog initialization process.
type Initializer struct {
	Root   string
	Input  io.Reader
	Output io.Writer
}

// Run performs the initialization steps sequentially.
func (i *Initializer) Run() error {
	fmt.Fprintln(i.Output, "Initializing declog...")

	if err := i.createDecisionsDir(); err != nil {
		return err
	}

	if err := i.setupWorkflow(); err != nil {
		return err
	}

	fmt.Fprintln(i.Output)
	fmt.Fprintln(i.Output, `Done! Run `+"`"+`why log "<title>"`+"`"+` to create your first decision.`)
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

func (i *Initializer) setupWorkflow() error {
	fmt.Fprintf(i.Output, "Add GitHub Actions ADR check workflow? (y/n): ")

	scanner := bufio.NewScanner(i.Input)
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
