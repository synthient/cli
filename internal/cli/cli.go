package cli

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

func root(cmd *cobra.Command, args []string) {
	err := cmd.Help()
	if err != nil {
		timber.Fatal(err, "output help")
	}
}

var rootCmd = &cobra.Command{
	Use:   "synthient",
	Short: "Official CLI tool for Synthient",
	Run:   root,
}

func Execute() {
	rootCmd.Execute() // nolint:errcheck
}
