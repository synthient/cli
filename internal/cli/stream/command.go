package stream

import (
	"io"
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

		s, err := client.StreamAnonymizersFeed(flags.query, nil)
		if err != nil {
			timber.Fatal(err, "failed to stream anonymizer feed")
		}
		defer func() { _ = s.Close() }()

		_, err = io.Copy(os.Stdout, s)
		if err != nil {
			timber.Fatal(err, "failed to copy stream to stdout")
		}
	},
}
