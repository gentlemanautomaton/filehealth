package filehealth

import (
	"context"
	"errors"
	"os"
)

// scanHandler is used to report issues encountered by the scanner.
type scanHandler struct{}

func (h scanHandler) Name() string {
	return "File Access"
}

func (h scanHandler) Examine(ctx context.Context, exam *Examination) []Issue {
	return nil
}

// ScanIssue describes an issue encounterd by a scanner.
type ScanIssue struct {
	Err error
}

// Handler returns the Handler that's responsible for handling the issue.
func (issue ScanIssue) Handler() IssueHandler {
	return scanHandler{}
}

// Summary returns a short summary of the issue.
func (issue ScanIssue) Summary() string {
	return "access failure"
}

// Description returns a description of the issue. It may return an empty
// string if the information provided by the summary is sufficient.
func (issue ScanIssue) Description() string {
	var pathErr *os.PathError
	if errors.As(issue.Err, &pathErr) {
		return pathErr.Op + ": " + pathErr.Err.Error()
	}
	return issue.Err.Error()
}

// Resolution returns a string describing the fix. It returns an empty
// string if no resolution is possible.
func (issue ScanIssue) Resolution() string {
	return ""
}

// FileOpenFlags returns the set of file permission flags required to fix
// the issue.
func (issue ScanIssue) FileOpenFlags() int {
	return 0
}

// Fix attempts to fix the issue.
func (issue ScanIssue) Fix(ctx context.Context, op *Operation) Outcome {
	return nil
}
