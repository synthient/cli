package utils

import (
	"io"
	"os"

	"github.com/charmbracelet/x/term"
	"go.mattglei.ch/timber"
)

// ReadPipedInput reads all data from standard input when stdin is being piped,
// and returns it as a UTF-8 string.
//
// If stdin is attached to an interactive terminal (TTY), there is no piped input,
// so ReadPipedInput returns an empty string.
//
// This is useful for CLIs that support both:
//
//	echo "hello" | cli
//
// and normal interactive usage without blocking on stdin.
//
// On read failure, this function terminates the process via timber.Fatal.
func ReadPipedInput() string {
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
