package cli

import (
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/options"
	"go.mattglei.ch/timber"
)

var Root = &cobra.Command{
	Use:     "synthient",
	Version: "v1.7.0",
	Short:   "Official CLI tool for Synthient",
	Long:    "Synthient's official CLI for account inspection, IP and domain lookups, gRPC schemas, real-time feed streams, and Parquet snapshot downloads.",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			timber.Fatal(err, "output help")
		}
	},
}

func init() {
	Root.PersistentFlags().
		StringVar(&options.ConfigPath, "config", "", "Path to config file")
	Root.PersistentFlags().
		StringVar(&options.Profile, "profile", "", "Config profile to use")
	Root.PersistentFlags().
		BoolVar(&options.NoColor, "no-color", false, "Disable color output")
	Root.PersistentFlags().
		BoolVarP(&options.Quiet, "quiet", "q", false, "Suppress human-oriented progress output")
}
