package conf

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"go.mattglei.ch/timber"
)

type Config struct {
	Host string `toml:"host"`
}

func Read() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		timber.Fatal(err, "failed to get the user's home directory")
	}

	bin, err := os.ReadFile(filepath.Join(home, ".config", "synthient", "config.toml"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{Host: "https://v3api.synthient.com/api/v3"}
		}
		timber.Fatal(err, "failed to read from config file")
	}

	var data Config
	err = toml.Unmarshal(bin, &data)
	if err != nil {
		timber.Fatal(err, "failed to unmarshal toml data")
	}
	return data
}
