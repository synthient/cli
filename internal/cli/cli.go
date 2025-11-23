package cli

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

var rootCmd = &cobra.Command{
	Use:   "synthient",
	Short: "Official CLI tool for Synthient [https://github.com/synthient/cli]",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			timber.Fatal(err, "output help")
		}
	},
}

func Execute() {
	rootCmd.Execute() // nolint:errcheck
}
