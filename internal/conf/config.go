package conf

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"go.mattglei.ch/timber"
)

type Config struct {
	Endpoints struct {
		BaseAPI   string `toml:"base_api"`
		BaseFeeds string `toml:"base_feeds"`
	} `toml:"endpoints"`

	// fields that are created based on existing fields
	BaseApiURL   *url.URL `toml:"-"`
	BaseFeedsURL *url.URL `toml:"-"`
}

func Read() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		timber.Fatal(err, "failed to get the user's home directory")
	}

	path := filepath.Join(home, ".config", "synthient", "config.toml")
	bin, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("reading from %s: %w", path, err)
	}

	var data Config
	err = toml.Unmarshal(bin, &data)
	if err != nil {
		return Config{}, fmt.Errorf("parsing toml: %w", err)
	}

	data.BaseApiURL, err = parseRawURL(data.Endpoints.BaseAPI)
	if err != nil {
		return Config{}, fmt.Errorf("parsing base url: %w", err)
	}

	data.BaseFeedsURL, err = parseRawURL(data.Endpoints.BaseFeeds)
	if err != nil {
		return Config{}, fmt.Errorf("parsing feeds url: %w", err)
	}

	return data, nil
}

func parseRawURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return nil, nil
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("url parsing: %w", err)
	}
	return parsedURL, nil
}
