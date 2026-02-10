package stream

import (
	"github.com/spf13/cobra"
	"github.com/synthient/go-synthient"
)

var (
	flags struct {
		query synthient.AnonymizersQuery
	}
)

func init() {
	AnonymizerQueryFlags(Command, &flags.query)
}

func AnonymizerQueryFlags(cmd *cobra.Command, query *synthient.AnonymizersQuery) {
	persistentFlags := cmd.PersistentFlags()
	persistentFlags.
		StringVar(
			&query.Provider,
			"provider",
			"",
			"The name of the provider to filter by. If not provided, returns all providers.",
		)
	persistentFlags.
		StringVar(
			&query.Type,
			"type",
			"",
			"The type of anonymization service to filter by. If not provided, returns all types. Please note it's redundant to provide both provider and type in the same command.",
		)
	persistentFlags.
		StringVar(
			&query.LastObserved,
			"last_observed",
			"",
			"Filter by how recently the IP was observed. For example 24H, 7D, or 1M. If not provided, returns all records regardless of last observed date.",
		)
	persistentFlags.
		StringVarP(
			&query.Format,
			"format",
			"f",
			"CSV",
			"The format of the returned data. Options are JSONL, CSV, or TEXT.",
		)
	persistentFlags.
		StringVar(
			&query.CountryCode,
			"country_code",
			"",
			"Filter results to only include IPs from the specified country code, like 'US' or 'DE'. If not provided, returns all countries.",
		)
	persistentFlags.
		BoolVar(
			&query.Full,
			"full",
			false,
			"If true, returns all records. If false, results are distinct on ip_address and provider.",
		)
	persistentFlags.
		StringVar(
			&query.Order,
			"order",
			"desc",
			"The sort order for results. Options are asc or desc.",
		)
}
