package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

var (
	lookupFormatFlag       string
	lookupOutputFlag       string
	lookupFormatFlagValues = []string{"text", "json", "csv"}
	lookupOutput           = os.Stdout
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
	if lookupOutputFlag != "-" {
		file, err := os.Create(lookupOutputFlag)
		if err != nil {
			timber.Fatal(err, "failed to create output file:", lookupOutputFlag)
		}
		lookupOutput = file
		defer file.Close() // nolint:errcheck
	}

	pipedIn := PipedInput()
	if pipedIn != "" {
		args = strings.Fields(pipedIn)
	}

	if len(args) == 0 {
		timber.Warning("Given zero IP addresses to lookup")
		return
	}

	config := conf.Read()
	synthientClient, err := synthient.CreateClient(config)
	if err != nil {
		timber.Fatal(err, "failed to create client")
	}

	styles := output.NewStyles(lookupOutput)

	spacing := len(args) == 1
	ips := []synthient.LookupResponse{}
	for i, ip := range args {
		resp, err := synthientClient.LookupIP(ip)
		if err != nil {
			timber.Fatal(err, "failed to lookup given IP")
		}

		switch lookupFormatFlag {
		case "text":
			resp.Output(lookupOutput, styles, spacing)
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
		output.WriteLine(lookupOutput, string(b))
	case "csv":
		writer := csv.NewWriter(lookupOutput)
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
		StringVarP(&lookupOutputFlag, "output", "o", "-", "Where to write output: '-' for stdout, or a file path (e.g. 'lookup.json' or 'lookup.csv)")
	lookupCmd.PersistentFlags().
		StringVarP(&lookupFormatFlag, "format", "f", "text", fmt.Sprintf("Output format [%s]", strings.Join(lookupFormatFlagValues, "|")))
	rootCmd.AddCommand(lookupCmd)
}
