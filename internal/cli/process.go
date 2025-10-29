package cli

import (
	"encoding/csv"
	"os"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli/process"
	"go.mattglei.ch/timber"
)

func runProcess(cmd *cobra.Command, args []string) {
	filename := args[0]
	file, err := os.Open(filename)
	if err != nil {
		timber.Fatal(err, "failed to read", filename)
	}

	cr := csv.NewReader(file)
	ipColumn := process.GetIpColumn(cr)
	timber.Debug(ipColumn)

	err = file.Close()
	if err != nil {
		timber.Fatal(err, "failed to close", filename)
	}
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process and inject synthient data into a CSV",
	Run:   runProcess,
}

func init() {
	rootCmd.AddCommand(processCmd)
}
