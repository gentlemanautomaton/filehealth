package filehealth

import (
	"context"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

// File describes a file that has been scanned.
type File struct {
	// Scanned file location
	Root  Dir
	Path  string
	Index int

	// FileInfo values collected during a scan (may be empty)
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	Issues  []Issue
}

// Description returns a multiline string of the file's issues.
func (f File) Description() string {
	var out strings.Builder
	for i, issue := range f.Issues {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(fmt.Sprintf("[%d.%d]: \"%s\": %s", f.Index, i, f.Path, issue.Description()))
		if r := issue.Resolution(); r != "" {
			out.WriteString(fmt.Sprintf(" (fix: %s)", r))
		}
	}
	return out.String()
}

// Operation executes an operation for the file.
func (f File) Operation(fn OperationFunc) error {
	op := Operation{
		scanned: f,
	}
	defer op.Close()
	return fn(&op)
}

// Fix attemtps to fix each of the issues with the given file.
func (f File) Fix(ctx context.Context) ([]Outcome, error) {
	var results []Outcome
	err := f.Operation(func(op *Operation) error {
		for _, issue := range f.Issues {
			if err := ctx.Err(); err != nil {
				return err
			}
			results = append(results, issue.Fix(ctx, op))
		}
		return nil
	})
	return results, err
}
