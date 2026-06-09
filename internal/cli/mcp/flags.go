package mcp

import (
	"fmt"
	"strings"
)

var (
	transports = []string{"stdio"}

	flags struct {
		transport string
	}
)

func init() {
	Command.Flags().
		StringVar(&flags.transport, "transport", "stdio", fmt.Sprintf("Transport the MCP server listens on [%s]", strings.Join(transports, "|")))
}
