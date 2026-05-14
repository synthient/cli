package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/go-synthient/v2"
	"github.com/zalando/go-keyring"
	"go.mattglei.ch/timber"
)

func SynthientClient(config conf.Config) (synthient.Client, error) {
	key := os.Getenv("SYNTHIENT_API_KEY")
	if key == "" {
		keyringKey, err := ReadKeyring()
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				timber.ErrorMsg("please login using `synthient auth`")
				os.Exit(0)
			}
			return synthient.Client{}, fmt.Errorf("getting api key from keyring: %w", err)
		}
		key = keyringKey
	}
	return synthient.NewClient(key), nil
}
