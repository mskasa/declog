package cmd

import (
	"os"

	"github.com/mskasa/kizami/internal/initializer"
	"github.com/spf13/cobra"
)

var initYesAll bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize kizami in the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := gitRepoRootFn()
		if err != nil {
			return err
		}
		i := &initializer.Initializer{
			Root:   root,
			Input:  os.Stdin,
			Output: os.Stdout,
			YesAll: initYesAll,
		}
		return i.Run()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initYesAll, "yes", "y", false, "Accept all prompts non-interactively")
	rootCmd.AddCommand(initCmd)
}
