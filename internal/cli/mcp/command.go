package mcp

import (
	"context"
	"errors"
	"io"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/app"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "mcp",
	Short: "Run a Model Context Protocol server backed by Synthient",
	Long:  "Run a Model Context Protocol (MCP) server that exposes Synthient lookups as tools for MCP-compatible clients. The server communicates over stdio.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := conf.Read()
		if err != nil {
			app.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			app.Fatal(err, "failed to create synthient client")
		}
		config.ApplyToClient(&client)

		if flags.transport != "stdio" {
			timber.FatalMsg("unsupported transport", timber.A("value", flags.transport), timber.A("valid", "stdio"))
		}

		err = run(cmd.Context(), client)
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
			app.Fatal(err, "mcp server exited with an error")
		}
	},
}
