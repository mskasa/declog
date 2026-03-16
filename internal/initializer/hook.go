package initializer

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:embed templates/pre-commit
var hookScript string

// InstallHook creates .git/hooks/pre-commit with the declog ADR warning script.
// If the hook already exists, a warning with the script content is printed instead.
func InstallHook(root string, output io.Writer) error {
	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")

	if _, err := os.Stat(hookPath); err == nil {
		fmt.Fprintf(output, "  ⚠️  pre-commit hook already exists.\n")
		fmt.Fprintf(output, "    To add declog hook manually, append the following to .git/hooks/pre-commit:\n\n")
		fmt.Fprintln(output, hookScript)
		return nil
	}

	hooksDir := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("creating .git/hooks/: %w", err)
	}

	if err := os.WriteFile(hookPath, []byte(hookScript), 0o755); err != nil {
		return fmt.Errorf("writing pre-commit hook: %w", err)
	}

	fmt.Fprintf(output, "  ✅ Created .git/hooks/pre-commit\n")
	return nil
}
