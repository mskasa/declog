package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var statusFilter string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all decision records in reverse chronological order",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()
		dirs := documentDirs(root, cfg)

		var decisions []*decision.Decision
		for _, dir := range dirs {
			d, err := decision.List(dir)
			if err != nil {
				return err
			}
			decisions = append(decisions, d...)
		}
		sort.Slice(decisions, func(i, j int) bool {
			if decisions[i].Date != decisions[j].Date {
				return decisions[i].Date > decisions[j].Date
			}
			return decisions[i].ID > decisions[j].ID
		})

		if statusFilter != "" {
			filter := strings.ToLower(statusFilter)
			filtered := decisions[:0]
			for _, d := range decisions {
				if strings.HasPrefix(strings.ToLower(d.Status), filter) {
					filtered = append(filtered, d)
				}
			}
			decisions = filtered
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
	listCmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (e.g. active, inactive, superseded)")
	rootCmd.AddCommand(listCmd)
}
