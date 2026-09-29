package download

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/cli/internal/feeddownload"
	"github.com/synthient/cli/internal/options"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "download <stream> [snapshot] <file>",
	Short: "Download a feed snapshot to a Parquet file",
	Long: fmt.Sprintf(`Download a feed snapshot to a Parquet file.

The stream always comes first and the output file always comes last. The
snapshot is optional and defaults to the latest one (or --date/--hour).

Streams: %s
Snapshots: latest, YYYY-MM-DD, or YYYY-MM-DD/HH`, streamList()),
	Example: `  synthient download proxies proxies.parquet
  synthient download proxies 2026-09-27 proxies.parquet
  synthient download proxies 2026-09-27/13 proxies.parquet
  synthient download anonymizers anonymizers.parquet --date 2026-09-27`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 || len(args) == 3 {
			return nil
		}
		if len(args) == 1 {
			return fmt.Errorf("missing stream; expected: synthient download <stream> %s", args[0])
		}
		return fmt.Errorf("expected 2 or 3 arguments (<stream> [snapshot] <file>), got %d", len(args))
	},
	Run: func(cmd *cobra.Command, args []string) {
		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			app.Fatal(err, "failed to create synthient client")
		}
		config.ApplyToClient(&client)

		var (
			streamName = args[0]
			filename   = args[len(args)-1]
			snapshot   = snapshotFromFlags()
		)
		if len(args) == 3 {
			snapshot = args[1]
		}
		stream, ok := feed.Find(streamName)
		if !ok {
			_, lastIsStream := feed.Find(filename)
			if lastIsStream {
				reordered := slices.Clone(args)
				slices.Reverse(reordered)
				timber.FatalMsgf(
					"arguments are out of order; the stream comes first and the file last: synthient download %s",
					strings.Join(reordered, " "),
				)
			}
			timber.FatalMsgf("unknown stream %q; valid streams: %s", streamName, streamList())
		}

		if !flags.noPreflight {
			err = access.Require(client, access.Required(stream, "feed"))
			if err != nil {
				app.Fatal(err, "failed feed access preflight")
			}
		}

		_, err = feeddownload.Run(feeddownload.Options{
			Client:   client,
			Stream:   stream,
			Snapshot: snapshot,
			Filename: filename,
			Force:    flags.force,
			Verify:   flags.verify,
			Quiet:    options.Quiet || flags.silent,
			Out:      os.Stdout,
		})
		if err != nil {
			app.Fatal(err, fmt.Sprintf("failed to download %s from stream %s", filename, stream.Name))
		}
	},
}

func streamList() string {
	return strings.ReplaceAll(feed.Names(), "|", ", ")
}

func snapshotFromFlags() string {
	if flags.hour < 0 {
		return flags.date
	}
	if flags.hour > 23 {
		timber.FatalMsgf("hour must be between 0 and 23: %d", flags.hour)
	}
	if flags.date == "latest" {
		timber.FatalMsg("hour cannot be used with latest snapshot")
	}
	return fmt.Sprintf("%s/%d", flags.date, flags.hour)
}
