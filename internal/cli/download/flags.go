package download

import (
	"github.com/synthient/cli/internal/cli/stream"
	"github.com/synthient/go-synthient"
)

var (
	flags struct {
		query  synthient.AnonymizersQuery
		silent bool
	}
)

func init() {
	Command.PersistentFlags().
		BoolVarP(&flags.silent, "silent", "s", false, "Do not output when downloading")
	stream.AnonymizerQueryFlags(Command, &flags.query)
}
