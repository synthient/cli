package stream

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "stream [anonymizer|torrent|proxy]",
	Short: "Stream a feed to stdout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := conf.Read()
		if err != nil {
			timber.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			timber.Fatal(err, "failed to create synthient client")
		}

		enc := json.NewEncoder(os.Stdout)

		switch args[0] {
		case "anonymizer":
			for event, err := range client.StreamAnonymizer(nil) {
				if err != nil {
					timber.Fatal(err, "failed to stream anonymizer feed")
				}
				err = enc.Encode(event)
				if err != nil {
					timber.Fatal(err, "failed to encode anonymizer event")
				}
			}
		case "torrent":
			for event, err := range client.StreamTorrent(nil) {
				if err != nil {
					timber.Fatal(err, "failed to stream torrent feed")
				}
				err = enc.Encode(event)
				if err != nil {
					timber.Fatal(err, "failed to encode torrent event")
				}
			}
		case "proxy":
			for event, err := range client.StreamProxy(nil) {
				if err != nil {
					timber.Fatal(err, "failed to stream proxy feed")
				}
				err = enc.Encode(event)
				if err != nil {
					timber.Fatal(err, "failed to encode proxy event")
				}
			}
		default:
			timber.Fatal(fmt.Errorf("unknown feed %q", args[0]), "valid feeds: anonymizer, torrent, proxy")
		}
	},
}
