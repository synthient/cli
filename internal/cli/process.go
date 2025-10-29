package cli

import (
	"github.com/spf13/cobra"
	"go.mattglei.ch/timber"
)

func process(cmd *cobra.Command, args []string) {
	timber.Debug("process command")
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process and inject synthient data into a CSV",
	Run:   process,
}

func init() {
	rootCmd.AddCommand(processCmd)
}
