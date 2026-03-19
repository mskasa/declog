package cmd

import (
	"fmt"
	"github.com/mskasa/kizami/internal/search"
	"github.com/spf13/cobra"
	"os"
)

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search decisions by keyword",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := documentDirs(root, cfg)

		var allResults []search.Result
		for _, dir := range dirs {
			r, err := search.Run(dir, args[0])
			if err != nil {
				return err
			}
			allResults = append(allResults, r...)
		}
		if len(allResults) == 0 {
			fmt.Fprintln(os.Stdout, "No matches found.")
			return nil
		}

		for _, r := range allResults {
			fmt.Fprintf(os.Stdout, "%s:%d: %s\n", r.File, r.Line, r.Text)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
