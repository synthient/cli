package account

import (
	"encoding/csv"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "account",
	Short: "Show account, scope, and quota details",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !slices.Contains(formats, flags.format) {
			timber.FatalMsg("invalid output format", timber.A("value", flags.format), timber.A("valid", strings.Join(formats, "|")))
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()

		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			app.Fatal(err, "failed to create synthient client")
		}
		config.ApplyToClient(&client)

		account, err := client.GetAccount(nil)
		if err != nil {
			app.Fatal(err, "failed to get account info")
		}

		switch flags.format {
		case "text":
			styles := output.NewStyles(out)
			output.Header(out, styles, "Account", account.Email)
			output.Divider(out, styles)
			blocks := []output.Block{
				{
					Name: "Identity",
					Values: []output.BlockValue{
						{Key: "First Name", Value: account.FirstName},
						{Key: "Last Name", Value: account.LastName},
						{Key: "Email", Value: account.Email},
					},
				},
				{
					Name: "Organization",
					Values: []output.BlockValue{
						{Key: "ID", Value: account.Organization.ID},
						{Key: "Name", Value: account.Organization.Name},
						{Key: "Relation", Value: account.Organization.Relation},
					},
				},
				{
					Name: "Quota",
					Values: []output.BlockValue{
						{Key: "Credits", Value: account.LookupQuota.Credits},
						{Key: "Resets In", Value: account.LookupQuota.ResetsIn},
						{Key: "Scopes", Value: account.Scopes},
					},
				},
			}
			for i, block := range blocks {
				block.Output(out, styles, 0, i+1 == len(blocks))
				if i+1 != len(blocks) {
					output.Divider(out, styles)
				}
			}
		case "json":
			output.JSON(out, account)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{
				"first_name",
				"last_name",
				"email",
				"organization.id",
				"organization.name",
				"organization.relation",
				"scopes",
				"lookup_quota.credits",
				"lookup_quota.resets_in",
			})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			err = writer.Write([]string{
				account.FirstName,
				account.LastName,
				account.Email,
				account.Organization.ID,
				account.Organization.Name,
				account.Organization.Relation,
				strings.Join(account.Scopes, "|"),
				strconv.Itoa(account.LookupQuota.Credits),
				strconv.Itoa(account.LookupQuota.ResetsIn),
			})
			if err != nil {
				app.Fatal(err, "failed to write csv row")
			}
			writer.Flush()
		}
	},
}
