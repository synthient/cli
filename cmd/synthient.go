package main

import (
	"time"

	"github.com/synthient/cli/internal/cli"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/cli/download"
	"github.com/synthient/cli/internal/cli/lookup"
	"github.com/synthient/cli/internal/cli/stream"
	"go.mattglei.ch/timber"
)

func main() {
	timber.Timezone(time.Local)
	timber.TimeFormat("03:04:05")
	timber.ShowErrorStack(false)

	cli.Root.AddCommand(lookup.Command)
	cli.Root.AddCommand(auth.Command)
	cli.Root.AddCommand(stream.Command)
	cli.Root.AddCommand(download.Command)
	_ = cli.Root.Execute()
}
