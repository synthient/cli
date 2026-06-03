package feeddownload

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/dustin/go-humanize"
	"github.com/synthient/cli/internal/feed"
	"github.com/synthient/cli/internal/output"
	"github.com/synthient/cli/internal/utils"
	"github.com/synthient/go-synthient/v2"
)

type Options struct {
	Client   synthient.Client
	Stream   feed.Stream
	Snapshot string
	Filename string
	Force    bool
	Verify   bool
	Quiet    bool
	Out      *os.File
}

type Result struct {
	Size     int64
	Checksum string
	Duration time.Duration
	Meta     synthient.FeedSnapshotMeta
}

type downloadResult struct {
	Size int64
	Err  error
}

func Run(opts Options) (Result, error) {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	interactive := term.IsTerminal(opts.Out.Fd())
	meta, hasMeta, err := getMeta(opts, interactive && !opts.Quiet)
	if err != nil {
		return Result{}, err
	}

	start := time.Now()
	done := make(chan downloadResult, 1)
	go func() {
		size, err := downloadSnapshot(opts)
		done <- downloadResult{Size: size, Err: err}
	}()

	if opts.Quiet {
		result := <-done
		return finish(opts, meta, result, start)
	}

	if !interactive {
		result := <-done
		finished, err := finish(opts, meta, result, start)
		if err != nil {
			return Result{}, err
		}
		writeResult(opts.Out, output.NewStyles(opts.Out), opts, finished)
		return finished, nil
	}

	renderer := progressRenderer{
		out:       opts.Out,
		styles:    output.NewStyles(opts.Out),
		stream:    opts.Stream,
		snapshot:  opts.Snapshot,
		filename:  opts.Filename,
		totalSize: meta.Size,
		hasMeta:   hasMeta,
		start:     start,
	}
	return renderer.run(opts, meta, done)
}

func getMeta(opts Options, progress bool) (synthient.FeedSnapshotMeta, bool, error) {
	if !opts.Verify && !progress {
		return synthient.FeedSnapshotMeta{}, false, nil
	}
	meta, err := opts.Client.FeedSnapshotMeta(opts.Stream.Name, opts.Snapshot, nil)
	if err != nil {
		if opts.Verify {
			return synthient.FeedSnapshotMeta{}, false, fmt.Errorf("getting feed snapshot metadata: %w", err)
		}
		return synthient.FeedSnapshotMeta{}, false, nil
	}
	return meta, true, nil
}

func downloadSnapshot(opts Options) (int64, error) {
	if !strings.HasSuffix(opts.Filename, ".parquet") {
		return 0, fmt.Errorf("file must have a .parquet extension")
	}
	_, statErr := os.Stat(opts.Filename)
	if statErr == nil && !opts.Force {
		return 0, fmt.Errorf("file exists already")
	}
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return 0, statErr
	}

	date, hour, err := feed.SnapshotDateHour(opts.Snapshot)
	if err != nil {
		return 0, err
	}

	r, err := opts.Client.DownloadFeedSnapshot(opts.Stream.Name, date, hour, "", nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = r.Close() }()

	tmp := opts.Filename + ".part"
	_ = os.Remove(tmp)
	file, err := os.Create(tmp)
	if err != nil {
		return 0, fmt.Errorf("creating file: %w", err)
	}

	size, copyErr := io.Copy(file, r)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return 0, fmt.Errorf("writing snapshot: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return 0, fmt.Errorf("closing snapshot file: %w", closeErr)
	}
	if opts.Force {
		_ = os.Remove(opts.Filename)
	}
	err = os.Rename(tmp, opts.Filename)
	if err != nil {
		_ = os.Remove(tmp)
		return 0, fmt.Errorf("moving snapshot into place: %w", err)
	}
	return size, nil
}

func finish(opts Options, meta synthient.FeedSnapshotMeta, download downloadResult, start time.Time) (Result, error) {
	if download.Err != nil {
		return Result{}, download.Err
	}
	result := Result{
		Size:     download.Size,
		Duration: time.Since(start),
		Meta:     meta,
	}
	if opts.Verify {
		checksum, err := verify(opts.Filename, meta)
		if err != nil {
			return Result{}, err
		}
		result.Checksum = checksum
	}
	return result, nil
}

func verify(filename string, meta synthient.FeedSnapshotMeta) (string, error) {
	checksum, err := feed.FileChecksum(filename)
	if err != nil {
		return "", fmt.Errorf("calculating downloaded file checksum: %w", err)
	}
	if checksum != meta.Checksum {
		return "", fmt.Errorf("checksum verification failed: expected %s, got %s", meta.Checksum, checksum)
	}
	return checksum, nil
}

type progressRenderer struct {
	out       *os.File
	styles    output.Styles
	stream    feed.Stream
	snapshot  string
	filename  string
	totalSize int64
	hasMeta   bool
	start     time.Time
	started   bool
}

func (r *progressRenderer) run(opts Options, meta synthient.FeedSnapshotMeta, done <-chan downloadResult) (Result, error) {
	var (
		downloadSize int64
		prevSize     int64
		prevTime     = time.Now()
		smoothBps    float64
		halfLife     = 6 * time.Second
		frames       = []string{"◒", "◐", "◓", "◑"}
		frameIndex   int
	)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			info, err := os.Stat(opts.Filename + ".part")
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return Result{}, fmt.Errorf("checking download status: %w", err)
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

			if frameIndex > len(frames)-1 {
				frameIndex = 0
			}
			r.render(downloadSize, smoothBps, frames[frameIndex])
			frameIndex++

		case result := <-done:
			r.clear()
			finished, err := finish(opts, meta, result, r.start)
			if err != nil {
				return Result{}, err
			}
			writeResult(r.out, r.styles, opts, finished)
			return finished, nil
		}
	}
}

func writeResult(out *os.File, styles output.Styles, opts Options, result Result) {
	if opts.Verify {
		output.Info(out, styles, fmt.Sprintf("Checksum: %s", result.Checksum))
	}
	output.Done(
		out,
		styles,
		fmt.Sprintf("Downloaded %s (%s, %s)", opts.Filename, humanize.Bytes(uint64(result.Size)), utils.FormatDuration(result.Duration)),
	)
}

func (r *progressRenderer) render(downloadSize int64, smoothBps float64, frame string) {
	if r.started {
		r.clear()
	} else {
		fmt.Fprintln(r.out)
		r.started = true
	}

	var out strings.Builder
	fmt.Fprintf(
		&out,
		"%s  %s %s\n",
		r.styles.Frame.Render("◆"),
		r.styles.Title.Render(fmt.Sprintf(`Downloading "%s"`, r.filename)),
		r.styles.Frame.Render(frame),
	)
	r.row(&out, "│", "Stream", r.styles.Warm.Render(r.stream.Name))
	r.row(&out, "│", "Snapshot", r.styles.Key.Render(r.snapshot))
	r.row(&out, "│", "Progress", fmt.Sprintf("%s %s", r.bar(downloadSize), r.styles.Num.Render(r.percent(downloadSize))))
	r.row(&out, "│", "Size", r.styles.Secondary.Render(r.size(downloadSize)))
	r.row(&out, "│", "Rate", r.styles.Num.Render(fmt.Sprintf("%s/s", humanize.Bytes(uint64(smoothBps)))))
	r.rowFinal(&out, "Elapsed", r.styles.Value.Render(utils.FormatDuration(time.Since(r.start))))
	fmt.Fprint(r.out, out.String())
}

func (r *progressRenderer) row(out *strings.Builder, connector string, label string, value string) {
	fmt.Fprintf(
		out,
		"%s  %s %s\n",
		r.styles.Frame.Render(connector),
		output.PadRight(r.styles.Key.Render(label+":"), 9),
		value,
	)
}

func (r *progressRenderer) rowFinal(out *strings.Builder, label string, value string) {
	fmt.Fprintf(
		out,
		"%s  %s %s",
		r.styles.Frame.Render("└"),
		output.PadRight(r.styles.Key.Render(label+":"), 9),
		value,
	)
}

func (r *progressRenderer) clear() {
	if !r.started {
		return
	}
	for range 7 {
		fmt.Fprint(r.out, "\r\033[2K\033[1A")
	}
	fmt.Fprint(r.out, "\r\033[2K")
}

func (r *progressRenderer) bar(downloadSize int64) string {
	width := 28
	if !r.hasMeta || r.totalSize <= 0 {
		return r.styles.Frame.Render(strings.Repeat("━", width))
	}
	progress := min(1, max(0, float64(downloadSize)/float64(r.totalSize)))
	filled := int(math.Round(progress * float64(width)))
	if downloadSize > 0 && filled == 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	return r.styles.Success.Render(strings.Repeat("█", filled)) + r.styles.Muted.Render(strings.Repeat("░", width-filled))
}

func (r *progressRenderer) percent(downloadSize int64) string {
	if !r.hasMeta || r.totalSize <= 0 {
		return ""
	}
	progress := min(1, max(0, float64(downloadSize)/float64(r.totalSize)))
	return fmt.Sprintf("%5.1f%%", progress*100)
}

func (r *progressRenderer) size(downloadSize int64) string {
	if !r.hasMeta || r.totalSize <= 0 {
		return humanize.Bytes(uint64(downloadSize))
	}
	return fmt.Sprintf("%s / %s", humanize.Bytes(uint64(downloadSize)), humanize.Bytes(uint64(r.totalSize)))
}
