package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli"
	"github.com/synthient/cli/internal/cli/account"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/cli/download"
	"github.com/synthient/cli/internal/cli/feeds"
	"github.com/synthient/cli/internal/cli/grpcschema"
	"github.com/synthient/cli/internal/cli/lookup"
	"github.com/synthient/cli/internal/cli/scopes"
	"github.com/synthient/cli/internal/cli/status"
	"github.com/synthient/cli/internal/cli/stream"
	"go.mattglei.ch/timber"
)

func main() {
	timber.DisplayTime(false)
	timber.ShowErrorStack(false)
	fatal := timber.GetLevels().Fatal
	fatal.Message = "ERROR"
	timber.SetFatal(fatal)
	timber.ShowFatalStack(
		strings.Contains(os.Args[0], "go-build") || os.Getenv("SYNTHIENT_DEBUG") == "true",
	) // if binary is being ran with go run

	commands := []*cobra.Command{
		account.Command,
		lookup.Command,
		auth.Command,
		feeds.Command,
		grpcschema.Command,
		scopes.Command,
		status.Command,
		stream.Command,
		download.Command,
	}

	for _, command := range commands {
		cli.Root.AddCommand(command)
	}

	_ = cli.Root.Execute()
}
