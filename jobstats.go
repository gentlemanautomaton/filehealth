package filehealth

import "fmt"

// JobStats report scanning tallies during and at the completion of scanning.
type JobStats struct {
	// Skipped is the number of files not scanned due to filters.
	Skipped int

	// Scanned is the number of files scanned.
	Scanned int

	// Healthy is the number of scanned files that had no issues.
	Healthy int

	// Unhealthy is the number of scanned files that had at least one issue.
	Unhealthy int

	// Issues is the total number of issues detected in scanned files.
	Issues int
}

// String returns a string representation of the job statistics.
func (s JobStats) String() string {
	return fmt.Sprintf("%d skipped, %d scanned, %d healthy, %d unhealthy, %d issues", s.Skipped, s.Scanned, s.Healthy, s.Unhealthy, s.Issues)
}
