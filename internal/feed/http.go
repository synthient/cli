package feed

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/synthient/go-synthient/v2"
)

type StatusError struct {
	StatusCode int
	Status     string
	Body       string
	RetryAfter string
}

func (e StatusError) Error() string {
	body := strings.TrimSpace(e.Body)
	if body == "" {
		return e.Status
	}
	return fmt.Sprintf("%s: %s", e.Status, body)
}

func NewRequest(client synthient.Client, method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	req.Header.Set("x-api-key", client.Token)
	return req, nil
}

func Do(client synthient.Client, req *http.Request, expected int) (*http.Response, error) {
	httpClient := client.HttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}

	if resp.StatusCode != expected {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		return nil, StatusError{
			StatusCode: resp.StatusCode,
			Status:     fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
			Body:       string(body),
			RetryAfter: resp.Header.Get("Retry-After"),
		}
	}

	return resp, nil
}

func GetJSON[T any](client synthient.Client, path string) (T, error) {
	var zero T
	req, err := NewRequest(client, http.MethodGet, path, nil)
	if err != nil {
		return zero, err
	}
	resp, err := Do(client, req, http.StatusOK)
	if err != nil {
		return zero, err
	}
	defer func() { _ = resp.Body.Close() }()

	value, err := DecodeJSON[T](resp.Body)
	if err != nil {
		return zero, err
	}
	return value, nil
}

func DecodeJSON[T any](body io.Reader) (T, error) {
	var zero T
	var value T
	err := json.NewDecoder(body).Decode(&value)
	if err != nil {
		return zero, fmt.Errorf("parsing json: %w", err)
	}
	return value, nil
}
