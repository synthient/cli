package auth

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyring_service_name = "synthient"
	keyring_key_name     = "token"
)

func ReadKeyring() (string, error) {
	keyringKey, err := keyring.Get(keyring_service_name, keyring_key_name)
	if err != nil {
		return "", fmt.Errorf("getting api key from keyring: %w", err)
	}
	return keyringKey, nil
}

func StoreApiKey(key string) error {
	err := keyring.Set(keyring_service_name, keyring_key_name, key)
	if err != nil {
		return fmt.Errorf("writing api key to keyring: %w", err)
	}
	return nil
}
