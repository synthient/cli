package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

func auth(cmd *cobra.Command, args []string) {
	fmt.Print("Please provide Synthient API key: ")
	var key string
	_, err := fmt.Scanln(&key)
	if err != nil {
		timber.Fatal(err, "reading user input failed")
	}

	key = strings.TrimSpace(strings.Trim(key, "\n"))
	if key == "" {
		timber.FatalMsg("Please provide valid key.")
	}

	err = synthient.StoreApiKey(key)
	if err != nil {
		timber.Fatal(err, "failed to store API key")
	}
	timber.Done("Stored API key encrypted in system's keychain")
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Run:   auth,
	Short: "Login using a Synthient API key",
	Args:  cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
