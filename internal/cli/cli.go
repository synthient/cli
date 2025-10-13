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
	err := rootCmd.Execute()
	if err != nil {
		timber.Fatal(err, "execute root command")
	}
}
