package feed

import (
	"slices"
	"strings"
)

type Stream struct {
	Name        string   `json:"name"`
	ListName    string   `json:"list_name"`
	Path        []string `json:"-"`
	Aliases     []string `json:"aliases"`
	Description string   `json:"description"`
}

var Streams = []Stream{
	{
		Name:        "proxies",
		ListName:    "proxies",
		Path:        []string{"feeds", "proxies"},
		Aliases:     []string{"proxy"},
		Description: "Proxy IP observations across residential, datacenter, and mobile networks.",
	},
	{
		Name:        "anonymizers",
		ListName:    "anonymizers",
		Path:        []string{"feeds", "anonymizers"},
		Aliases:     []string{"anonymizer"},
		Description: "VPN, Tor, private relay, and other anonymizer ranges.",
	},
	{
		Name:        "torrents",
		ListName:    "torrents",
		Path:        []string{"feeds", "torrents"},
		Aliases:     []string{"torrent"},
		Description: "DHT and tracker peer sightings with metadata and observed peers.",
	},
	{
		Name:        "honeypot_http",
		ListName:    "honeypot_http",
		Path:        []string{"feeds", "helio", "http"},
		Aliases:     []string{"helios_http", "http"},
		Description: "HTTP request captures from Helios honeypot sensors.",
	},
	{
		Name:        "honeypot_https",
		ListName:    "honeypot_https",
		Path:        []string{"feeds", "helio", "https"},
		Aliases:     []string{"helios_https", "https", "tls"},
		Description: "TLS ClientHello captures from Helios honeypot sensors.",
	},
	{
		Name:        "honeypot_dns",
		ListName:    "honeypot_dns",
		Path:        []string{"feeds", "helio", "dns"},
		Aliases:     []string{"helios_dns", "dns"},
		Description: "DNS resolution observations from Helios honeypot tunnels.",
	},
	{
		Name:        "honeypot_adb",
		ListName:    "honeypot_adb",
		Path:        []string{"feeds", "helio", "adb"},
		Aliases:     []string{"helios_adb", "adb"},
		Description: "Android Debug Bridge shell commands captured by Helios sensors.",
	},
}

func Find(name string) (Stream, bool) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, stream := range Streams {
		if stream.Name == normalized || stream.ListName == normalized || slices.Contains(stream.Aliases, normalized) {
			return stream, true
		}
	}
	return Stream{}, false
}

func Names() string {
	names := make([]string, 0, len(Streams))
	for _, stream := range Streams {
		names = append(names, stream.Name)
	}
	return strings.Join(names, "|")
}
