package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/synthient"
	"github.com/zalando/go-keyring"
	"go.mattglei.ch/timber"
)

func runAuth(cmd *cobra.Command, args []string) {
	currentKey, err := synthient.ReadApiKey()
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		timber.Fatal(err, "unexpected error when checking for existing key")
	}
	if currentKey != "" {
		var overwrite bool
		err = huh.NewConfirm().
			Title("API key already saved. Do you want to overwrite?").
			Value(&overwrite).
			WithTheme(output.HuhTheme).
			Run()
		if err != nil {
			timber.Fatal(err, "failed to confirm overwriting api key")
		}
		if !overwrite {
			return
		}
	}

	fmt.Print("Please provide Synthient API key: ")
	var key string
	err = huh.NewInput().
		EchoMode(huh.EchoModePassword).
		Title("Synthient API key").
		Value(&key).
		WithTheme(output.HuhTheme).
		Run()
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
	Run:   runAuth,
	Short: "Login using a Synthient API key",
	Args:  cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
