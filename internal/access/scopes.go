package access

import (
	"fmt"
	"net/http"

	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/go-synthient/v2"
)

type ScopeInfo struct {
	Scope  string `json:"scope"`
	Grants string `json:"grants"`
}

type MissingScopeError struct {
	Scope string
}

func (e MissingScopeError) Error() string {
	return fmt.Sprintf("missing required scope %s", e.Scope)
}

type FeedScopes struct {
	Feed   string
	Stream string
}

var Scopes = []ScopeInfo{
	{Scope: "BASIC", Grants: "Account, IP lookup, batch IP lookup, and domain lookup."},
	{Scope: "PROXY_FEEDS", Grants: "proxies parquet snapshot list, metadata, and downloads."},
	{Scope: "PROXY_FIREHOSE", Grants: "proxies real-time NDJSON stream."},
	{Scope: "ANONYMIZERS_FEED", Grants: "anonymizers parquet snapshot list, metadata, and downloads."},
	{Scope: "ANONYMIZERS_STREAM", Grants: "anonymizers real-time NDJSON stream."},
	{Scope: "TORRENTS_FEED", Grants: "torrents parquet snapshot list, metadata, and downloads."},
	{Scope: "TORRENTS_STREAM", Grants: "torrents real-time NDJSON stream."},
	{Scope: "HONEYPOT_HTTP_FEED", Grants: "honeypot_http parquet snapshot list, metadata, and downloads."},
	{Scope: "HONEYPOT_HTTP_STREAM", Grants: "honeypot_http real-time NDJSON stream."},
	{Scope: "HONEYPOT_HTTPS_FEED", Grants: "honeypot_https parquet snapshot list, metadata, and downloads."},
	{Scope: "HONEYPOT_HTTPS_STREAM", Grants: "honeypot_https real-time NDJSON stream."},
	{Scope: "HONEYPOT_DNS_FEED", Grants: "honeypot_dns parquet snapshot list, metadata, and downloads."},
	{Scope: "HONEYPOT_DNS_STREAM", Grants: "honeypot_dns real-time NDJSON stream."},
	{Scope: "HONEYPOT_ADB_FEED", Grants: "honeypot_adb parquet snapshot list, metadata, and downloads."},
	{Scope: "HONEYPOT_ADB_STREAM", Grants: "honeypot_adb real-time NDJSON stream."},
}

var StreamScopes = map[string]FeedScopes{
	"proxies":        {Feed: "PROXY_FEEDS", Stream: "PROXY_FIREHOSE"},
	"anonymizers":    {Feed: "ANONYMIZERS_FEED", Stream: "ANONYMIZERS_STREAM"},
	"torrents":       {Feed: "TORRENTS_FEED", Stream: "TORRENTS_STREAM"},
	"honeypot_http":  {Feed: "HONEYPOT_HTTP_FEED", Stream: "HONEYPOT_HTTP_STREAM"},
	"honeypot_https": {Feed: "HONEYPOT_HTTPS_FEED", Stream: "HONEYPOT_HTTPS_STREAM"},
	"honeypot_dns":   {Feed: "HONEYPOT_DNS_FEED", Stream: "HONEYPOT_DNS_STREAM"},
	"honeypot_adb":   {Feed: "HONEYPOT_ADB_FEED", Stream: "HONEYPOT_ADB_STREAM"},
}

func Required(stream feed.Stream, kind string) string {
	scopes := StreamScopes[stream.Name]
	if kind == "stream" {
		return scopes.Stream
	}
	return scopes.Feed
}

func Has(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}

func Require(client synthient.Client, required string) error {
	if required == "" {
		return nil
	}
	path, err := feed.Path(client.BaseAPI, "account", "me")
	if err != nil {
		return err
	}
	req, err := feed.NewRequest(client, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	resp, err := feed.Do(client, req, http.StatusOK)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	account, err := feed.DecodeJSON[synthient.Account](resp.Body)
	if err != nil {
		return err
	}
	if !Has(account.Scopes, required) {
		return MissingScopeError{Scope: required}
	}
	return nil
}
