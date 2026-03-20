package download

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/synthient/cli/internal/cli/auth"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/utils"
	"go.mattglei.ch/timber"
)

var Command = &cobra.Command{
	Use:   "download",
	Short: "Download an anonymizer feed to a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := conf.Read()
		if err != nil {
			timber.Fatal(err, "failed to read configuration file")
		}

		client, err := auth.SynthientClient(config)
		if err != nil {
			timber.Fatal(err, "failed to create synthient client")
		}

		var (
			done     = make(chan struct{})
			filename = args[0]
			start    = time.Now()
		)

		_, err = os.Stat(filename)
		if !errors.Is(err, fs.ErrNotExist) {
			timber.ErrorMsg("exists already", timber.A("file", filename))
			os.Exit(1)
		}

		go func() {
			defer close(done)
			_, err := client.DownloadAnonymizersFeed(flags.query, filename, nil)
			if err != nil {
				timber.Fatal(err, "failed to download", timber.A("file", filename))
			}
		}()

		if flags.silent {
			<-done
			return
		}

		output.AnonymizerQuery(flags.query)
		fmt.Print("\n\n")

		var (
			downloadSize int64
			prevSize     int64
			prevTime     = time.Now()
			smoothBps    float64
			halfLife     = 6 * time.Second

			spinnerFrames = []string{
				"[=    ]",
				"[==   ]",
				"[ === ]",
				"[  ===]",
				"[   ==]",
				"[    =]",
				"[   ==]",
				"[  ===]",
				"[ === ]",
				"[==   ]",
			}
			spinnerIndex   = 0
			spinnerStarted = false
			spinnerHeight  = 2
		)

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:

				info, err := os.Stat(filename)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						continue
					}
					timber.Fatal(err, "failed to check status of", timber.A("file", filename))
				}
				downloadSize = info.Size()
				now := time.Now()

				dt := now.Sub(prevTime)
				if dt > 0 {
					delta := max(downloadSize-prevSize, 0)
					instBps := float64(delta) / dt.Seconds()
					alpha := 1 - math.Exp(-float64(dt)/float64(halfLife))
					if smoothBps == 0 {
						smoothBps = instBps
					} else {
						smoothBps += alpha * (instBps - smoothBps)
					}
				}
				prevSize = downloadSize
				prevTime = now

				if spinnerStarted {
					clearSpinner(spinnerHeight)
				} else {
					spinnerStarted = true
				}
				if spinnerIndex > len(spinnerFrames)-1 {
					spinnerIndex = 0
				}
				out := strings.Builder{}
				for range spinnerHeight {
					out.WriteString("\r\033[2K\033[1A\r")
				}
				fmt.Fprintf(
					&out,
					"\n%s %s",
					output.StdoutStyles.Bold.Render(fmt.Sprintf(`Downloading "%s"`, filename)),
					output.StdoutStyles.SynthientColor.Render(spinnerFrames[spinnerIndex]),
				)
				fmt.Fprintln(&out)
				fmt.Fprintf(&out, "   Size: %s\n", humanize.Bytes(uint64(downloadSize)))
				fmt.Fprintf(&out, "   Rate: %s/s\n", humanize.Bytes(uint64(smoothBps)))
				fmt.Fprintf(&out, "Elapsed: %s", utils.FormatDuration(time.Since(start)))
				fmt.Fprint(&out, "    ") // add extra padding for cursor
				fmt.Print(out.String())
				spinnerIndex++

			case <-done:
				clearSpinner(spinnerHeight)
				timber.Done(
					"downloaded",
					timber.A("file", filename),
					timber.A("size", humanize.Bytes(uint64(downloadSize))),
					timber.A("elapsed", utils.FormatDuration(time.Since(start))),
				)
				return
			}
		}
	},
}

func clearSpinner(height int) {
	for range height {
		fmt.Print("\r\033[2K\033[1A\r")
	}
}
