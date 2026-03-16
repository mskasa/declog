package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mskasa/declog/internal/decision"
	"github.com/spf13/cobra"
)

var reviewMonths int

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "List Active ADRs that have not been updated recently",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRoot()
		if err != nil {
			return fmt.Errorf("not a git repository")
		}
		dir := filepath.Join(root, "docs", "decisions")

		stale, err := decision.StaleADRs(dir, reviewMonths)
		if err != nil {
			return err
		}

		if len(stale) == 0 {
			fmt.Fprintf(os.Stdout, "All ADRs have been reviewed within the last %d months. ✅\n", reviewMonths)
			return nil
		}

		now := time.Now()
		fmt.Fprintf(os.Stdout, "ADRs not reviewed in %d+ months:\n\n", reviewMonths)
		for _, s := range stale {
			rel, err := filepath.Rel(root, s.File)
			if err != nil {
				rel = s.File
			}
			ago := decision.MonthsAgo(s.LastUpdated, now)
			fmt.Fprintf(os.Stdout, "  [%04d] Last updated: %s (%d months ago)\n", s.ID, s.LastUpdated.Format("2006-01-02"), ago)
			fmt.Fprintf(os.Stdout, "  Title: %s\n", s.Title)
			fmt.Fprintf(os.Stdout, "  Path: %s\n\n", rel)
		}
		fmt.Fprintf(os.Stdout, "%d ADR(s) need review.\n", len(stale))
		fmt.Fprintln(os.Stdout, "Consider updating, marking as Inactive, or superseding them.")
		return nil
	},
}

func init() {
	reviewCmd.Flags().IntVar(&reviewMonths, "months", 6, "Number of months without update to consider stale")
	rootCmd.AddCommand(reviewCmd)
}
