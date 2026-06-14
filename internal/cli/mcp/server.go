package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/synthient/cli/internal/cli"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/go-synthient/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/descriptorpb"
)

// run stands up the MCP server, registers the Synthient tools, and serves over
// stdio until the context is cancelled.
func run(ctx context.Context, config conf.Config, client synthient.Client) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "synthient",
		Title:   "Synthient",
		Version: cli.Root.Version,
	}, nil)

	registerLookupIP(server, client)
	registerLookupDomain(server, client)
	registerGetAccount(server, client)
	registerListFeedStreams(server)
	registerListFeedSnapshots(server, client)
	registerFeedSnapshotMeta(server, client)
	registerSampleStream(server, client)
	registerGRPCSchema(server, config, client)

	return server.Run(ctx, &mcp.StdioTransport{})
}

// --- IP lookup ---

type ipLookupInput struct {
	IPs []string `json:"ips" jsonschema:"one or more IPv4 or IPv6 addresses to look up"`
}

type ipLookupOutput struct {
	IPs []synthient.IP `json:"ips"`
}

func registerLookupIP(server *mcp.Server, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lookup_ip",
		Description: "Look up Synthient intelligence for one or more IP addresses, including network, location, and risk data.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in ipLookupInput) (*mcp.CallToolResult, ipLookupOutput, error) {
		if len(in.IPs) == 0 {
			return toolError("no IP addresses provided"), ipLookupOutput{}, nil
		}

		var ips []synthient.IP
		if len(in.IPs) == 1 {
			ip, err := client.GetIP(in.IPs[0], nil)
			if err != nil {
				return nil, ipLookupOutput{}, err
			}
			ips = []synthient.IP{ip}
		} else {
			resp, err := client.GetIPs(in.IPs, nil)
			if err != nil {
				return nil, ipLookupOutput{}, err
			}
			ips = resp
		}

		return nil, ipLookupOutput{IPs: ips}, nil
	})
}

// --- Domain lookup ---

type domainLookupInput struct {
	Domain string `json:"domain" jsonschema:"the domain name to look up"`
}

func registerLookupDomain(server *mcp.Server, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lookup_domain",
		Description: "Look up Synthient intelligence for a domain name.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in domainLookupInput) (*mcp.CallToolResult, synthient.Domain, error) {
		if strings.TrimSpace(in.Domain) == "" {
			return toolError("no domain provided"), synthient.Domain{}, nil
		}
		domain, err := client.GetDomain(in.Domain, nil)
		if err != nil {
			return nil, synthient.Domain{}, err
		}
		return nil, domain, nil
	})
}

// --- Account ---

func registerGetAccount(server *mcp.Server, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_account",
		Description: "Get the authenticated account: organization, email, granted scopes, and lookup quota (credits and reset timing).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, synthient.Account, error) {
		account, err := client.GetAccount(nil)
		if err != nil {
			return nil, synthient.Account{}, err
		}
		return nil, account, nil
	})
}

// --- Feed catalog ---

type listFeedStreamsOutput struct {
	Streams []feed.Stream `json:"streams"`
}

func registerListFeedStreams(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_feed_streams",
		Description: "List the available Synthient feed streams (proxies, anonymizers, torrents, and Helios honeypot feeds) with their descriptions and aliases.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listFeedStreamsOutput, error) {
		return nil, listFeedStreamsOutput{Streams: feed.Streams}, nil
	})
}

type listFeedSnapshotsInput struct {
	Stream string `json:"stream" jsonschema:"feed stream name, e.g. proxies, anonymizers, torrents, honeypot_http, honeypot_https, honeypot_dns, honeypot_adb"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; defaults to 100, capped at 500 by the API"`
	Cursor string `json:"cursor,omitempty" jsonschema:"pagination token from a previous page's next_cursor; omit on the first call"`
}

func registerListFeedSnapshots(server *mcp.Server, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_feed_snapshots",
		Description: "List available Parquet snapshots (daily and hourly) for a feed stream, newest first. Returns a page with a next_cursor for pagination.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listFeedSnapshotsInput) (*mcp.CallToolResult, synthient.FeedSnapshotsPage, error) {
		stream, ok := feed.Find(in.Stream)
		if !ok {
			return toolError("unknown stream %q; valid streams: %s", in.Stream, feed.Names()), synthient.FeedSnapshotsPage{}, nil
		}
		page, err := client.FeedSnapshots(stream.Name, &synthient.FeedSnapshotsOptions{Limit: in.Limit, Cursor: in.Cursor}, nil)
		if err != nil {
			return nil, synthient.FeedSnapshotsPage{}, err
		}
		return nil, page, nil
	})
}

type feedSnapshotMetaInput struct {
	Stream string `json:"stream" jsonschema:"feed stream name, e.g. proxies, anonymizers, torrents, honeypot_http, honeypot_https, honeypot_dns, honeypot_adb"`
	Date   string `json:"date" jsonschema:"snapshot identifier: YYYY-MM-DD for daily, YYYY-MM-DD/HH for an hourly, or 'latest' for the most recent hourly"`
}

func registerFeedSnapshotMeta(server *mcp.Server, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "feed_snapshot_meta",
		Description: "Get metadata for a single feed snapshot: SHA-256 checksum, byte size, row count, parquet schema, and canonical date.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in feedSnapshotMetaInput) (*mcp.CallToolResult, synthient.FeedSnapshotMeta, error) {
		stream, ok := feed.Find(in.Stream)
		if !ok {
			return toolError("unknown stream %q; valid streams: %s", in.Stream, feed.Names()), synthient.FeedSnapshotMeta{}, nil
		}
		if strings.TrimSpace(in.Date) == "" {
			return toolError("no date provided; use YYYY-MM-DD, YYYY-MM-DD/HH, or 'latest'"), synthient.FeedSnapshotMeta{}, nil
		}
		meta, err := client.FeedSnapshotMeta(stream.Name, in.Date, nil)
		if err != nil {
			return nil, synthient.FeedSnapshotMeta{}, err
		}
		return nil, meta, nil
	})
}

// --- Stream sampling ---

const (
	defaultSampleCount   = 10
	maxSampleCount       = 500
	defaultSampleTimeout = 15
	maxSampleTimeout     = 120
)

type sampleStreamInput struct {
	Stream         string `json:"stream" jsonschema:"real-time stream name: proxies, anonymizers, torrents, honeypot_http, or honeypot_https"`
	Count          int    `json:"count,omitempty" jsonschema:"number of events to collect before returning; defaults to 10, capped at 500"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"stop and return early if the stream is idle this long; defaults to 15, capped at 120"`
}

type sampleStreamOutput struct {
	Stream string           `json:"stream"`
	Count  int              `json:"count"`
	Events []map[string]any `json:"events"`
}

func registerSampleStream(server *mcp.Server, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sample_stream",
		Description: "Sample recent events from a real-time Synthient feed. Connects to the stream, collects up to `count` events (or until `timeout_seconds` elapses), then returns them.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in sampleStreamInput) (*mcp.CallToolResult, sampleStreamOutput, error) {
		stream, ok := feed.Find(in.Stream)
		if !ok {
			return toolError("unknown stream %q; valid streams: %s", in.Stream, feed.Names()), sampleStreamOutput{}, nil
		}

		count := in.Count
		if count <= 0 {
			count = defaultSampleCount
		}
		if count > maxSampleCount {
			count = maxSampleCount
		}

		timeout := in.TimeoutSeconds
		if timeout <= 0 {
			timeout = defaultSampleTimeout
		}
		if timeout > maxSampleTimeout {
			timeout = maxSampleTimeout
		}

		sampleCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		events, err := sampleStream(sampleCtx, &client, stream.Name, count)
		if err != nil {
			return nil, sampleStreamOutput{}, err
		}
		return nil, sampleStreamOutput{Stream: stream.Name, Count: len(events), Events: events}, nil
	})
}

func sampleStream(ctx context.Context, client *synthient.Client, stream string, count int) ([]map[string]any, error) {
	opts := &synthient.RequestOptions{Context: ctx}
	switch stream {
	case "proxies":
		return collectEvents(client.StreamProxy(opts), count)
	case "anonymizers":
		return collectEvents(client.StreamAnonymizer(opts), count)
	case "torrents":
		return collectEvents(client.StreamTorrent(opts), count)
	case "honeypot_http":
		return collectEvents(client.StreamHeliosHTTP(opts), count)
	case "honeypot_https":
		return collectEvents(client.StreamHeliosTLS(opts), count)
	default:
		return nil, fmt.Errorf("real-time sampling is not supported for stream %q", stream)
	}
}

// collectEvents drains up to count events from a stream iterator into generic
// JSON objects. A timeout or cancellation returns the events gathered so far.
func collectEvents[T any](seq iter.Seq2[T, error], count int) ([]map[string]any, error) {
	events := []map[string]any{}
	for event, err := range seq {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return events, nil
			}
			return events, err
		}
		raw, marshalErr := json.Marshal(event)
		if marshalErr != nil {
			return events, marshalErr
		}
		obj := map[string]any{}
		unmarshalErr := json.Unmarshal(raw, &obj)
		if unmarshalErr != nil {
			return events, unmarshalErr
		}
		events = append(events, obj)
		if len(events) >= count {
			break
		}
	}
	return events, nil
}

// --- gRPC schema ---

type grpcSchemaInput struct {
	Symbols           []string `json:"symbols,omitempty" jsonschema:"fully-qualified service or message names to resolve; omit to fetch every exposed service"`
	Endpoint          string   `json:"endpoint,omitempty" jsonschema:"override gRPC endpoint (host:port); defaults to the configured endpoint"`
	Plaintext         bool     `json:"plaintext,omitempty" jsonschema:"connect without TLS"`
	IncludeReflection bool     `json:"include_reflection,omitempty" jsonschema:"include the gRPC reflection service descriptors in the output"`
}

type grpcFile struct {
	Name    string `json:"name"`
	Package string `json:"package"`
}

type grpcMethod struct {
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type grpcService struct {
	Name    string       `json:"name"`
	Methods []grpcMethod `json:"methods"`
}

type grpcSchemaOutput struct {
	Endpoint   string        `json:"endpoint"`
	Symbols    []string      `json:"symbols"`
	Files      []grpcFile    `json:"files"`
	Services   []grpcService `json:"services"`
	Descriptor string        `json:"descriptor_json"`
}

func registerGRPCSchema(server *mcp.Server, config conf.Config, client synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "grpc_schema",
		Description: "Fetch Synthient's protobuf schema over gRPC reflection. Returns a summary of files, services, and methods plus the full descriptor set as protojson.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in grpcSchemaInput) (*mcp.CallToolResult, grpcSchemaOutput, error) {
		endpoint := strings.TrimSpace(in.Endpoint)
		if endpoint == "" {
			endpoint = config.GRPCEndpoint(synthient.DefaultGRPCEndpoint)
		}

		schemaCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		result, err := client.GRPCSchema(schemaCtx, &synthient.GRPCSchemaOptions{
			Endpoint:          endpoint,
			Plaintext:         in.Plaintext,
			IncludeReflection: in.IncludeReflection,
			Symbols:           in.Symbols,
		})
		if err != nil {
			if msg := synthient.ExplainGRPCError(err); msg != "" {
				return toolError("%s", msg), grpcSchemaOutput{}, nil
			}
			return nil, grpcSchemaOutput{}, err
		}

		descriptor, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(result.DescriptorSet)
		if err != nil {
			return nil, grpcSchemaOutput{}, fmt.Errorf("marshaling descriptor set: %w", err)
		}

		out := grpcSchemaOutput{
			Endpoint:   result.Endpoint,
			Symbols:    result.Symbols,
			Services:   summarizeServices(result.DescriptorSet),
			Descriptor: string(descriptor),
		}
		for _, file := range result.DescriptorSet.GetFile() {
			out.Files = append(out.Files, grpcFile{Name: file.GetName(), Package: file.GetPackage()})
		}
		return nil, out, nil
	})
}

func summarizeServices(set *descriptorpb.FileDescriptorSet) []grpcService {
	services := []grpcService{}
	for _, file := range set.GetFile() {
		pkg := file.GetPackage()
		for _, service := range file.GetService() {
			name := service.GetName()
			if pkg != "" {
				name = pkg + "." + name
			}
			summary := grpcService{Name: name}
			for _, method := range service.GetMethod() {
				summary.Methods = append(summary.Methods, grpcMethod{
					Name:   method.GetName(),
					Input:  strings.TrimPrefix(method.GetInputType(), "."),
					Output: strings.TrimPrefix(method.GetOutputType(), "."),
				})
			}
			services = append(services, summary)
		}
	}
	sort.Slice(services, func(i int, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services
}

// toolError builds a CallToolResult flagged as an error with a formatted message.
func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
	}
}
