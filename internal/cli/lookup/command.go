package lookup

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/utils"
	"github.com/synthient/go-synthient"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "lookup",
	Short: "Lookup information about a given IP",
	Run: func(cmd *cobra.Command, args []string) {
		if !slices.Contains(formats, flags.format) {
			timber.ErrorMsg(
				"invalid output flag value of",
				flags.format,
				"must be either",
				strings.Join(formats, "|"),
			)
		}

		out := os.Stdout
		if flags.output != "-" {
			file, err := os.Create(flags.output)
			if err != nil {
				timber.Fatal(err, "failed to create output file:", flags.output)
			}
			out = file
			defer func() { _ = file.Close() }()
		}

		pipedIn := utils.ReadPipedInput()
		if pipedIn != "" {
			args = strings.Fields(pipedIn)
		}

		if len(args) == 0 {
			timber.Warning("Given zero IP addresses to lookup")
			return
		}

		config, err := conf.Read()
		if err != nil {
			timber.Fatal(err, "failed to read from config file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			timber.Fatal(err, "failed to create synthient client")
		}

		styles := output.NewStyles(out)
		spacing := len(args) == 1
		ips := []synthient.IP{}
		for i, ip := range args {
			resp, err := client.GetIP(ip, nil)
			if err != nil {
				timber.Fatal(err, "failed to lookup given IP")
			}

			switch flags.format {
			case "text":
				output.IP(resp, out, styles, spacing)
				if !spacing && i != len(args)-1 {
					fmt.Println()
				}
			case "json", "csv":
				ips = append(ips, resp)
			}
		}

		switch flags.format {
		case "json":
			b, err := json.Marshal(ips)
			if err != nil {
				timber.Fatal(err, "failed to marshal IPs into json data")
			}
			output.WriteLine(out, string(b))
		case "csv":
			writer := csv.NewWriter(out)
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
				output.IpCSV(ip, writer)
			}
		}
	},
}
