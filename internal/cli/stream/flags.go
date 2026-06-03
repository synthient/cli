package stream

import "time"

var flags struct {
	reconnect   bool
	output      string
	pretty      bool
	filters     []string
	maxEvents   int
	duration    time.Duration
	noPreflight bool
}

func init() {
	Command.PersistentFlags().
		BoolVar(&flags.reconnect, "reconnect", false, "Reconnect when the server closes a stream")
	Command.PersistentFlags().
		StringVarP(&flags.output, "output", "o", "-", "Where to write output: '-' for stdout, or a file path")
	Command.PersistentFlags().
		BoolVar(&flags.pretty, "pretty", false, "Render events as pretty JSON")
	Command.PersistentFlags().
		StringArrayVar(&flags.filters, "filter", nil, "Filter events with field=value, repeatable")
	Command.PersistentFlags().
		IntVar(&flags.maxEvents, "max-events", 0, "Stop after receiving this many matching events")
	Command.PersistentFlags().
		DurationVar(&flags.duration, "duration", 0, "Stop after this duration, for example 5s or 1m")
	Command.PersistentFlags().
		BoolVar(&flags.noPreflight, "no-preflight", false, "Skip account scope preflight checks")
}
