package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mskasa/declog/internal/decision"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log <title>",
	Short: "Create a new decision record and open it in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		dir := filepath.Join(root, "docs", "decisions")

		path, err := decision.Create(dir, args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Created: %s\n", path)

		return openEditor(path)
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}

func gitRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}

	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
