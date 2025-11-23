package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli/process"
	"go.mattglei.ch/timber"
)

var processCmd = &cobra.Command{
	Use:    "process",
	Short:  "Process and inject synthient data into a CSV",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		file, err := os.Open(filename)
		if err != nil {
			timber.Fatal(err, "failed to read", filename)
		}
		defer func() {
			err = file.Close()
			if err != nil {
				timber.Fatal(err, "failed to close", filename)
			}
		}()

		// cr := csv.NewReader(file)
		// ipColumn := process.GetIpColumn(cr)
		process.SelectColumnsToAdd()
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
}
