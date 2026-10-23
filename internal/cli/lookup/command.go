package lookup

import (
	"context"
	"encoding/csv"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/utils"
	"github.com/synthient/synthient-go"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "lookup [ip ...]",
	Short: "Lookup information about a given IP",
	Long:  "Lookup IP intelligence. Multiple IPs use the documented batch endpoint, and stdin can provide additional whitespace-separated IPs.",
	Run: func(cmd *cobra.Command, args []string) {
		runIPs(args)
	},
}

var IPCommand = &cobra.Command{
	Use:   "ip [ip ...]",
	Short: "Lookup IP intelligence",
	Run: func(cmd *cobra.Command, args []string) {
		runIPs(args)
	},
}

func init() {
	Command.AddCommand(IPCommand)
}

func runIPs(args []string) {
	validateFormat()

	out, closeOut := output.Open(flags.output)
	defer closeOut()

	pipedIn := utils.ReadPipedInput()
	if pipedIn != "" {
		args = append(args, strings.Fields(pipedIn)...)
	}

	if len(args) == 0 {
		output.Warn(out, output.NewStyles(out), "Given zero IP addresses to lookup")
		return
	}

	addrs, err := utils.ParseIPs(args)
	if err != nil {
		timber.FatalMsg(err.Error())
	}

	config, err := conf.Read()
	if err != nil {
		app.Fatal(err, "failed to read from config file")
	}

	client, err := auth.SynthientClient(config)
	if err != nil {
		app.Fatal(err, "failed to create synthient client")
	}

	ctx := context.Background()
	styles := output.NewStyles(out)
	ips := []synthient.IPInfo{}
	if len(addrs) == 1 {
		resp, err := client.LookupIP(ctx, addrs[0])
		if err != nil {
			app.Fatal(err, "failed to lookup given IP")
		}
		ips = append(ips, *resp)
	} else {
		resp, err := client.LookupIPs(ctx, addrs)
		if err != nil {
			app.Fatal(err, "failed to lookup given IPs")
		}
		ips = append(ips, resp...)
	}

	switch flags.format {
	case "text":
		for i, ip := range ips {
			output.IP(ip, out, styles, len(ips) == 1)
			if i+1 != len(ips) {
				output.WriteLine(out)
			}
		}
	case "json":
		output.JSON(out, ips)
	case "csv":
		writer := csv.NewWriter(out)
		err := writer.Write([]string{
			"ip",
			"network.asn",
			"network.isp",
			"network.type",
			"network.org",
			"network.domain",
			"network.abuse_email",
			"network.abuse_phone",
			"location.city",
			"location.state",
			"location.country",
			"location.timezone",
			"location.longitude",
			"location.latitude",
			"location.geohash",
			"intelligence.behavior",
			"intelligence.categories",
			"intelligence.riskscore",
			"intelligence.devices",
			"intelligence.providers",
		})
		if err != nil {
			app.Fatal(err, "failed to write csv header")
		}
		for _, ip := range ips {
			output.IpCSV(ip, writer)
		}
	}
}

func validateFormat() {
	if !slices.Contains(formats, flags.format) {
		timber.FatalMsgf("invalid output format %q; valid formats: %s", flags.format, strings.Join(formats, "|"))
	}
}
