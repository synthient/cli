package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/cli/download"
	"github.com/synthient/cli/internal/cli/lookup"
	"github.com/synthient/cli/internal/cli/stream"
	"go.mattglei.ch/timber"
)

func main() {
	timber.DisplayTime(false)
	timber.ShowErrorStack(false)
	timber.ShowFatalStack(
		strings.Contains(os.Args[0], "go-build") || os.Getenv("SYNTHIENT_DEBUG") == "true",
	) // if binary is being ran with go run

	commands := []*cobra.Command{
		lookup.Command,
		auth.Command,
		stream.Command,
		download.Command,
	}

	for _, command := range commands {
		cli.Root.AddCommand(command)
	}

	_ = cli.Root.Execute()
}
