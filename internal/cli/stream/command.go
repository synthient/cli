package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/access"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/go-synthient/v2"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "stream <stream>",
	Short: "Stream a feed to stdout",
	Long:  "Stream newline-delimited JSON from proxies, anonymizers, torrents, and Helios honeypot feeds.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stream, ok := feed.Find(args[0])
		if !ok {
			timber.FatalMsg("unknown stream", timber.A("value", args[0]), timber.A("valid", feed.Names()))
		}
		validateFilters()
		if flags.duration < 0 {
			timber.FatalMsg("duration must be positive", timber.A("duration", flags.duration.String()))
		}

		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			app.Fatal(err, "failed to create synthient client")
		}
		config.ApplyToClient(&client)

		if !flags.noPreflight {
			err = access.Require(client, access.Required(stream, "stream"))
			if err != nil {
				app.Fatal(err, "failed stream access preflight")
			}
		}

		out, closeOut := output.Open(flags.output)
		defer closeOut()

		ctx := context.Background()
		if flags.duration > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, flags.duration)
			defer cancel()
		}

		if !flags.reconnect {
			seq, buildErr := buildSeq(ctx, &client, stream)
			if buildErr != nil {
				app.Fatal(buildErr, "failed to initialize stream")
			}
			result := streamOnce(ctx, seq, out, flags.maxEvents)
			if result.Err != nil {
				app.Fatal(result.Err, "failed to stream feed")
			}
			return
		}

		attempt := 0
		total := 0
		for {
			remaining := 0
			if flags.maxEvents > 0 {
				remaining = flags.maxEvents - total
				if remaining <= 0 {
					return
				}
			}
			seq, buildErr := buildSeq(ctx, &client, stream)
			if buildErr != nil {
				app.Fatal(buildErr, "failed to initialize stream")
			}
			result := streamOnce(ctx, seq, out, remaining)
			total += result.Count
			if result.Done {
				return
			}
			err = result.Err
			if err == nil {
				attempt = 0
				continue
			}
			if errors.Is(err, synthient.ErrBadRequest) ||
				errors.Is(err, synthient.ErrUnauthorized) ||
				errors.Is(err, synthient.ErrPaymentRequired) ||
				errors.Is(err, synthient.ErrNoToken) {
				app.Fatal(err, "failed to stream feed")
			}
			delay := retryDelay(attempt)
			timber.Warning("stream disconnected, reconnecting", timber.A("stream", stream.Name), timber.A("delay", delay.String()))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
			attempt++
		}
	},
}

type streamResult struct {
	Count int
	Done  bool
	Err   error
}

func buildSeq(ctx context.Context, client *synthient.Client, s feed.Stream) (iter.Seq2[json.RawMessage, error], error) {
	opts := &synthient.RequestOptions{Context: ctx}
	switch s.Name {
	case "proxies":
		return toRaw(client.StreamProxy(opts)), nil
	case "anonymizers":
		return toRaw(client.StreamAnonymizer(opts)), nil
	case "torrents":
		return toRaw(client.StreamTorrent(opts)), nil
	case "honeypot_http":
		return toRaw(client.StreamHeliosHTTP(opts)), nil
	case "honeypot_https":
		return toRaw(client.StreamHeliosTLS(opts)), nil
	default:
		return nil, fmt.Errorf("real-time stream not supported for %q", s.Name)
	}
}

func toRaw[T any](seq iter.Seq2[T, error]) iter.Seq2[json.RawMessage, error] {
	return func(yield func(json.RawMessage, error) bool) {
		for event, err := range seq {
			if err != nil {
				yield(nil, err)
				return
			}
			b, marshalErr := json.Marshal(event)
			if !yield(b, marshalErr) {
				return
			}
		}
	}
}

func streamOnce(ctx context.Context, seq iter.Seq2[json.RawMessage, error], out *os.File, maxEvents int) streamResult {
	count := 0
	for raw, err := range seq {
		if err != nil {
			if ctx.Err() != nil {
				return streamResult{Count: count, Done: true}
			}
			return streamResult{Count: count, Err: err}
		}
		line := string(raw)
		event, err := parseEvent(line)
		if err != nil {
			return streamResult{Count: count, Err: err}
		}
		if !matchesFilters(event) {
			continue
		}
		err = writeEvent(out, line, event)
		if err != nil {
			return streamResult{Count: count, Err: err}
		}
		count++
		if maxEvents > 0 && count >= maxEvents {
			return streamResult{Count: count, Done: true}
		}
	}
	return streamResult{Count: count}
}

func retryDelay(attempt int) time.Duration {
	seconds := math.Min(60, math.Pow(2, float64(attempt)))
	jitter := 0.75 + rand.Float64()*0.5
	return time.Duration(seconds * jitter * float64(time.Second))
}

func parseEvent(line string) (map[string]any, error) {
	var event map[string]any
	err := json.Unmarshal([]byte(line), &event)
	if err != nil {
		return nil, fmt.Errorf("parsing stream event: %w", err)
	}
	return event, nil
}

func matchesFilters(event map[string]any) bool {
	for _, raw := range flags.filters {
		field, expected, _ := strings.Cut(raw, "=")
		actual, ok := fieldValue(event, field)
		if !ok || fmt.Sprint(actual) != expected {
			return false
		}
	}
	return true
}

func validateFilters() {
	for _, raw := range flags.filters {
		field, expected, ok := strings.Cut(raw, "=")
		if !ok || strings.TrimSpace(field) == "" || strings.TrimSpace(expected) == "" {
			timber.FatalMsg("invalid filter, expected field=value", timber.A("filter", raw))
		}
	}
}

func fieldValue(event map[string]any, field string) (any, bool) {
	parts := strings.Split(field, ".")
	var current any = event
	for _, part := range parts {
		data, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = data[part]
		if !ok {
			return nil, false
		}
	}
	switch v := current.(type) {
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10), true
		}
	}
	return current, true
}

func writeEvent(out *os.File, line string, event map[string]any) error {
	if !flags.pretty {
		_, err := fmt.Fprintln(out, line)
		return err
	}
	b, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(b))
	return err
}
