package cmd

import (
	"fmt"
	"os"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:          "lint",
	Short:        "Validate document structure and report issues",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:         cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := auditDirs(root, cfg)

		var issues []*decision.LintIssue
		for _, dir := range dirs {
			r, err := decision.Lint(dir, root)
			if err != nil {
				return err
			}
			issues = append(issues, r...)
		}

		if len(issues) == 0 {
			fmt.Fprintln(os.Stdout, "All documents pass lint checks. ✅")
			return nil
		}

		for _, issue := range issues {
			fmt.Fprintf(os.Stdout, "%s: %s\n", issue.File, issue.Message)
		}
		fmt.Fprintf(os.Stdout, "\n%d issue(s) found.\n", len(issues))
		return fmt.Errorf("lint failed: %d issue(s) found", len(issues))
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
