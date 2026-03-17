package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Display a single decision record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil || id < 1 {
			return fmt.Errorf("id must be a positive integer")
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

		content, err := os.ReadFile(d.File)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		fmt.Fprint(os.Stdout, string(content))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
