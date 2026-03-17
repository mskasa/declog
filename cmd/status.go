package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mskasa/declog/internal/decision"
	"github.com/spf13/cobra"
)

var supersededBy int

var statusCmd = &cobra.Command{
	Use:   "status <id> <status>",
	Short: "Update the status of a decision record",
	Long: `Update the status of a decision record.

Valid statuses: Proposed, Accepted, Superseded, Deprecated

Examples:
  why status 3 accepted
  why status 3 superseded --by 5`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil || id < 1 {
			return fmt.Errorf("id must be a positive integer")
		}

		status, err := decision.NormalizeStatus(args[1])
		if err != nil {
			return err
		}

		if supersededBy > 0 && !strings.EqualFold(status, "Superseded") {
			return fmt.Errorf("--by flag is only valid with status 'superseded'")
		}

		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		dir := decisionsDir(root, loadCfg())

		d, err := decision.FindByID(dir, id)
		if err != nil {
			return err
		}

		if err := decision.UpdateStatus(d.File, status, supersededBy); err != nil {
			return err
		}

		msg := fmt.Sprintf("Updated %04d: Status → %s", id, status)
		if supersededBy > 0 {
			msg += fmt.Sprintf(" (superseded by %04d)", supersededBy)
		}
		fmt.Fprintln(os.Stdout, msg)
		return nil
	},
}

func init() {
	statusCmd.Flags().IntVar(&supersededBy, "by", 0, "ID of the decision that supersedes this one")
	rootCmd.AddCommand(statusCmd)
}
