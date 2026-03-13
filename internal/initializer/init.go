package initializer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const adrCheckWorkflow = `name: ADR Check

on:
  pull_request:
    types: [opened, edited, synchronize]

jobs:
  adr-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Check for ADR
        env:
          PR_TITLE: ${{ github.event.pull_request.title }}
          PR_BODY: ${{ github.event.pull_request.body }}
          BASE_SHA: ${{ github.event.pull_request.base.sha }}
          HEAD_SHA: ${{ github.event.pull_request.head.sha }}
        run: |
          if echo "$PR_TITLE" | grep -qF '[skip-adr]'; then
            echo "✅ [skip-adr] found in PR title. Skipping ADR check."
            exit 0
          fi

          if echo "$PR_BODY" | grep -qF 'docs/decisions/'; then
            echo "✅ docs/decisions/ referenced in PR body."
            exit 0
          fi

          if git diff --name-only "$BASE_SHA" "$HEAD_SHA" | grep -q '^docs/decisions/.*\.md$'; then
            echo "✅ ADR file found in changed files."
            exit 0
          fi

          echo "::warning::No ADR found. Consider adding a decision record to docs/decisions/ or include [skip-adr] in the PR title."
`

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
