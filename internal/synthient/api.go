package synthient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mattglei.ch/timber"
)

const env_var_key = "SYNTHIENT_API_KEY"

type Client struct {
	ApiKey     string
	httpClient *http.Client
}

func CreateClient() (Client, error) {
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

	return Client{ApiKey: apiKey, httpClient: &http.Client{Timeout: 20 * time.Second}}, nil
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

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return zeroValue, fmt.Errorf("%w parsing json failed", err)
	}
	return data, nil
}
