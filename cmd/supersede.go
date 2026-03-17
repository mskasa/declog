package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mskasa/declog/internal/decision"
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

		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		dir := decisionsDir(root, loadCfg())

		old, err := decision.FindByID(dir, oldID)
		if err != nil {
			return fmt.Errorf("ADR %04d not found", oldID)
		}

		if err := decision.CheckSupersedable(old); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Superseding ADR:\n  [%04d] %s\n\n", old.ID, old.Title)
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

		newID, err := decision.NextID(dir)
		if err != nil {
			return err
		}

		newPath, err := decision.Create(dir, newTitle, oldID)
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
