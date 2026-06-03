package conf

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/synthient/cli/internal/options"
	"go.mattglei.ch/timber"
)

type Config struct {
	Endpoints struct {
		BaseAPI   string `toml:"base_api"`
		BaseFeeds string `toml:"base_feeds"`
		BaseGRPC  string `toml:"base_grpc"`
	} `toml:"endpoints"`

	// fields that are created based on existing fields
	BaseApiURL   *url.URL `toml:"-"`
	BaseFeedsURL *url.URL `toml:"-"`

	Profiles map[string]ProfileConfig `toml:"profiles"`
	Path     string                   `toml:"-"`
	Profile  string                   `toml:"-"`
}

type ProfileConfig struct {
	Endpoints struct {
		BaseAPI   string `toml:"base_api"`
		BaseFeeds string `toml:"base_feeds"`
		BaseGRPC  string `toml:"base_grpc"`
	} `toml:"endpoints"`
}

func Read() (Config, error) {
	path := Path()
	bin, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{Path: path, Profile: options.Profile}, nil
		}
		return Config{}, fmt.Errorf("reading from %s: %w", path, err)
	}

	var data Config
	err = toml.Unmarshal(bin, &data)
	if err != nil {
		return Config{}, fmt.Errorf("parsing toml: %w", err)
	}
	data.Path = path
	data.Profile = options.Profile

	if options.Profile != "" {
		profile, ok := data.Profiles[options.Profile]
		if !ok {
			return Config{}, fmt.Errorf("profile %q not found", options.Profile)
		}
		if profile.Endpoints.BaseAPI != "" {
			data.Endpoints.BaseAPI = profile.Endpoints.BaseAPI
		}
		if profile.Endpoints.BaseFeeds != "" {
			data.Endpoints.BaseFeeds = profile.Endpoints.BaseFeeds
		}
		if profile.Endpoints.BaseGRPC != "" {
			data.Endpoints.BaseGRPC = profile.Endpoints.BaseGRPC
		}
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

func Path() string {
	if options.ConfigPath != "" {
		return options.ConfigPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		timber.Fatal(err, "failed to get the user's home directory")
	}
	return filepath.Join(home, ".config", "synthient", "config.toml")
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
