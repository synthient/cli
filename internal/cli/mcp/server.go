package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/synthient/cli/internal/cli"
	"github.com/synthient/go-synthient/v2"
)

// ipLookupInput is the argument schema for the lookup_ip tool.
type ipLookupInput struct {
	IPs []string `json:"ips" jsonschema:"one or more IPv4 or IPv6 addresses to look up"`
}

// ipLookupOutput is the structured result of the lookup_ip tool.
type ipLookupOutput struct {
	IPs []synthient.IP `json:"ips"`
}

// domainLookupInput is the argument schema for the lookup_domain tool.
type domainLookupInput struct {
	Domain string `json:"domain" jsonschema:"the domain name to look up"`
}

// run stands up the MCP server, registers the Synthient tools, and serves over
// stdio until the context is cancelled.
func run(ctx context.Context, client synthient.Client) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "synthient",
		Title:   "Synthient",
		Version: cli.Root.Version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "lookup_ip",
		Description: "Look up Synthient intelligence for one or more IP addresses, including network, location, and risk data.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in ipLookupInput) (*mcp.CallToolResult, ipLookupOutput, error) {
		if len(in.IPs) == 0 {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "no IP addresses provided"}}}, ipLookupOutput{}, nil
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "lookup_domain",
		Description: "Look up Synthient intelligence for a domain name.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in domainLookupInput) (*mcp.CallToolResult, synthient.Domain, error) {
		domain, err := client.GetDomain(in.Domain, nil)
		if err != nil {
			return nil, synthient.Domain{}, err
		}
		return nil, domain, nil
	})

	return server.Run(ctx, &mcp.StdioTransport{})
}
