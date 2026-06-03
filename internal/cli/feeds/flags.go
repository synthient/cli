package feeds

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	formats = []string{"text", "json", "csv"}

	flags struct {
		format      string
		output      string
		limit       int
		cursor      string
		force       bool
		verify      bool
		noPreflight bool
	}
)

func init() {
	for _, command := range []*cobra.Command{StreamsCommand, SnapshotsCommand, MetaCommand, SchemaCommand, ChecksumCommand} {
		command.Flags().
			StringVarP(&flags.output, "output", "o", "-", "Where to write output: '-' for stdout, or a file path")
		command.Flags().
			StringVarP(&flags.format, "format", "f", "text", fmt.Sprintf("Output format [%s]", strings.Join(formats, "|")))
	}
	SnapshotsCommand.Flags().
		IntVarP(&flags.limit, "limit", "l", 100, "Snapshot page size")
	SnapshotsCommand.Flags().
		StringVar(&flags.cursor, "cursor", "", "Snapshot pagination cursor")
	DownloadCommand.Flags().
		BoolVar(&flags.force, "force", false, "Overwrite an existing file")
	DownloadCommand.Flags().
		BoolVar(&flags.verify, "verify", false, "Verify downloaded file against snapshot metadata checksum")
	for _, command := range []*cobra.Command{SnapshotsCommand, MetaCommand, SchemaCommand, ChecksumCommand, DownloadCommand} {
		command.Flags().
			BoolVar(&flags.noPreflight, "no-preflight", false, "Skip account scope preflight checks")
	}
}
