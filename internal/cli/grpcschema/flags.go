package grpcschema

import (
	"fmt"
	"strings"
	"time"
)

var (
	formats = []string{"text", "json", "binpb"}

	flags struct {
		format            string
		output            string
		endpoint          string
		timeout           time.Duration
		plaintext         bool
		includeReflection bool
	}
)

func init() {
	SchemaCommand.Flags().
		StringVarP(&flags.output, "output", "o", "-", "Where to write output: '-' for stdout, or a file path")
	SchemaCommand.Flags().
		StringVarP(&flags.format, "format", "f", "text", fmt.Sprintf("Output format [%s]", strings.Join(formats, "|")))
	SchemaCommand.Flags().
		StringVar(&flags.endpoint, "endpoint", "", "gRPC endpoint; defaults to config endpoints.base_grpc")
	SchemaCommand.Flags().
		DurationVar(&flags.timeout, "timeout", 15*time.Second, "Reflection request timeout")
	SchemaCommand.Flags().
		BoolVar(&flags.plaintext, "plaintext", false, "Connect without TLS")
	SchemaCommand.Flags().
		BoolVar(&flags.includeReflection, "include-reflection", false, "Include gRPC reflection service descriptors")
}
