package lookup

import (
	"encoding/csv"
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"go.mattglei.ch/timber"
)

var DomainCommand = &cobra.Command{
	Use:   "domain <domain>",
	Short: "Lookup domain intelligence from Helios observations",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateFormat()

		out, closeOut := output.Open(flags.output)
		defer closeOut()

		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read from config file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			app.Fatal(err, "failed to create synthient client")
		}
		config.ApplyToClient(&client)

		resp, err := client.GetDomain(args[0], nil)
		if err != nil {
			app.Fatal(err, "failed to lookup given domain")
		}

		switch flags.format {
		case "text":
			output.Domain(resp, out, output.NewStyles(out))
		case "json":
			output.JSON(out, resp)
		case "csv":
			topSubdomains, err := json.Marshal(resp.TopSubdomains)
			if err != nil {
				app.Fatal(err, "failed to marshal top subdomains")
			}
			topPorts, err := json.Marshal(resp.TopPorts)
			if err != nil {
				app.Fatal(err, "failed to marshal top ports")
			}
			recentEvents, err := json.Marshal(resp.RecentEvents)
			if err != nil {
				app.Fatal(err, "failed to marshal recent events")
			}
			writer := csv.NewWriter(out)
			err = writer.Write([]string{
				"domain",
				"status",
				"events_24h",
				"total_events_30d",
				"unique_ips_24h",
				"unique_ips_30d",
				"top_subdomains",
				"top_ports",
				"recent_events",
			})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			err = writer.Write([]string{
				resp.Domain,
				resp.Status,
				strconv.Itoa(resp.Stats.Events24H),
				strconv.Itoa(resp.Stats.TotalEvents30D),
				strconv.Itoa(resp.UniqueIPs.Value24H),
				strconv.Itoa(resp.UniqueIPs.Value30D),
				string(topSubdomains),
				string(topPorts),
				string(recentEvents),
			})
			if err != nil {
				app.Fatal(err, "failed to write csv row")
			}
			writer.Flush()
		default:
			timber.FatalMsg("invalid output format")
		}
	},
}

func init() {
	Command.AddCommand(DomainCommand)
}
