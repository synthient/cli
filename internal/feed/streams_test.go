package feed

import (
	"testing"

	"github.com/synthient/synthient-go"
)

func TestNames(t *testing.T) {
	want := "proxies|anonymizers|torrents|ja4t|honeypot_http|honeypot_https|honeypot_dns|honeypot_adb"
	got := Names()
	if got != want {
		t.Fatalf("Names() = %q, want %q", got, want)
	}
}

func TestEveryStreamHasFeed(t *testing.T) {
	client, err := synthient.NewClient("test")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	for _, stream := range Streams {
		if stream.Feed(client) == nil {
			t.Errorf("stream %q has no feed", stream.Name)
		}
	}
}
