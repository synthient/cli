package main

import (
	"time"

	"github.com/synthient/cli/internal/cli"
	"go.mattglei.ch/timber"
)

func main() {
	timber.Timezone(time.Local)
	timber.TimeFormat("03:04:05")
	cli.Execute()
}
