package cmd

import (
	"fmt"
	"os"
	"github.com/mskasa/kizami/internal/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search decisions by keyword",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		dir := decisionsDir(root, loadCfg())

		results, err := search.Run(dir, args[0])
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Fprintln(os.Stdout, "No matches found.")
			return nil
		}

		for _, r := range results {
			fmt.Fprintf(os.Stdout, "%s:%d: %s\n", r.File, r.Line, r.Text)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
