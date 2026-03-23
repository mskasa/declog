package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var supersededBySlug string

var statusCmd = &cobra.Command{
	Use:   "status <slug> <status>",
	Short: "Update the status of a decision record",
	Long: `Update the status of a decision record.

Valid statuses: Proposed, Accepted, Superseded, Deprecated

Examples:
  kizami status use-postgresql accepted
  kizami status use-postgresql superseded --by use-cockroachdb`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := decision.NormalizeStatus(args[1])
		if err != nil {
			return err
		}

		if supersededBySlug != "" && !strings.EqualFold(status, "Superseded") {
			return fmt.Errorf("--by flag is only valid with status 'superseded'")
		}

		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}

		d, err := findBySlug(root, loadCfg(), args[0])
		if err != nil {
			return err
		}

		if err := decision.UpdateStatus(d.File, status, supersededBySlug); err != nil {
			return err
		}

		msg := fmt.Sprintf("Updated %s: Status → %s", args[0], status)
		if supersededBySlug != "" {
			msg += fmt.Sprintf(" (superseded by %s)", supersededBySlug)
		}
		fmt.Fprintln(os.Stdout, msg)
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&supersededBySlug, "by", "", "Slug of the document that supersedes this one")
	rootCmd.AddCommand(statusCmd)
}
