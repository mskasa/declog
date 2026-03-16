package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mskasa/declog/internal/decision"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Check that Related Files in ADRs still exist in the repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		dir := filepath.Join(root, "docs", "decisions")

		fmt.Fprintln(os.Stdout, "Checking Related Files in ADRs...")

		results, err := decision.Audit(dir, root)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Fprintln(os.Stdout, "All Related Files are up to date. ✅")
			return nil
		}

		fmt.Fprintf(os.Stdout, "\nStale file references detected:\n\n")
		for _, r := range results {
			rel, err := filepath.Rel(root, r.File)
			if err != nil {
				rel = r.File
			}
			fmt.Fprintf(os.Stdout, "  [%04d] %s\n", r.ID, rel)
			fmt.Fprintf(os.Stdout, "  Title: %s\n", r.Title)
			fmt.Fprintf(os.Stdout, "  Missing files:\n")
			for _, f := range r.MissingFiles {
				fmt.Fprintf(os.Stdout, "    ⚠️  %s\n", f)
			}
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintf(os.Stdout, "%d ADR(s) have stale file references.\n", len(results))
		fmt.Fprintln(os.Stdout, "Consider updating Related Files, marking as Inactive, or superseding them.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
}
