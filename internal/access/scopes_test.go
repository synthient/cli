package access

import (
	"testing"

	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/synthient-go"
)

func TestEveryStreamHasScopes(t *testing.T) {
	for _, stream := range feed.Streams {
		if Required(stream, "feed") == "" {
			t.Errorf("stream %q has no feed scope", stream.Name)
		}
		if Required(stream, "stream") == "" {
			t.Errorf("stream %q has no stream scope", stream.Name)
		}
	}
}

func TestRequiredJA4TScopes(t *testing.T) {
	stream, ok := feed.Find("ja4t")
	if !ok {
		t.Fatal("feed.Find(\"ja4t\") did not find stream")
	}
	if got := Required(stream, "feed"); got != synthient.ScopeJA4TFeed {
		t.Errorf("Required(ja4t, feed) = %q, want %q", got, synthient.ScopeJA4TFeed)
	}
	if got := Required(stream, "stream"); got != synthient.ScopeJA4TStream {
		t.Errorf("Required(ja4t, stream) = %q, want %q", got, synthient.ScopeJA4TStream)
	}
}
