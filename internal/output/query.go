package output

import (
	"os"

	"github.com/synthient/cli/internal/utils"
	"github.com/synthient/go-synthient"
)

func AnonymizerQuery(query synthient.AnonymizersQuery) {
	values := []BlockValue{
		{Key: "Provider", Value: query.Provider},
		{Key: "Type", Value: query.Type},
		{Key: "Last Observed", Value: query.LastObserved},
		{Key: "Country Code", Value: query.CountryCode},
		{Key: "Format", Value: query.Format},
		{Key: "Full", Value: query.Full},
		{Key: "Order", Value: query.Order},
	}
	filteredValues := []BlockValue{}
	for _, value := range values {
		zero := utils.IsZeroAny(value.Value)
		if !zero {
			filteredValues = append(filteredValues, value)
		}
	}
	Block{
		Name:   "Query:",
		Values: filteredValues,
	}.Output(os.Stdout, StdoutStyles, 0)
}
