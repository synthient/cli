package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

var (
	lookupFormatFlag       string
	lookupFormatFlagValues = []string{"text", "json", "csv"}
)

func lookup(cmd *cobra.Command, args []string) {
	if !slices.Contains(lookupFormatFlagValues, lookupFormatFlag) {
		timber.ErrorMsg(
			"invalid output flag value of",
			lookupFormatFlag,
			"must be either",
			strings.Join(lookupFormatFlagValues, "|"),
		)
	}

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
	ips := []synthient.LookupResponse{}
	for i, ip := range args {
		resp, err := synthientClient.LookupIP(ip)
		if err != nil {
			timber.Fatal(err, "failed to lookup given IP")
		}

		switch lookupFormatFlag {
		case "text":
			resp.Output(spacing)
			if !spacing && i != len(args)-1 {
				fmt.Println()
			}
		case "json", "csv":
			ips = append(ips, resp)
		}
	}

	switch lookupFormatFlag {
	case "json":
		b, err := json.Marshal(ips)
		if err != nil {
			timber.Fatal(err, "failed to marshal IPs into json data")
		}
		fmt.Println(string(b))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		err := writer.Write([]string{
			"ip", "network.asn", "network.isp", "network.type", "location.city",
			"location.state", "location.country", "location.timezone", "location.longitude",
			"location.latitude", "location.geohash", "ipdata.devicecount", "ipdata.behavior",
			"ipdata.categories", "ipdata.iprisk", "ipdata.enriched",
		})
		if err != nil {
			timber.Fatal(err, "failed to write csv header")
		}
		for _, ip := range ips {
			ip.OutputCSV(writer)
		}
	}
}

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup information about a given IP",
	Run:   lookup,
}

func init() {
	lookupCmd.PersistentFlags().
		StringVarP(&lookupFormatFlag, "format", "f", "text", fmt.Sprintf("Output format [%s]", strings.Join(lookupFormatFlagValues, "|")))
	rootCmd.AddCommand(lookupCmd)
}
