package feeds

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/cli/internal/feeddownload"
	"github.com/synthient/cli/internal/options"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "feeds",
	Short: "List feed streams, snapshots, and metadata",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			timber.Fatal(err, "output help")
		}
	},
}

var StreamsCommand = &cobra.Command{
	Use:   "streams",
	Short: "List supported feed streams",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		validateFormat()
		out, closeOut := output.Open(flags.output)
		defer closeOut()

		switch flags.format {
		case "text":
			styles := output.NewStyles(out)
			output.Header(out, styles, "Feed Streams")
			output.Divider(out, styles)
			for i, stream := range feed.Streams {
				name := output.PadRight(styles.Warm.Render(stream.Name), 15)
				output.WriteLine(
					out,
					fmt.Sprintf(
						"%s  %s %s",
						styles.Frame.Render(connector(i+1 == len(feed.Streams))),
						name,
						styles.Muted.Render(stream.Description),
					),
				)
			}
		case "json":
			output.JSON(out, feed.Streams)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"name", "aliases", "description"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			for _, stream := range feed.Streams {
				err = writer.Write([]string{stream.Name, strings.Join(stream.Aliases, "|"), stream.Description})
				if err != nil {
					app.Fatal(err, "failed to write csv row")
				}
			}
			writer.Flush()
		}
	},
}

var SnapshotsCommand = &cobra.Command{
	Use:     "snapshots <stream>",
	Aliases: []string{"list"},
	Short:   "List available daily and hourly snapshots",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		validateFormat()
		stream, ok := feed.Find(args[0])
		if !ok {
			timber.FatalMsgf("unknown stream %q; valid streams: %s", args[0], feed.Names())
		}
		client := newClient()
		requireFeedScope(client, stream)
		opts := &synthient.FeedSnapshotsOptions{}
		if flags.limit > 0 {
			opts.Limit = flags.limit
		}
		if flags.cursor != "" {
			opts.Cursor = flags.cursor
		}
		page, err := client.FeedSnapshots(stream.Name, opts, nil)
		if err != nil {
			app.Fatal(err, "failed to list feed snapshots")
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()
		switch flags.format {
		case "text":
			writeSnapshotList(out, output.NewStyles(out), stream, page)
		case "json":
			output.JSON(out, page)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"stream", "kind", "id", "date", "hour", "size_bytes", "row_count", "checksum", "created_at", "download_path"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			for _, snap := range page.Feeds {
				hour := ""
				if snap.Hour != nil {
					hour = strconv.Itoa(*snap.Hour)
				}
				err = writer.Write([]string{
					page.Stream,
					snap.Kind,
					snap.ID,
					snap.Date,
					hour,
					strconv.FormatInt(snap.SizeBytes, 10),
					strconv.FormatInt(snap.RowCount, 10),
					snap.Checksum,
					strconv.FormatInt(snap.CreatedAt, 10),
					snap.DownloadPath,
				})
				if err != nil {
					app.Fatal(err, "failed to write csv row")
				}
			}
			writer.Flush()
		}
	},
}

var MetaCommand = &cobra.Command{
	Use:   "meta <stream> <snapshot>",
	Short: "Show snapshot metadata and parquet schema",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		validateFormat()
		stream, ok := feed.Find(args[0])
		if !ok {
			timber.FatalMsgf("unknown stream %q; valid streams: %s", args[0], feed.Names())
		}
		client := newClient()
		requireFeedScope(client, stream)
		meta, err := client.FeedSnapshotMeta(stream.Name, args[1], nil)
		if err != nil {
			app.Fatal(err, "failed to get feed snapshot metadata")
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()
		switch flags.format {
		case "text":
			writeSnapshotMeta(out, output.NewStyles(out), meta)
		case "json":
			output.JSON(out, meta)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"stream", "kind", "id", "format", "date", "hour", "created_at", "size", "rows", "checksum", "field.name", "field.type"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			hour := ""
			if meta.Hour != nil {
				hour = strconv.Itoa(*meta.Hour)
			}
			for _, field := range meta.Schema.Fields {
				err = writer.Write([]string{
					meta.Stream,
					meta.Kind,
					meta.ID,
					meta.Format,
					strconv.FormatInt(meta.Date, 10),
					hour,
					strconv.FormatInt(meta.CreatedAt, 10),
					strconv.FormatInt(meta.Size, 10),
					strconv.FormatInt(meta.Rows, 10),
					meta.Checksum,
					field.Name,
					field.Type,
				})
				if err != nil {
					app.Fatal(err, "failed to write csv row")
				}
			}
			writer.Flush()
		}
	},
}

var SchemaCommand = &cobra.Command{
	Use:   "schema <stream> <snapshot>",
	Short: "Show the parquet schema for a snapshot",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		validateFormat()
		stream, ok := feed.Find(args[0])
		if !ok {
			timber.FatalMsgf("unknown stream %q; valid streams: %s", args[0], feed.Names())
		}
		client := newClient()
		requireFeedScope(client, stream)
		meta, err := client.FeedSnapshotMeta(stream.Name, args[1], nil)
		if err != nil {
			app.Fatal(err, "failed to get feed snapshot schema")
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()
		switch flags.format {
		case "text":
			writeSchema(out, output.NewStyles(out), meta)
		case "json":
			output.JSON(out, meta.Schema.Fields)
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"name", "type"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			for _, field := range meta.Schema.Fields {
				err = writer.Write([]string{field.Name, field.Type})
				if err != nil {
					app.Fatal(err, "failed to write csv row")
				}
			}
			writer.Flush()
		}
	},
}

var ChecksumCommand = &cobra.Command{
	Use:   "checksum <stream> <snapshot>",
	Short: "Show the expected SHA-256 checksum for a snapshot",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		validateFormat()
		stream, ok := feed.Find(args[0])
		if !ok {
			timber.FatalMsgf("unknown stream %q; valid streams: %s", args[0], feed.Names())
		}
		client := newClient()
		requireFeedScope(client, stream)
		meta, err := client.FeedSnapshotMeta(stream.Name, args[1], nil)
		if err != nil {
			app.Fatal(err, "failed to get feed snapshot checksum")
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()
		switch flags.format {
		case "text":
			output.Header(out, output.NewStyles(out), "Checksum", meta.ID)
			output.Outro(out, output.NewStyles(out), meta.Checksum)
		case "json":
			output.JSON(out, struct {
				Stream   string `json:"stream"`
				Snapshot string `json:"snapshot"`
				Checksum string `json:"checksum"`
			}{Stream: meta.Stream, Snapshot: meta.ID, Checksum: meta.Checksum})
		case "csv":
			writer := csv.NewWriter(out)
			err := writer.Write([]string{"stream", "snapshot", "checksum"})
			if err != nil {
				app.Fatal(err, "failed to write csv header")
			}
			err = writer.Write([]string{meta.Stream, meta.ID, meta.Checksum})
			if err != nil {
				app.Fatal(err, "failed to write csv row")
			}
			writer.Flush()
		}
	},
}

var DownloadCommand = &cobra.Command{
	Use:   "download <stream> <snapshot> <file>",
	Short: "Download a feed snapshot to a Parquet file",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		stream, ok := feed.Find(args[0])
		if !ok {
			timber.FatalMsgf("unknown stream %q; valid streams: %s", args[0], feed.Names())
		}
		client := newClient()
		requireFeedScope(client, stream)

		_, err := feeddownload.Run(feeddownload.Options{
			Client:   client,
			Stream:   stream,
			Snapshot: args[1],
			Filename: args[2],
			Force:    flags.force,
			Verify:   flags.verify,
			Quiet:    options.Quiet,
			Out:      os.Stdout,
		})
		if err != nil {
			app.Fatal(err, "failed to download feed snapshot")
		}
	},
}

func init() {
	Command.AddCommand(StreamsCommand, SnapshotsCommand, MetaCommand, SchemaCommand, ChecksumCommand, DownloadCommand)
}

func validateFormat() {
	if !slices.Contains(formats, flags.format) {
		timber.FatalMsgf("invalid output format %q; valid formats: %s", flags.format, strings.Join(formats, "|"))
	}
}

func newClient() synthient.Client {
	config, err := conf.Read()
	if err != nil {
		app.Fatal(err, "failed to read configuration file")
	}
	client, err := auth.SynthientClient(config)
	if err != nil {
		app.Fatal(err, "failed to create synthient client")
	}
	config.ApplyToClient(&client)
	return client
}

func requireFeedScope(client synthient.Client, stream feed.Stream) {
	if flags.noPreflight {
		return
	}
	err := access.Require(client, access.Required(stream, "feed"))
	if err != nil {
		app.Fatal(err, "failed feed access preflight")
	}
}

func connector(final bool) string {
	if final {
		return "└"
	}
	return "├"
}

func writeSnapshotList(out *os.File, styles output.Styles, stream feed.Stream, page synthient.FeedSnapshotsPage) {
	output.Header(out, styles, "Snapshots", stream.Name)
	output.Divider(out, styles)
	output.WriteLine(
		out,
		fmt.Sprintf(
			"%s  %s %s %s %s %s %s",
			styles.Frame.Render("│"),
			output.PadRight(styles.Muted.Render("ID"), 18),
			output.PadRight(styles.Muted.Render("Kind"), 8),
			output.PadRight(styles.Muted.Render("Date"), 10),
			output.PadRight(styles.Muted.Render("Hour"), 4),
			output.PadLeft(styles.Muted.Render("Rows"), 12),
			output.PadLeft(styles.Muted.Render("Size"), 12),
		),
	)
	if len(page.Feeds) == 0 {
		output.Outro(out, styles, "No snapshots returned")
		return
	}
	for i, snap := range page.Feeds {
		hour := ""
		if snap.Hour != nil {
			hour = strconv.Itoa(*snap.Hour)
		}
		output.WriteLine(
			out,
			fmt.Sprintf(
				"%s  %s %s %s %s %s %s",
				styles.Frame.Render(connector(i+1 == len(page.Feeds) && page.NextCursor == "")),
				output.PadRight(styles.Key.Render(snap.ID), 18),
				output.PadRight(styles.Warm.Render(snap.Kind), 8),
				output.PadRight(styles.Value.Render(snap.Date), 10),
				output.PadRight(styles.Muted.Render(hour), 4),
				output.PadLeft(styles.Num.Render(humanize.Comma(snap.RowCount)), 12),
				output.PadLeft(styles.Secondary.Render(humanize.Bytes(uint64(snap.SizeBytes))), 12),
			),
		)
	}
	if page.NextCursor != "" {
		output.Divider(out, styles)
		output.Outro(out, styles, fmt.Sprintf("Next cursor: %s", page.NextCursor))
	}
}

func writeSnapshotMeta(out *os.File, styles output.Styles, meta synthient.FeedSnapshotMeta) {
	output.Header(out, styles, "Snapshot", meta.ID)
	output.Divider(out, styles)
	hour := "None"
	if meta.Hour != nil {
		hour = strconv.Itoa(*meta.Hour)
	}
	blocks := []output.Block{
		{
			Name: "Metadata",
			Values: []output.BlockValue{
				{Key: "Stream", Value: meta.Stream},
				{Key: "Kind", Value: meta.Kind},
				{Key: "Format", Value: meta.Format},
				{Key: "Date", Value: time.Unix(meta.Date, 0).UTC().Format(time.RFC3339)},
				{Key: "Hour", Value: hour},
				{Key: "Created", Value: time.Unix(meta.CreatedAt, 0).UTC().Format(time.RFC3339)},
				{Key: "Rows", Value: humanize.Comma(meta.Rows)},
				{Key: "Size", Value: humanize.Bytes(uint64(meta.Size))},
				{Key: "Checksum", Value: meta.Checksum},
			},
		},
	}
	schema := output.Block{Name: "Schema"}
	for _, field := range meta.Schema.Fields {
		schema.Values = append(schema.Values, output.BlockValue{Key: field.Name, Value: field.Type})
	}
	blocks = append(blocks, schema)
	for i, block := range blocks {
		block.Output(out, styles, 0, i+1 == len(blocks))
		if i+1 != len(blocks) {
			output.Divider(out, styles)
		}
	}
}

func writeSchema(out *os.File, styles output.Styles, meta synthient.FeedSnapshotMeta) {
	output.Header(out, styles, "Schema", meta.ID)
	output.Divider(out, styles)
	block := output.Block{Name: "Fields"}
	for _, field := range meta.Schema.Fields {
		block.Values = append(block.Values, output.BlockValue{Key: field.Name, Value: field.Type})
	}
	block.Output(out, styles, 0, true)
}
