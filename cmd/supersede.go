package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mskasa/kizami/internal/decision"
	"github.com/spf13/cobra"
)

var supersedeCmd = &cobra.Command{
	Use:   "supersede <id> <title>",
	Short: "Supersede an existing decision and create a new one",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldID, err := strconv.Atoi(args[0])
		if err != nil || oldID < 1 {
			return fmt.Errorf("id must be a positive integer")
		}
		newTitle := args[1]

		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		cfg := loadCfg()

		old, err := findByID(root, cfg, oldID)
		if err != nil {
			return err
		}

		if err := decision.CheckSupersedable(old); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Superseding decision:\n  [%04d] %s\n\n", old.ID, old.Title)
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
		newID, err := decision.NextID(newDir)
		if err != nil {
			return err
		}

		newPath, err := decision.Create(newDir, newTitle, oldID)
		if err != nil {
			return err
		}

		status := fmt.Sprintf("Superseded by %04d", newID)
		if err := decision.UpdateStatus(old.File, status, 0); err != nil {
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
