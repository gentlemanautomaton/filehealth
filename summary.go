package filehealth

import (
	"fmt"
	"time"
)

// Summary is a summary of a file system scan.
type Summary struct {
	Scanned    int
	Matched    int
	Issues     int
	Start, End time.Time
}

// String returns a string representation of the summary.
func (s Summary) String() string {
	scanned := pluralize(s.Scanned, "file", "files")
	matched := pluralize(s.Matched, "file", "files")
	return fmt.Sprintf("%s scanned, %s with issues, %d total issues (%s)", scanned, matched, s.Issues, s.End.Sub(s.Start))
}

func pluralize(v int, singular, plural string) string {
	if v == 1 {
		return fmt.Sprintf("%d %s", v, singular)
	}
	return fmt.Sprintf("%d %s", v, plural)
}
