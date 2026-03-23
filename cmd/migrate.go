package cmd

import (
	"fmt"
	"os"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate legacy NNNN-slug.md files to YYYY-MM-DD-slug.md format",
	Long: `Rename legacy NNNN-slug.md files to YYYY-MM-DD-slug.md and update internal
references (headings, Supersedes, Superseded by) to use slugs instead of numeric IDs.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()

		total := 0
		for _, dir := range documentDirs(root, cfg) {
			n, err := decision.MigrateLegacyFiles(dir)
			if err != nil {
				return fmt.Errorf("migrating %s: %w", dir, err)
			}
			if n > 0 {
				fmt.Fprintf(os.Stdout, "Migrated %d file(s) in %s\n", n, dir)
			}
			total += n
		}

		if total == 0 {
			fmt.Fprintln(os.Stdout, "No legacy files found to migrate.")
		} else {
			fmt.Fprintf(os.Stdout, "Migration complete: %d file(s) renamed.\n", total)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
