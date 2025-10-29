package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/99designs/keyring"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/synthient"
	"go.mattglei.ch/timber"
)

func runAuth(cmd *cobra.Command, args []string) {
	ring, err := synthient.OpenKeyring()
	if err != nil {
		timber.Fatal(err, "failed to open keyring")
	}

	currentKey, err := synthient.ReadApiKey(ring)
	if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
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

	err = synthient.StoreApiKey(ring, key)
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
