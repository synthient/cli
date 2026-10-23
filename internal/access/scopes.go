package access

import (
	"context"
	"fmt"
	"slices"

	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/synthient-go"
)

type ScopeInfo struct {
	Scope  synthient.Scope `json:"scope"`
	Grants string          `json:"grants"`
}

type MissingScopeError struct {
	Scope synthient.Scope
}

func (e MissingScopeError) Error() string {
	return fmt.Sprintf("missing required scope %s", e.Scope)
}

type FeedScopes struct {
	Feed   synthient.Scope
	Stream synthient.Scope
}

var Scopes = []ScopeInfo{
	{Scope: synthient.ScopeBasic, Grants: "Account, IP lookup, batch IP lookup, and domain lookup."},
	{Scope: synthient.ScopeProxyFeeds, Grants: "proxies parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeProxyFirehose, Grants: "proxies real-time NDJSON stream."},
	{Scope: synthient.ScopeAnonymizersFeed, Grants: "anonymizers parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeAnonymizersStream, Grants: "anonymizers real-time NDJSON stream."},
	{Scope: synthient.ScopeTorrentsFeed, Grants: "torrents parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeTorrentsStream, Grants: "torrents real-time NDJSON stream."},
	{Scope: synthient.ScopeJA4TFeed, Grants: "ja4t parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeJA4TStream, Grants: "ja4t real-time NDJSON stream."},
	{Scope: synthient.ScopeHoneypotHTTPFeed, Grants: "honeypot_http parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeHoneypotHTTPStream, Grants: "honeypot_http real-time NDJSON stream."},
	{Scope: synthient.ScopeHoneypotHTTPSFeed, Grants: "honeypot_https parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeHoneypotHTTPSStream, Grants: "honeypot_https real-time NDJSON stream."},
	{Scope: synthient.ScopeHoneypotDNSFeed, Grants: "honeypot_dns parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeHoneypotDNSStream, Grants: "honeypot_dns real-time NDJSON stream."},
	{Scope: synthient.ScopeHoneypotADBFeed, Grants: "honeypot_adb parquet snapshot list, metadata, and downloads."},
	{Scope: synthient.ScopeHoneypotADBStream, Grants: "honeypot_adb real-time NDJSON stream."},
}

var StreamScopes = map[string]FeedScopes{
	"proxies":        {Feed: synthient.ScopeProxyFeeds, Stream: synthient.ScopeProxyFirehose},
	"anonymizers":    {Feed: synthient.ScopeAnonymizersFeed, Stream: synthient.ScopeAnonymizersStream},
	"torrents":       {Feed: synthient.ScopeTorrentsFeed, Stream: synthient.ScopeTorrentsStream},
	"ja4t":           {Feed: synthient.ScopeJA4TFeed, Stream: synthient.ScopeJA4TStream},
	"honeypot_http":  {Feed: synthient.ScopeHoneypotHTTPFeed, Stream: synthient.ScopeHoneypotHTTPStream},
	"honeypot_https": {Feed: synthient.ScopeHoneypotHTTPSFeed, Stream: synthient.ScopeHoneypotHTTPSStream},
	"honeypot_dns":   {Feed: synthient.ScopeHoneypotDNSFeed, Stream: synthient.ScopeHoneypotDNSStream},
	"honeypot_adb":   {Feed: synthient.ScopeHoneypotADBFeed, Stream: synthient.ScopeHoneypotADBStream},
}

func Required(stream feed.Stream, kind string) synthient.Scope {
	scopes := StreamScopes[stream.Name]
	if kind == "stream" {
		return scopes.Stream
	}
	return scopes.Feed
}

func Strings(scopes []synthient.Scope) []string {
	names := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		names = append(names, string(scope))
	}
	return names
}

func Require(client *synthient.Client, required synthient.Scope) error {
	if required == "" {
		return nil
	}
	account, err := client.Account(context.Background())
	if err != nil {
		return nil
	}
	if !slices.Contains(account.Scopes, required) {
		return MissingScopeError{Scope: required}
	}
	return nil
}
