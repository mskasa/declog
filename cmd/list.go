package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mskasa/declog/internal/decision"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all decision records in reverse chronological order",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		dir := filepath.Join(root, "docs", "decisions")

		decisions, err := decision.List(dir)
		if err != nil {
			return err
		}
		if len(decisions) == 0 {
			fmt.Fprintln(os.Stdout, "No decisions found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tDate\tStatus\tTitle")
		fmt.Fprintln(w, "--\t----\t------\t-----")
		for _, d := range decisions {
			fmt.Fprintf(w, "%04d\t%s\t%s\t%s\n", d.ID, d.Date, d.Status, d.Title)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
