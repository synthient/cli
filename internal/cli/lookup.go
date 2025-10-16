package cli

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

func lookup(cmd *cobra.Command, args []string) {
	ip := args[0]
	timber.Debug("ip:", ip)
}

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup information about a given IP",
	Args:  cobra.ExactArgs(1),
	Run:   lookup,
}

func init() {
	rootCmd.AddCommand(lookupCmd)
}
