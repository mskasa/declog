package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Display a single decision record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}

		d, err := findBySlug(root, loadCfg(), args[0])
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
