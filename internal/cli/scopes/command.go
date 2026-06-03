package scopes

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"go.mattglei.ch/timber"
)

type Row struct {
	Scope   string `json:"scope"`
	Grants  string `json:"grants"`
	Granted bool   `json:"granted"`
}

var Command = &cobra.Command{
	Use:   "scopes",
	Short: "Show Synthient scopes and whether the current key has them",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !slices.Contains(formats, flags.format) {
			timber.FatalMsg("invalid output format", timber.A("value", flags.format), timber.A("valid", strings.Join(formats, "|")))
		}

		granted := []string{}
		config, err := conf.Read()
		if err == nil {
			client, clientErr := auth.SynthientClient(config)
			if clientErr == nil {
				config.ApplyToClient(&client)
				account, accountErr := client.GetAccount(nil)
				if accountErr == nil {
					granted = account.Scopes
				}
			}
		}

		rows := []Row{}
		for _, scope := range access.Scopes {
			rows = append(rows, Row{
				Scope:   scope.Scope,
				Grants:  scope.Grants,
				Granted: access.Has(granted, scope.Scope),
			})
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()
		switch flags.format {
		case "text":
			writeText(out, rows)
		case "json":
			output.JSON(out, rows)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"scope", "grants", "granted"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			for _, row := range rows {
				err = writer.Write([]string{row.Scope, row.Grants, fmt.Sprint(row.Granted)})
				if err != nil {
					app.Fatal(err, "failed to write csv row")
				}
			}
			writer.Flush()
		}
	},
}

func writeText(out *os.File, rows []Row) {
	styles := output.NewStyles(out)
	output.Header(out, styles, "Scopes")
	output.Divider(out, styles)
	for i, row := range rows {
		marker := " "
		markerStyle := styles.Muted
		if row.Granted {
			marker = "✓"
			markerStyle = styles.Success
		} else {
			marker = "·"
		}
		connector := "├"
		if i+1 == len(rows) {
			connector = "└"
		}
		output.WriteLine(
			out,
			fmt.Sprintf(
				"%s  %s %s %s",
				styles.Frame.Render(connector),
				output.PadRight(styles.Key.Render(row.Scope), 24),
				markerStyle.Render(marker),
				styles.Muted.Render(row.Grants),
			),
		)
	}
}
