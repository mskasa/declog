package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	kizamicontext "github.com/mskasa/kizami/internal/context"
	"github.com/spf13/cobra"
)

var (
	contextJSON bool
	contextFull bool
)

var contextCmd = &cobra.Command{
	Use:   "context <files...>",
	Short: "Show which decisions govern the given files",
	Long: `Show which decisions (ADRs and design documents) govern the given files.

Resolves each file against every document's Related Files section (exact path,
directory prefix, or glob) and reports the governing decisions, their drift
status, and any queried files that matched no decision.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := documentDirs(root, cfg)

		result, err := kizamicontext.Resolve(dirs, root, args, contextFull)
		if err != nil {
			return err
		}

		if contextJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		printContextResult(result)
		return nil
	},
}

func printContextResult(result *kizamicontext.Result) {
	if len(result.Decisions) == 0 {
		fmt.Fprintln(os.Stdout, "No decisions govern the given file(s).")
	} else {
		fmt.Fprintf(os.Stdout, "Found %d decision(s) governing %d file(s):\n\n", len(result.Decisions), len(result.Query))
		for _, d := range result.Decisions {
			status := d.Status
			if d.SupersededBy != "" {
				status = fmt.Sprintf("%s (see %s)", d.Status, d.SupersededBy)
			}
			fmt.Fprintf(os.Stdout, "[%s] %s | %s\n", d.Slug, d.Date, status)
			fmt.Fprintf(os.Stdout, "Title: %s\n", d.Title)
			for _, m := range d.Matched {
				fmt.Fprintf(os.Stdout, "Matched: %s (%s: %s)\n", m.File, m.Kind, m.Rule)
			}
			fmt.Fprintf(os.Stdout, "Drift: %s\n", d.Drift.State)
			if len(d.Drift.Missing) > 0 {
				fmt.Fprintf(os.Stdout, "  Missing: %v\n", d.Drift.Missing)
			}
			if d.Decision != "" {
				fmt.Fprintf(os.Stdout, "Decision: %s\n", d.Decision)
			}
			fmt.Fprintf(os.Stdout, "Path: %s\n\n", d.Path)
		}
	}

	if len(result.Unmatched) > 0 {
		fmt.Fprintln(os.Stdout, "Unmatched files:")
		for _, f := range result.Unmatched {
			fmt.Fprintf(os.Stdout, "  %s\n", f)
		}
	}
}

func init() {
	contextCmd.Flags().BoolVar(&contextJSON, "json", false, "Output as JSON")
	contextCmd.Flags().BoolVar(&contextFull, "full", false, "Return each decision's full document content instead of just its summary section")
	rootCmd.AddCommand(contextCmd)
}
