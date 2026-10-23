package feed

import (
	"context"
	"encoding/json"
	"io"
	"iter"

	"github.com/synthient/synthient-go"
)

type Feed interface {
	ListExports(ctx context.Context, params ...synthient.ListExportsParams) (*synthient.ExportPage, error)
	ExportMeta(ctx context.Context, id string) (*synthient.ExportMeta, error)
	Download(ctx context.Context, id string) (io.ReadCloser, error)
	Events(ctx context.Context) iter.Seq2[json.RawMessage, error]
}

type sdkFeed[T any] struct {
	*synthient.Feed[T]
}

func wrap[T any](feed *synthient.Feed[T]) Feed {
	return sdkFeed[T]{feed}
}

func (f sdkFeed[T]) Events(ctx context.Context) iter.Seq2[json.RawMessage, error] {
	return func(yield func(json.RawMessage, error) bool) {
		for event, err := range f.Stream(ctx) {
			if err != nil {
				yield(nil, err)
				return
			}
			b, err := json.Marshal(event)
			if !yield(b, err) {
				return
			}
		}
	}
}

func (s Stream) Feed(client *synthient.Client) Feed {
	switch s.Name {
	case "proxies":
		return wrap(client.Proxies)
	case "anonymizers":
		return wrap(client.Anonymizers)
	case "torrents":
		return wrap(client.Torrents)
	case "ja4t":
		return wrap(client.JA4T)
	case "honeypot_http":
		return wrap(client.HeliosHTTP)
	case "honeypot_https":
		return wrap(client.HeliosHTTPS)
	case "honeypot_dns":
		return wrap(client.HeliosDNS)
	case "honeypot_adb":
		return wrap(client.HeliosADB)
	default:
		return nil
	}
}
