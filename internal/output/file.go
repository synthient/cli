package output

import (
	"fmt"
	"os"

	"go.mattglei.ch/timber"
)

func WriteLine(out *os.File, v ...any) {
	_, err := fmt.Fprintln(out, v...)
	if err != nil {
		timber.Fatal(err, "failed to write string to output")
	}
}
