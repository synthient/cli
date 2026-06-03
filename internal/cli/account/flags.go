package account

import (
	"fmt"
	"strings"
)

var (
	formats = []string{"text", "json", "csv"}

	flags struct {
		format string
		output string
	}
)

func init() {
	Command.PersistentFlags().
		StringVarP(&flags.output, "output", "o", "-", "Where to write output: '-' for stdout, or a file path")
	Command.PersistentFlags().
		StringVarP(&flags.format, "format", "f", "text", fmt.Sprintf("Output format [%s]", strings.Join(formats, "|")))
}
