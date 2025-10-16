package cli

import (
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

func lookup(cmd *cobra.Command, args []string) {
	synthientClient, err := synthient.CreateClient()
	if err != nil {
		timber.Fatal(err, "failed to create client")
	}

	ip := args[0]
	resp, err := synthientClient.LookupIP(ip)
	if err != nil {
		timber.Fatal(err, "failed to lookup given IP")
	}
	timber.Debug(resp.Location.City)
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
