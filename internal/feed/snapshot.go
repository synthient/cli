package feed

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SnapshotDateHour parses a snapshot string into a date and optional hour for
// use with the synthient module's download and meta methods.
func SnapshotDateHour(snapshot string) (string, *int, error) {
	snapshot = strings.TrimSpace(snapshot)
	if snapshot == "latest" {
		return "latest", nil, nil
	}

	date, hourText, hasHour := strings.Cut(snapshot, "/")
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", nil, fmt.Errorf("snapshot must be latest, YYYY-MM-DD, or YYYY-MM-DD/HH")
	}
	if !hasHour {
		return date, nil, nil
	}
	if strings.TrimSpace(hourText) == "" {
		return "", nil, fmt.Errorf("snapshot hour is required after /")
	}
	hour, err := strconv.Atoi(hourText)
	if err != nil || hour < 0 || hour > 23 {
		return "", nil, fmt.Errorf("snapshot hour must be between 0 and 23")
	}
	return date, &hour, nil
}
