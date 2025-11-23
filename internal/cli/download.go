package cli

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

var downloadCmd = &cobra.Command{
	Use:    "download",
	Short:  "Download a stream",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		timber.Debug("download command")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
