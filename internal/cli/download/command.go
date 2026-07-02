package download

import (
	"fmt"
	"os"

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
	Use:   "download [stream] [snapshot] <file>",
	Short: "Download a feed snapshot to a Parquet file",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) >= 1 && len(args) <= 3 {
			return nil
		}
		return fmt.Errorf("accepts <file>, <stream> <file>, or <stream> <snapshot> <file>")
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
			streamName = "anonymizers"
			filename   = args[0]
			snapshot   = snapshotFromFlags()
		)
		if len(args) == 2 {
			streamName = args[0]
			filename = args[1]
		} else if len(args) == 3 {
			streamName = args[0]
			snapshot = args[1]
			filename = args[2]
		}
		stream, ok := feed.Find(streamName)
		if !ok {
			timber.FatalMsgf("unknown stream %q; valid streams: %s", streamName, feed.Names())
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
