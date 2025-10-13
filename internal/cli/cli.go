package cli

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		timber.Fatal(err, "execute root command")
	}
}

var rootCmd = &cobra.Command{
	Use:   "synthient",
	Short: "official cli tool for synthient",
	Run:   root,
}

func root(cmd *cobra.Command, args []string) {
	err := cmd.Help()
	if err != nil {
		timber.Fatal(err, "output help")
	}
}
