package stream

import (
	"context"
	"testing"

	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/go-synthient/v2"
)

func TestBuildSeqSupportsJA4TStream(t *testing.T) {
	client := synthient.NewClient("test")
	tests := []string{"ja4t"}

	for _, name := range tests {
		stream, ok := feed.Find(name)
		if !ok {
			t.Fatalf("feed.Find(%q) did not find stream", name)
		}
		seq, err := buildSeq(context.Background(), &client, stream)
		if err != nil {
			t.Fatalf("buildSeq(%q) returned error: %v", name, err)
		}
		if seq == nil {
			t.Fatalf("buildSeq(%q) returned a nil iterator", name)
		}
	}
}
