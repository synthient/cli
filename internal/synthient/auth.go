package synthient

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyring_service_name = "synthient"
	keyring_key_name     = "token"
)

func ReadApiKey() (string, error) {
	v, err := keyring.Get(keyring_service_name, keyring_key_name)
	if err != nil {
		return "", err
	}
	return v, nil
}

func StoreApiKey(key string) error {
	err := keyring.Set(keyring_service_name, keyring_key_name, key)
	if err != nil {
		return fmt.Errorf("%w write api key to keyring", err)
	}

	return nil
}
