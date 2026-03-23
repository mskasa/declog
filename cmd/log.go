package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

		supersededSlug, err := promptSimilar(dir, args[0])
		if err != nil {
			return err
		}

		var path string
		if aiFlag {
			path, err = runWithAI(dir, root, args[0], supersededSlug)
		} else {
			path, err = decision.Create(dir, args[0], supersededSlug)
		}
		if err != nil {
			return err
		}

		if supersededSlug != "" {
			old, err := decision.FindBySlug(dir, supersededSlug)
			if err != nil {
				return err
			}
			newSlug := decision.Slugify(args[0])
			status := "Superseded by " + newSlug
			if err := decision.UpdateStatus(old.File, status, ""); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Updated: %s\n  Status: %s\n\n", old.File, status)
		}

		fmt.Fprintf(os.Stdout, "Creating new ADR...\nCreated: %s\n", path)
		return openEditor(path)
	},
}

var (
	designAIFlag     bool
	designDryRunFlag bool
)

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

		supersededSlug, err := promptSimilar(dir, args[0])
		if err != nil {
			return err
		}

		var path string
		if designAIFlag {
			path, err = runWithAIDesign(dir, root, args[0], supersededSlug)
		} else {
			path, err = decision.CreateDesign(dir, args[0], supersededSlug)
		}
		if err != nil {
			return err
		}

		if supersededSlug != "" {
			old, err := decision.FindBySlug(dir, supersededSlug)
			if err != nil {
				return err
			}
			newSlug := decision.Slugify(args[0])
			status := "Superseded by " + newSlug
			if err := decision.UpdateStatus(old.File, status, ""); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Updated: %s\n  Status: %s\n\n", old.File, status)
		}

		fmt.Fprintf(os.Stdout, "Creating new design document...\nCreated: %s\n", path)
		return openEditor(path)
	},
}

// runWithAIDesign generates a design document draft via the Anthropic API and writes the file.
func runWithAIDesign(dir, root, title, supersededSlug string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set.\nPlease set the environment variable and try again.\n\n  export ANTHROPIC_API_KEY=your-api-key")
	}

	cfg, err := config.Load(root)
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}
	model := config.ResolveModel(modelFlag, cfg)

	input := ai.GatherInput(root, title)
	prompt := ai.BuildDesignPrompt(input)

	if designDryRunFlag {
		if !ai.DryRun(prompt, os.Stdin, os.Stdout) {
			return "", fmt.Errorf("aborted")
		}
	}

	fmt.Fprintln(os.Stdout, "Generating design document draft with AI...")
	draft, err := ai.GenerateDraft(prompt, model, apiKey)
	if err != nil {
		return "", err
	}

	return decision.CreateDesignFromDraft(dir, title, draft, supersededSlug)
}

// runWithAI generates an ADR draft via the Anthropic API and writes the file.
func runWithAI(dir, root, title, supersededSlug string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set.\nPlease set the environment variable and try again.\n\n  export ANTHROPIC_API_KEY=your-api-key")
	}

	cfg, err := config.Load(root)
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

	return decision.CreateFromDraft(dir, title, draft, supersededSlug)
}

// promptSimilar searches for similar documents and asks the user if any should be superseded.
// Returns the slug to supersede (empty string if none).
func promptSimilar(dir, title string) (string, error) {
	fmt.Fprintln(os.Stdout, "Searching for similar decisions...")

	similar, err := search.Similar(dir, title)
	if err != nil {
		// Non-fatal: skip if search fails (e.g. docs/decisions doesn't exist yet).
		fmt.Fprintf(os.Stdout, "No similar decisions found.\n\n")
		return "", nil
	}
	if len(similar) == 0 {
		fmt.Fprintf(os.Stdout, "No similar decisions found.\n\n")
		return "", nil
	}

	fmt.Fprintln(os.Stdout, "Similar decisions found:")
	for _, d := range similar {
		fmt.Fprintf(os.Stdout, "  [%s] %s | %s\n  Title: %s\n\n", d.Slug, d.Date, d.Status, d.Title)
	}

	fmt.Fprintln(os.Stdout, "Does this supersede any of the above?")
	fmt.Fprint(os.Stdout, "Enter slug to supersede (or press Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil
	}
	line = strings.TrimSpace(line)
	return line, nil
}

func init() {
	rootCmd.AddCommand(adrCmd)
	adrCmd.Flags().BoolVar(&aiFlag, "ai", false, "Generate ADR draft using Anthropic API")
	adrCmd.Flags().StringVar(&modelFlag, "model", "", "Anthropic model to use (overrides config file)")
	adrCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show the prompt to be sent to the API without calling it")

	rootCmd.AddCommand(designCmd)
	designCmd.Flags().BoolVar(&designAIFlag, "ai", false, "Generate design document draft using Anthropic API")
	designCmd.Flags().StringVar(&modelFlag, "model", "", "Anthropic model to use (overrides config file)")
	designCmd.Flags().BoolVar(&designDryRunFlag, "dry-run", false, "Show the prompt to be sent to the API without calling it")
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
