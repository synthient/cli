package utils

import (
	"fmt"
	"strings"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d < time.Second {
		ms := d.Milliseconds()
		if ms == 0 && d > 0 {
			ms = 1
		}
		return fmt.Sprintf("%dms", ms)
	}

	var parts []string

	if hours := d / time.Hour; hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
		d %= time.Hour
	}
	if minutes := d / time.Minute; minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dmin", minutes))
		d %= time.Minute
	}

	sec := float64(d) / float64(time.Second)
	parts = append(parts, fmt.Sprintf("%.1fs", sec))

	return strings.Join(parts, " ")
}
