// Package update implements a non-blocking notifier that lets the user know
// when a newer release of the CLI is available on GitHub. The latest version is
// fetched at most once per day and cached alongside the config so normal
// commands never pay a network cost.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/synthient/cli/internal/conf"
	"github.com/synthient/cli/internal/options"
	"github.com/synthient/cli/internal/output"
)

const (
	releasesURL   = "https://api.github.com/repos/synthient/cli/releases/latest"
	checkInterval = 24 * time.Hour
	cacheFileName = "update_check.json"
	fetchTimeout  = 3 * time.Second
	noticeWait    = 400 * time.Millisecond
)

type cache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

// StartBackgroundCheck refreshes the cached latest version in the background if
// it is stale. It returns a channel that is closed once the check completes so
// the caller can briefly wait for an up-to-date result before exiting.
func StartBackgroundCheck() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if !enabled() {
			return
		}
		c, err := readCache()
		if err == nil && time.Since(c.CheckedAt) < checkInterval {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		latest, err := fetchLatest(ctx)
		if err != nil {
			return // network issues should never disrupt the command
		}
		_ = writeCache(cache{CheckedAt: time.Now(), LatestVersion: latest})
	}()
	return done
}

// Notify prints an upgrade notice to stderr if the cached latest version is
// newer than current. It waits briefly for the background check to finish so a
// freshly fetched version can surface on the same run.
func Notify(current string, done <-chan struct{}) {
	if !enabled() {
		return
	}
	select {
	case <-done:
	case <-time.After(noticeWait):
	}

	c, err := readCache()
	if err != nil {
		return
	}
	if compare(c.LatestVersion, current) <= 0 {
		return
	}

	styles := output.NewStyles(os.Stderr)
	output.WriteLine(os.Stderr, "")
	output.WriteLine(os.Stderr, fmt.Sprintf(
		"%s  %s %s %s %s",
		styles.Warning.Render("▲"),
		styles.Value.Render("Update available"),
		styles.Muted.Render(current),
		styles.Muted.Render("→"),
		styles.Accent.Render(c.LatestVersion),
	))
	hint := "Download the latest release from https://github.com/synthient/cli/releases/latest"
	exe, err := os.Executable()
	if err == nil && strings.HasPrefix(exe, "/opt/homebrew/") {
		hint = "Run `brew upgrade synthient` to update."
	}
	output.WriteLine(os.Stderr, fmt.Sprintf("   %s", styles.Muted.Render(hint)))
}

// enabled reports whether the notifier should run. It stays out of the way of
// scripts, piped output, quiet mode, and local development builds.
func enabled() bool {
	if options.Quiet {
		return false
	}
	if os.Getenv("SYNTHIENT_NO_UPDATE_CHECK") != "" {
		return false
	}
	if strings.Contains(os.Args[0], "go-build") {
		return false // running via `go run`
	}
	return term.IsTerminal(os.Stderr.Fd())
}

func fetchLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("User-Agent", "synthient-cli")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}
	return body.TagName, nil
}

func cachePath() string {
	return filepath.Join(filepath.Dir(conf.Path()), cacheFileName)
}

func readCache() (cache, error) {
	bin, err := os.ReadFile(cachePath())
	if err != nil {
		return cache{}, fmt.Errorf("reading cache: %w", err)
	}
	var c cache
	err = json.Unmarshal(bin, &c)
	if err != nil {
		return cache{}, fmt.Errorf("parsing cache: %w", err)
	}
	return c, nil
}

func writeCache(c cache) error {
	bin, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}
	path := cachePath()
	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}
	err = os.WriteFile(path, bin, 0o644)
	if err != nil {
		return fmt.Errorf("writing cache: %w", err)
	}
	return nil
}

// compare returns 1 if a is a newer version than b, -1 if older, and 0 if
// equal. Pre-release and build suffixes are ignored.
func compare(a, b string) int {
	as := parseSemver(a)
	bs := parseSemver(b)
	for i := range as {
		if as[i] == bs[i] {
			continue
		}
		if as[i] > bs[i] {
			return 1
		}
		return -1
	}
	return 0
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	cut := strings.IndexAny(v, "-+")
	if cut != -1 {
		v = v[:cut]
	}

	var out [3]int
	parts := strings.Split(v, ".")
	for i := 0; i < len(out) && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return [3]int{}
		}
		out[i] = n
	}
	return out
}
