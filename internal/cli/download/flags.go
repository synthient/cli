package download

var (
	flags struct {
		date        string
		hour        int
		silent      bool
		force       bool
		verify      bool
		noPreflight bool
	}
)

func init() {
	Command.PersistentFlags().
		BoolVarP(&flags.silent, "silent", "s", false, "Do not output when downloading")
	Command.PersistentFlags().
		StringVarP(&flags.date, "date", "d", "latest", "Snapshot date to download (YYYY-MM-DD or 'latest')")
	Command.PersistentFlags().
		IntVar(&flags.hour, "hour", -1, "UTC hour for an hourly snapshot (0-23)")
	Command.PersistentFlags().
		BoolVar(&flags.force, "force", false, "Overwrite an existing file")
	Command.PersistentFlags().
		BoolVar(&flags.verify, "verify", false, "Verify downloaded file against snapshot metadata checksum")
	Command.PersistentFlags().
		BoolVar(&flags.noPreflight, "no-preflight", false, "Skip account scope preflight checks")
}
