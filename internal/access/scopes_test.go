package access

import (
	"testing"

	"github.com/synthient/cli/internal/feed"
)

func TestRequiredJA4Scopes(t *testing.T) {
	tests := []struct {
		stream      string
		feedScope   string
		streamScope string
	}{
		{stream: "ja4", feedScope: "JA4_FEED", streamScope: "JA4_STREAM"},
		{stream: "ja4t", feedScope: "JA4T_FEED", streamScope: "JA4T_STREAM"},
	}

	for _, test := range tests {
		stream, ok := feed.Find(test.stream)
		if !ok {
			t.Fatalf("feed.Find(%q) did not find stream", test.stream)
		}
		if got := Required(stream, "feed"); got != test.feedScope {
			t.Errorf("Required(%q, feed) = %q, want %q", test.stream, got, test.feedScope)
		}
		if got := Required(stream, "stream"); got != test.streamScope {
			t.Errorf("Required(%q, stream) = %q, want %q", test.stream, got, test.streamScope)
		}
	}
}
