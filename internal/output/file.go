package output

import (
	"encoding/json"
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

func Open(path string) (*os.File, func()) {
	if path == "-" {
		return os.Stdout, func() {}
	}

	file, err := os.Create(path)
	if err != nil {
		timber.Fatal(err, "failed to create output file", timber.A("file", path))
	}

	return file, func() {
		err := file.Close()
		if err != nil {
			timber.Fatal(err, "failed to close output file", timber.A("file", path))
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
