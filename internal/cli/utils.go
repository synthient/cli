package cli

import (
	"io"
	"os"

	"github.com/charmbracelet/x/term"
	"go.mattglei.ch/timber"
)

func PipedInput() string {
	isTTY := term.IsTerminal(os.Stdin.Fd())
	if isTTY {
		return ""
	}
	inputBinary, err := io.ReadAll(os.Stdin)
	if err != nil {
		timber.Fatal(err, "failed to read standard input")
	}
	return string(inputBinary)
}
