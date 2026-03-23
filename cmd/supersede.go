package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var supersedeCmd = &cobra.Command{
	Use:   "supersede <slug> <title>",
	Short: "Supersede an existing decision and create a new one",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldSlug := args[0]
		newTitle := args[1]

		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()

		old, err := findBySlug(root, cfg, oldSlug)
		if err != nil {
			return err
		}

		if err := decision.CheckSupersedable(old); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Superseding decision:\n  [%s] %s\n\n", old.Slug, old.Title)
		fmt.Fprint(os.Stdout, "Supersede this decision? (y/n): ")

		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Fprintln(os.Stdout, "Aborted.")
			return nil
		}

		// New superseding document is always created in the same directory as the old one.
		newDir := decisionsDir(root, cfg)
		newPath, err := decision.Create(newDir, newTitle, oldSlug)
		if err != nil {
			return err
		}

		newSlug := decision.Slugify(newTitle)
		status := "Superseded by " + newSlug
		if err := decision.UpdateStatus(old.File, status, ""); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "\nUpdated: %s\n  Status: %s\n\n", old.File, status)
		fmt.Fprintf(os.Stdout, "Created: %s\n", newPath)

		return openEditor(newPath)
	},
}

func init() {
	rootCmd.AddCommand(supersedeCmd)
}
