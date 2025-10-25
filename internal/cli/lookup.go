package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

func lookup(cmd *cobra.Command, args []string) {
	synthientClient, err := synthient.CreateClient()
	if err != nil {
		timber.Fatal(err, "failed to create client")
	}

	spacing := len(args) == 1
	for i, ip := range args {
		resp, err := synthientClient.LookupIP(ip)
		if err != nil {
			timber.Fatal(err, "failed to lookup given IP")
		}
		resp.Output(spacing)
		if !spacing && i != len(args)-1 {
			fmt.Println()
		}
	}
}

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup information about a given IP",
	Args:  cobra.MinimumNArgs(1),
	Run:   lookup,
}

func init() {
	rootCmd.AddCommand(lookupCmd)
}
