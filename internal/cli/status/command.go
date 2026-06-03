package status

import (
	"encoding/csv"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
)

type Result struct {
	ConfigPath       string   `json:"config_path"`
	Profile          string   `json:"profile"`
	BaseAPI          string   `json:"base_api"`
	BaseFeeds        string   `json:"base_feeds"`
	BaseGRPC         string   `json:"base_grpc"`
	AuthSource       string   `json:"auth_source"`
	Authenticated    bool     `json:"authenticated"`
	AccountEmail     string   `json:"account_email"`
	Organization     string   `json:"organization"`
	LookupCredits    int      `json:"lookup_credits"`
	QuotaResetsIn    int      `json:"quota_resets_in"`
	Scopes           []string `json:"scopes"`
	AccountReachable bool     `json:"account_reachable"`
	AccountError     string   `json:"account_error"`
}

var Command = &cobra.Command{
	Use:   "status",
	Short: "Show CLI configuration, auth source, account, and quota status",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if !slices.Contains(formats, flags.format) {
			timber.FatalMsg("invalid output format", timber.A("value", flags.format), timber.A("valid", strings.Join(formats, "|")))
		}

		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read configuration file")
		}
		defaultClient := synthient.NewClient("")
		result := Result{
			ConfigPath:   config.Path,
			Profile:      config.Profile,
			AuthSource:   auth.ReadCredentialsStatus().Source,
			BaseAPI:      defaultClient.BaseAPI.String(),
			BaseFeeds:    defaultClient.BaseFeeds.String(),
			BaseGRPC:     config.GRPCEndpoint(synthient.DefaultGRPCEndpoint),
			AccountError: "",
		}
		client, err := auth.SynthientClient(config)
		if err == nil {
			config.ApplyToClient(&client)
			result.BaseAPI = client.BaseAPI.String()
			result.BaseFeeds = client.BaseFeeds.String()
			account, accountErr := client.GetAccount(nil)
			if accountErr == nil {
				result.Authenticated = true
				result.AccountReachable = true
				result.AccountEmail = account.Email
				result.Organization = account.Organization.Name
				result.LookupCredits = account.LookupQuota.Credits
				result.QuotaResetsIn = account.LookupQuota.ResetsIn
				result.Scopes = account.Scopes
			} else {
				result.AccountError = app.Explain(accountErr)
				if result.AccountError == "" {
					result.AccountError = accountErr.Error()
				}
			}
		} else {
			result.AccountError = app.Explain(err)
		}
		if config.BaseApiURL != nil {
			result.BaseAPI = config.BaseApiURL.String()
		}
		if config.BaseFeedsURL != nil {
			result.BaseFeeds = config.BaseFeedsURL.String()
		}
		result.BaseGRPC = config.GRPCEndpoint(synthient.DefaultGRPCEndpoint)

		out, closeOut := output.Open(flags.output)
		defer closeOut()
		switch flags.format {
		case "text":
			writeText(out, result)
		case "json":
			output.JSON(out, result)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"config_path", "profile", "base_api", "base_feeds", "base_grpc", "auth_source", "authenticated", "account_email", "organization", "lookup_credits", "quota_resets_in", "scopes", "account_reachable", "account_error"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			err = writer.Write([]string{
				result.ConfigPath,
				result.Profile,
				result.BaseAPI,
				result.BaseFeeds,
				result.BaseGRPC,
				result.AuthSource,
				strconv.FormatBool(result.Authenticated),
				result.AccountEmail,
				result.Organization,
				strconv.Itoa(result.LookupCredits),
				strconv.Itoa(result.QuotaResetsIn),
				strings.Join(result.Scopes, "|"),
				strconv.FormatBool(result.AccountReachable),
				result.AccountError,
			})
			if err != nil {
				app.Fatal(err, "failed to write csv row")
			}
			writer.Flush()
		}
	},
}

func writeText(out *os.File, result Result) {
	styles := output.NewStyles(out)
	output.Header(out, styles, "Status")
	output.Divider(out, styles)
	blocks := []output.Block{
		{
			Name: "Configuration",
			Values: []output.BlockValue{
				{Key: "Config", Value: result.ConfigPath},
				{Key: "Profile", Value: result.Profile},
				{Key: "Base API", Value: result.BaseAPI},
				{Key: "Base Feeds", Value: result.BaseFeeds},
				{Key: "Base gRPC", Value: result.BaseGRPC},
			},
		},
		{
			Name: "Authentication",
			Values: []output.BlockValue{
				{Key: "Source", Value: result.AuthSource},
				{Key: "Authenticated", Value: result.Authenticated},
				{Key: "Account Reachable", Value: result.AccountReachable},
				{Key: "Error", Value: result.AccountError},
			},
		},
		{
			Name: "Account",
			Values: []output.BlockValue{
				{Key: "Email", Value: result.AccountEmail},
				{Key: "Organization", Value: result.Organization},
				{Key: "Credits", Value: result.LookupCredits},
				{Key: "Resets In", Value: result.QuotaResetsIn},
				{Key: "Scopes", Value: result.Scopes},
			},
		},
	}
	for i, block := range blocks {
		block.Output(out, styles, 0, i+1 == len(blocks))
		if i+1 != len(blocks) {
			output.Divider(out, styles)
		}
	}
}
