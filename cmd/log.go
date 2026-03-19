package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/mskasa/kizami/internal/ai"
	"github.com/mskasa/kizami/internal/config"
	"github.com/mskasa/kizami/internal/decision"
	"github.com/mskasa/kizami/internal/search"
	"github.com/spf13/cobra"
)

var (
	aiFlag     bool
	modelFlag  string
	dryRunFlag bool
)

var adrCmd = &cobra.Command{
	Use:   "adr <title>",
	Short: "Create a new ADR and open it in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		dir := decisionsDir(root, loadCfg())

		supersededID, err := promptSimilar(dir, args[0])
		if err != nil {
			return err
		}

		newID, err := decision.NextID(dir)
		if err != nil {
			return err
		}

		var path string
		if aiFlag {
			path, err = runWithAI(dir, root, args[0], supersededID)
		} else {
			path, err = decision.Create(dir, args[0], supersededID)
		}
		if err != nil {
			return err
		}

		if supersededID > 0 {
			old, err := decision.FindByID(dir, supersededID)
			if err != nil {
				return err
			}
			status := fmt.Sprintf("Superseded by %04d", newID)
			if err := decision.UpdateStatus(old.File, status, 0); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Updated: %s\n  Status: %s\n\n", old.File, status)
		}

		fmt.Fprintf(os.Stdout, "Creating new ADR...\nCreated: %s\n", path)
		return openEditor(path)
	},
}

var designCmd = &cobra.Command{
	Use:   "design <title>",
	Short: "Create a new design document and open it in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dir := designDir(root, cfg)

		supersededID, err := promptSimilar(dir, args[0])
		if err != nil {
			return err
		}

		newID, err := decision.NextID(dir)
		if err != nil {
			return err
		}

		path, err := decision.CreateDesign(dir, args[0], supersededID)
		if err != nil {
			return err
		}

		if supersededID > 0 {
			old, err := decision.FindByID(dir, supersededID)
			if err != nil {
				return err
			}
			status := fmt.Sprintf("Superseded by %04d", newID)
			if err := decision.UpdateStatus(old.File, status, 0); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Updated: %s\n  Status: %s\n\n", old.File, status)
		}

		fmt.Fprintf(os.Stdout, "Creating new design document...\nCreated: %s\n", path)
		return openEditor(path)
	},
}

// runWithAI generates an ADR draft via the Anthropic API and writes the file.
func runWithAI(dir, root, title string, supersededID int) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set.\nPlease set the environment variable and try again.\n\n  export ANTHROPIC_API_KEY=your-api-key")
	}

	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}
	model := config.ResolveModel(modelFlag, cfg)

	input := ai.GatherInput(root, title)
	prompt := ai.BuildPrompt(input)

	if dryRunFlag {
		if !ai.DryRun(prompt, os.Stdin, os.Stdout) {
			return "", fmt.Errorf("aborted")
		}
	}

	fmt.Fprintln(os.Stdout, "Generating ADR draft with AI...")
	draft, err := ai.GenerateDraft(prompt, model, apiKey)
	if err != nil {
		return "", err
	}

	return decision.CreateFromDraft(dir, title, draft, supersededID)
}

// promptSimilar searches for similar documents and asks the user if any should be superseded.
// Returns the ID to supersede (0 if none).
func promptSimilar(dir, title string) (int, error) {
	fmt.Fprintln(os.Stdout, "Searching for similar decisions...")

	similar, err := search.Similar(dir, title)
	if err != nil {
		// Non-fatal: skip if search fails (e.g. docs/decisions doesn't exist yet).
		fmt.Fprintf(os.Stdout, "No similar decisions found.\n\n")
		return 0, nil
	}
	if len(similar) == 0 {
		fmt.Fprintf(os.Stdout, "No similar decisions found.\n\n")
		return 0, nil
	}

	fmt.Fprintln(os.Stdout, "Similar decisions found:")
	for _, d := range similar {
		fmt.Fprintf(os.Stdout, "  [%04d] %s | %s\n  Title: %s\n\n", d.ID, d.Date, d.Status, d.Title)
	}

	fmt.Fprintln(os.Stdout, "Does this supersede any of the above?")
	fmt.Fprint(os.Stdout, "Enter ID to supersede (or press Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, nil
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, nil
	}

	id, err := strconv.Atoi(line)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid ID: %q", line)
	}
	return id, nil
}

func init() {
	rootCmd.AddCommand(adrCmd)
	adrCmd.Flags().BoolVar(&aiFlag, "ai", false, "Generate ADR draft using Anthropic API")
	adrCmd.Flags().StringVar(&modelFlag, "model", "", "Anthropic model to use (overrides config file)")
	adrCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show the prompt to be sent to the API without calling it")

	rootCmd.AddCommand(designCmd)
}

func gitRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// gitRepoRootFn is the function used to locate the git repository root.
// It can be overridden in tests.
var gitRepoRootFn = gitRepoRoot

func openEditor(path string) error {
	// Priority: EDITOR env > VISUAL env > config editor.command > platform default
	var parts []string
	if e := os.Getenv("EDITOR"); e != "" {
		parts = []string{e, path}
	} else if e := os.Getenv("VISUAL"); e != "" {
		parts = []string{e, path}
	} else if cmd := loadCfg().Editor.Command; cmd != "" {
		parts = append(strings.Fields(cmd), path)
	} else if runtime.GOOS == "windows" {
		parts = []string{"notepad", path}
	} else {
		parts = []string{"vi", path}
	}

	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
