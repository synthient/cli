package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/output"
	"github.com/zalando/go-keyring"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "auth",
	Short: "Login using a Synthient API key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		styles := output.NewStyles(os.Stdout)
		currentKey, err := ReadKeyring()
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			timber.Fatal(err, "unexpected error when checking for existing key")
		}
		if flags.logout {
			if currentKey == "" {
				output.Warn(os.Stdout, styles, "User is not logged in, no credentials to remove")
			} else {
				err := StoreApiKey("")
				if err != nil {
					timber.Fatal(err, "failed to reset api key")
				}
				output.Done(os.Stdout, styles, "Successfully logged out")
			}
			return
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

		var key string
		err = huh.NewInput().
			EchoMode(huh.EchoModePassword).
			Title("Enter your Synthient API key").
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

		err = StoreApiKey(key)
		if err != nil {
			timber.Fatal(err, "failed to store API key")
		}
		fmt.Fprint(os.Stdout, "\033[1A\033[2K")
		output.Done(os.Stdout, styles, "Stored API key encrypted in system keychain")
	},
}
