package synthient

import (
	"fmt"

	"github.com/99designs/keyring"
)

const (
	keyring_service_name = "synthient"
	keyring_item_name    = "synthient-api-key"
)

func StoreApiKey(key string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyring_service_name,
	})
	if err != nil {
		return fmt.Errorf("%w open keyring", err)
	}

	err = ring.Set(keyring.Item{Key: keyring_item_name, Data: []byte(key)})
	if err != nil {
		return fmt.Errorf("%w write api key to keyring", err)
	}

	return nil
}
