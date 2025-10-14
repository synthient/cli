package synthient

import (
	"fmt"

	"github.com/99designs/keyring"
)

const (
	keyring_item_name = "synthient-api-key"
)

func OpenKeyring() (keyring.Keyring, error) {
	ring, err := keyring.Open(keyring.Config{ServiceName: "synthient"})
	if err != nil {
		return nil, fmt.Errorf("%w failed to open keyring", err)
	}
	return ring, nil
}

func ReadApiKey(ring keyring.Keyring) (string, error) {
	v, err := ring.Get(keyring_item_name)
	if err != nil {
		return "", err
	}
	return string(v.Data), nil
}

func StoreApiKey(ring keyring.Keyring, key string) error {
	err := ring.Set(keyring.Item{Key: keyring_item_name, Data: []byte(key)})
	if err != nil {
		return fmt.Errorf("%w write api key to keyring", err)
	}

	return nil
}
