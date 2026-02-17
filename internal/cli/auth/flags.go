package auth

var (
	flags struct {
		logout bool
	}
)

func init() {
	Command.PersistentFlags().
		BoolVarP(&flags.logout, "logout", "l", false, "Remove credentials from your system's keychain")
}
