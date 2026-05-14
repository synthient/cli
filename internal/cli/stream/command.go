package stream

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "stream",
	Short: "Stream anonymizer feed to stdout",
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
		for event, err := range client.StreamAnonymizer(nil) {
			if err != nil {
				timber.Fatal(err, "failed to stream anonymizer feed")
			}
			err = enc.Encode(event)
			if err != nil {
				timber.Fatal(err, "failed to encode anonymizer event")
			}
		}
	},
}
