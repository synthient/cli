package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

func lookup(cmd *cobra.Command, args []string) {
	isTTY := term.IsTerminal(os.Stdin.Fd())
	if !isTTY {
		inputBinary, err := io.ReadAll(os.Stdin)
		if err != nil {
			timber.Fatal(err, "failed to read standard input")
		}
		args = strings.Fields(string(inputBinary))
	}

	if len(args) == 0 {
		timber.Warning("Given zero IP addresses to lookup")
		return
	}

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
	Run:   lookup,
}

func init() {
	rootCmd.AddCommand(lookupCmd)
}
