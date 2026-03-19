package cmd

import (
	"fmt"
	"os"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/mskasa/kizami/internal/search"
	"github.com/spf13/cobra"
)

var blameCmd = &cobra.Command{
	Use:   "blame <file>",
	Short: "Find decisions that mention the given file path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := documentDirs(root, cfg)

		seen := make(map[int]struct{})
		var decisions []*decision.Decision
		for _, dir := range dirs {
			d, err := search.Blame(dir, filePath)
			if err != nil {
				return err
			}
			for _, dec := range d {
				if _, ok := seen[dec.ID]; ok {
					continue
				}
				seen[dec.ID] = struct{}{}
				decisions = append(decisions, dec)
			}
		}
		if len(decisions) == 0 {
			fmt.Fprintf(os.Stdout, "No decisions found mentioning %q.\n", filePath)
			return nil
		}

		fmt.Fprintf(os.Stdout, "Found %d decision(s) mentioning %q:\n\n", len(decisions), filePath)
		for _, d := range decisions {
			fmt.Fprintf(os.Stdout, "[%04d] %s | %s\n", d.ID, d.Date, d.Status)
			fmt.Fprintf(os.Stdout, "Title: %s\n", d.Title)
			fmt.Fprintf(os.Stdout, "Path: %s\n\n", d.File)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(blameCmd)
}
