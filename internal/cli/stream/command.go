package stream

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "stream",
	Short: "Stream a feed",
	Run: func(cmd *cobra.Command, args []string) {
		timber.Debug("stream command")
	},
}
