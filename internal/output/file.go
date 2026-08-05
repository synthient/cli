package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/colorprofile"
	"go.mattglei.ch/timber"
)

func WriteLine(out *os.File, v ...any) {
	writer := colorprofile.NewWriter(out, os.Environ())
	writer.Profile = detectProfile(out)
	_, err := fmt.Fprintln(writer, v...)
	if err != nil {
		timber.Fatal(err, "failed to write string to output")
	}
}

func Open(path string) (*os.File, func()) {
	if path == "-" {
		return os.Stdout, func() {}
	}

	file, err := os.Create(path)
	if err != nil {
		timber.Fatalf(err, "failed to create output file: %s", path)
	}

	return file, func() {
		err := file.Close()
		if err != nil {
			timber.Fatalf(err, "failed to close output file: %s", path)
		}
	}
}

func JSON(out *os.File, value any) {
	enc := json.NewEncoder(out)
	err := enc.Encode(value)
	if err != nil {
		timber.Fatal(err, "failed to write json output")
	}
}
