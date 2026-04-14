package cmd

import (
	"fmt"
	"os"
	"strings"

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

		matches, err := findAllBySlug(root, loadCfg(), args[0])
		if err != nil {
			return err
		}

		for i, d := range matches {
			if len(matches) > 1 {
				fmt.Fprintf(os.Stdout, "=== %s ===\n", d.File)
			}
			content, err := os.ReadFile(d.File)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			fmt.Fprint(os.Stdout, string(content))
			if len(matches) > 1 && i < len(matches)-1 {
				fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
