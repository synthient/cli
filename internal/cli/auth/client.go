package auth

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/go-synthient/v2"
	"github.com/zalando/go-keyring"
)

var ErrNoCredentials = errors.New("no synthient api key found")

type Credentials struct {
	Key    string
	Source string
}

func SynthientClient(config conf.Config) (synthient.Client, error) {
	credentials, err := ReadCredentials()
	if err != nil {
		return synthient.Client{}, err
	}
	return synthient.NewClient(credentials.Key), nil
}

func ReadCredentials() (Credentials, error) {
	key := strings.TrimSpace(os.Getenv("SYNTHIENT_API_KEY"))
	if key != "" {
		return Credentials{Key: key, Source: "env"}, nil
	}

	key = readDotEnvKey()
	if key != "" {
		return Credentials{Key: key, Source: ".env"}, nil
	}

	keyringKey, err := ReadKeyring()
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return Credentials{}, ErrNoCredentials
		}
		return Credentials{}, fmt.Errorf("getting api key from keyring: %w", err)
	}
	return Credentials{Key: keyringKey, Source: "keychain"}, nil
}

func ReadCredentialsStatus() Credentials {
	key := strings.TrimSpace(os.Getenv("SYNTHIENT_API_KEY"))
	if key != "" {
		return Credentials{Key: key, Source: "env"}
	}

	key = readDotEnvKey()
	if key != "" {
		return Credentials{Key: key, Source: ".env"}
	}

	keyringKey, err := ReadKeyring()
	if err == nil && keyringKey != "" {
		return Credentials{Key: keyringKey, Source: "keychain"}
	}
	return Credentials{Source: "missing"}
}

func readDotEnvKey() string {
	file, err := os.Open(".env")
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "SYNTHIENT_API_KEY" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}
