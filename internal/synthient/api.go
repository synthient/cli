package synthient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/synthient/cli/internal/conf"
	"go.mattglei.ch/timber"
)

const env_var_key = "SYNTHIENT_API_KEY"

// Error encountered when making a request to the API. Given when the response is not within the
// 200-299 range
var ErrApi = errors.New("error from API")

type Client struct {
	ApiKey     string
	Base       *url.URL
	httpClient *http.Client
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func CreateClient(config conf.Config) (Client, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		timber.Fatal(err, "failed to load .env file")
	}

	apiKey := strings.TrimSpace(os.Getenv(env_var_key))
	if apiKey == "" {
		ring, err := OpenKeyring()
		if err != nil {
			return Client{}, fmt.Errorf("%w opening keyring failed", err)
		}
		apiKey, err = ReadApiKey(ring)
		if err != nil {
			return Client{}, fmt.Errorf("%w reading api key failed", err)
		}
	}

	if apiKey == "" {
		timber.ErrorMsg(
			fmt.Sprintf(
				"User is not logged in. Please provide a API Key via %s or by logging in with `synthient auth`",
				env_var_key,
			),
		)
		os.Exit(1)
	}

	base, err := url.Parse(config.Host)
	if err != nil {
		timber.Fatal(err, "failed to parse host of:", config.Host)
	}

	return Client{
		ApiKey:     apiKey,
		Base:       base,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}, nil
}

func synthientRequest[T any](client *Client, request *http.Request) (T, error) {
	var zeroValue T // to be used as "nil"

	request.Header.Add("Authorization", client.ApiKey)

	resp, err := client.httpClient.Do(request)
	if err != nil {
		return zeroValue, fmt.Errorf("%w performing request failed", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroValue, fmt.Errorf("%w reading response body failed", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return zeroValue, fmt.Errorf("%w closing response body failed", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var data ErrorResponse
		err = json.Unmarshal(body, &data)
		if err != nil {
			return zeroValue, fmt.Errorf("%w parsing json for error response failed", err)
		}
		return zeroValue, fmt.Errorf("%w %s", ErrApi, data.Error)
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return zeroValue, fmt.Errorf("%w parsing json failed", err)
	}
	return data, nil
}
