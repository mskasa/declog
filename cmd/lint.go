package cmd

import (
	"fmt"
	"os"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:           "lint",
	Short:         "Validate document structure and report issues",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.NoArgs,
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

		var errCount, warnCount int
		for _, issue := range issues {
			if issue.Severity == "warning" {
				fmt.Fprintf(os.Stdout, "[warn]  %s: %s\n", issue.File, issue.Message)
				warnCount++
			} else {
				fmt.Fprintf(os.Stdout, "[error] %s: %s\n", issue.File, issue.Message)
				errCount++
			}
		}

		fmt.Fprintln(os.Stdout)
		switch {
		case errCount > 0 && warnCount > 0:
			fmt.Fprintf(os.Stdout, "%d error(s), %d warning(s) found.\n", errCount, warnCount)
		case errCount > 0:
			fmt.Fprintf(os.Stdout, "%d error(s) found.\n", errCount)
		default:
			fmt.Fprintf(os.Stdout, "%d warning(s) found. ⚠️\n", warnCount)
		}

		if errCount > 0 {
			return fmt.Errorf("lint failed: %d error(s) found", errCount)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
