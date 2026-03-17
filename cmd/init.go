package cmd

import (
	"github.com/mskasa/kizami/internal/initializer"
	"github.com/spf13/cobra"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kizami in the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRoot()
		if err != nil {
			return err
		}
		i := &initializer.Initializer{
			Root:   root,
			Input:  os.Stdin,
			Output: os.Stdout,
		}
		return i.Run()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
