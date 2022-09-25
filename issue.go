package filehealth

import (
	"context"
)

// IssueHandler handles file issues of a particular type.
type IssueHandler interface {
	// Name returns the name of the handler.
	Name() string

	// Examine checks the file under examination for issues. It returns nil if no
	// issues are identified.
	Examine(context.Context, *Examination) []Issue
}

// Issue describes a problem with a file.
type Issue interface {
	// Handler returns the Handler that's responsible for handling the issue.
	Handler() IssueHandler

	// Description returns a string describing the issue.
	Description() string

	// Resolution returns a string describing the fix. It returns an empty
	// string if no resolution is possible.
	Resolution() string

	// FileOpenFlags returns the set of file permission flags required to fix
	// the issue.
	FileOpenFlags() int

	// Fix attempts to fix the issue.
	Fix(context.Context, *Operation) Outcome
}
