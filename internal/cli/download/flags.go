package download

var (
	flags struct {
		date   string
		silent bool
	}
)

func init() {
	Command.PersistentFlags().
		BoolVarP(&flags.silent, "silent", "s", false, "Do not output when downloading")
	Command.PersistentFlags().
		StringVarP(&flags.date, "date", "d", "latest", "Snapshot date to download (YYYY-MM-DD or 'latest')")
}
