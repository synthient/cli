package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/netip"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/synthient/cli/internal/cli"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/cli/internal/utils"
	"github.com/synthient/synthient-go"
	"github.com/synthient/synthient-go/grpcschema"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/descriptorpb"
)

// run stands up the MCP server, registers the Synthient tools, and serves over
// stdio until the context is cancelled.
func run(ctx context.Context, config conf.Config, client *synthient.Client) error {
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
	registerGRPCSchema(server, config)

	return server.Run(ctx, &mcp.StdioTransport{})
}

// --- IP lookup ---

type ipLookupInput struct {
	IPs []string `json:"ips" jsonschema:"one or more IPv4 or IPv6 addresses to look up"`
}

type ipLookupOutput struct {
	IPs []synthient.IPInfo `json:"ips"`
}

// netip.Addr marshals to a string, but its inferred schema is an empty object.
func ipLookupSchema() (*jsonschema.Schema, error) {
	return jsonschema.For[ipLookupOutput](&jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[netip.Addr](): {Type: "string"},
		},
	})
}

func registerLookupIP(server *mcp.Server, client *synthient.Client) {
	schema, err := ipLookupSchema()
	if err != nil {
		panic(fmt.Sprintf("building lookup_ip output schema: %v", err))
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:         "lookup_ip",
		Description:  "Look up Synthient intelligence for one or more IP addresses, including network, location, and risk data.",
		OutputSchema: schema,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ipLookupInput) (*mcp.CallToolResult, ipLookupOutput, error) {
		if len(in.IPs) == 0 {
			return toolError("no IP addresses provided"), ipLookupOutput{}, nil
		}
		addrs, err := utils.ParseIPs(in.IPs)
		if err != nil {
			return toolError("%s", err), ipLookupOutput{}, nil
		}

		var ips []synthient.IPInfo
		if len(addrs) == 1 {
			ip, err := client.LookupIP(ctx, addrs[0])
			if err != nil {
				return nil, ipLookupOutput{}, err
			}
			ips = []synthient.IPInfo{*ip}
		} else {
			resp, err := client.LookupIPs(ctx, addrs)
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

func registerLookupDomain(server *mcp.Server, client *synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "lookup_domain",
		Description: "Look up Synthient intelligence for a domain name.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in domainLookupInput) (*mcp.CallToolResult, synthient.DomainInfo, error) {
		if strings.TrimSpace(in.Domain) == "" {
			return toolError("no domain provided"), synthient.DomainInfo{}, nil
		}
		domain, err := client.LookupDomain(ctx, in.Domain)
		if err != nil {
			return nil, synthient.DomainInfo{}, err
		}
		return nil, *domain, nil
	})
}

// --- Account ---

func registerGetAccount(server *mcp.Server, client *synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_account",
		Description: "Get the authenticated account: organization, email, granted scopes, and lookup quota (credits and reset timing).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, synthient.Account, error) {
		account, err := client.Account(ctx)
		if err != nil {
			return nil, synthient.Account{}, err
		}
		return nil, *account, nil
	})
}

// --- Feed catalog ---

type listFeedStreamsOutput struct {
	Streams []feed.Stream `json:"streams"`
}

func registerListFeedStreams(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_feed_streams",
		Description: "List the available Synthient feed streams (proxies, anonymizers, torrents, JA4T, and Helios honeypot feeds) with their descriptions and aliases.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listFeedStreamsOutput, error) {
		return nil, listFeedStreamsOutput{Streams: feed.Streams}, nil
	})
}

type listFeedSnapshotsInput struct {
	Stream string `json:"stream" jsonschema:"feed stream name, e.g. proxies, anonymizers, torrents, ja4t, honeypot_http, honeypot_https, honeypot_dns, honeypot_adb"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; defaults to 100, capped at 500 by the API"`
	Cursor string `json:"cursor,omitempty" jsonschema:"pagination token from a previous page's next_cursor; omit on the first call"`
}

func registerListFeedSnapshots(server *mcp.Server, client *synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_feed_snapshots",
		Description: "List available Parquet snapshots (daily and hourly) for a feed stream, newest first. Returns a page with a next_cursor for pagination.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listFeedSnapshotsInput) (*mcp.CallToolResult, synthient.ExportPage, error) {
		stream, ok := feed.Find(in.Stream)
		if !ok {
			return toolError("unknown stream %q; valid streams: %s", in.Stream, feed.Names()), synthient.ExportPage{}, nil
		}
		page, err := stream.Feed(client).ListExports(ctx, synthient.ListExportsParams{Limit: in.Limit, Cursor: in.Cursor})
		if err != nil {
			return nil, synthient.ExportPage{}, err
		}
		return nil, *page, nil
	})
}

type feedSnapshotMetaInput struct {
	Stream string `json:"stream" jsonschema:"feed stream name, e.g. proxies, anonymizers, torrents, ja4t, honeypot_http, honeypot_https, honeypot_dns, honeypot_adb"`
	Date   string `json:"date" jsonschema:"snapshot identifier: YYYY-MM-DD for daily, YYYY-MM-DD/HH for an hourly, or 'latest' for the most recent hourly"`
}

func registerFeedSnapshotMeta(server *mcp.Server, client *synthient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "feed_snapshot_meta",
		Description: "Get metadata for a single feed snapshot: SHA-256 checksum, byte size, row count, parquet schema, and canonical date.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in feedSnapshotMetaInput) (*mcp.CallToolResult, synthient.ExportMeta, error) {
		stream, ok := feed.Find(in.Stream)
		if !ok {
			return toolError("unknown stream %q; valid streams: %s", in.Stream, feed.Names()), synthient.ExportMeta{}, nil
		}
		if strings.TrimSpace(in.Date) == "" {
			return toolError("no date provided; use YYYY-MM-DD, YYYY-MM-DD/HH, or 'latest'"), synthient.ExportMeta{}, nil
		}
		meta, err := stream.Feed(client).ExportMeta(ctx, in.Date)
		if err != nil {
			return nil, synthient.ExportMeta{}, err
		}
		return nil, *meta, nil
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
	Stream         string `json:"stream" jsonschema:"real-time stream name: proxies, anonymizers, torrents, ja4t, honeypot_http, honeypot_https, honeypot_dns, or honeypot_adb"`
	Count          int    `json:"count,omitempty" jsonschema:"number of events to collect before returning; defaults to 10, capped at 500"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"stop and return early if the stream is idle this long; defaults to 15, capped at 120"`
}

type sampleStreamOutput struct {
	Stream string           `json:"stream"`
	Count  int              `json:"count"`
	Events []map[string]any `json:"events"`
}

func registerSampleStream(server *mcp.Server, client *synthient.Client) {
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

		events, err := collectEvents(stream.Feed(client).Events(sampleCtx), count)
		if err != nil {
			return nil, sampleStreamOutput{}, err
		}
		return nil, sampleStreamOutput{Stream: stream.Name, Count: len(events), Events: events}, nil
	})
}

// collectEvents drains up to count events from a stream iterator into generic
// JSON objects. A timeout or cancellation returns the events gathered so far.
func collectEvents(seq iter.Seq2[json.RawMessage, error], count int) ([]map[string]any, error) {
	events := []map[string]any{}
	for event, err := range seq {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return events, nil
			}
			return events, err
		}
		obj := map[string]any{}
		unmarshalErr := json.Unmarshal(event, &obj)
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
	Symbols   []string `json:"symbols,omitempty" jsonschema:"fully-qualified service or message names to resolve; omit to fetch every exposed service"`
	Endpoint  string   `json:"endpoint,omitempty" jsonschema:"override gRPC endpoint (host:port); defaults to the configured endpoint"`
	Plaintext bool     `json:"plaintext,omitempty" jsonschema:"connect without TLS"`
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

func registerGRPCSchema(server *mcp.Server, config conf.Config) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "grpc_schema",
		Description: "Fetch Synthient's protobuf schema over gRPC reflection. Returns a summary of files, services, and methods plus the full descriptor set as protojson.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in grpcSchemaInput) (*mcp.CallToolResult, grpcSchemaOutput, error) {
		credentials, err := auth.ReadCredentials()
		if err != nil {
			return nil, grpcSchemaOutput{}, err
		}

		endpoint := strings.TrimSpace(in.Endpoint)
		if endpoint == "" {
			endpoint = config.GRPCEndpoint(grpcschema.DefaultEndpoint)
		}

		schemaCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		result, err := grpcschema.Fetch(schemaCtx, credentials.Key, grpcschema.Options{
			Endpoint:  endpoint,
			Plaintext: in.Plaintext,
			Symbols:   in.Symbols,
		})
		if err != nil {
			if msg := grpcschema.Explain(err); msg != "" {
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
