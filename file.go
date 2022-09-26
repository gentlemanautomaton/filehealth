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
	Root    Dir
	Path    string
	Index   int
	Skipped bool

	// FileInfo values collected during a scan (may be empty)
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	Issues  []Issue
}

// String returns a string representation of f, including its index and path.
func (f File) String() string {
	return fmt.Sprintf("[%d]: \"%s\"", f.Index, f.Path)
}

// Description returns a multiline string of the file's issues. It returns an
// empty string if the file has no issue.
func (f File) Description() string {
	var out strings.Builder
	for i, issue := range f.Issues {
		if i > 0 {
			out.WriteByte('\n')
		}
		suffix := ""
		if desc := issue.Description(); desc != "" {
			suffix += fmt.Sprintf(": %s", desc)
		}
		if r := issue.Resolution(); r != "" {
			suffix += fmt.Sprintf(": (fix: %s)", r)
		}
		out.WriteString(fmt.Sprintf("[%d.%d] %s: \"%s\"%s", f.Index, i, issue.Summary(), f.Path, suffix))
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

// DryOperation executes an operation for the file as a dry run.
func (f File) DryOperation(fn OperationFunc) error {
	op := Operation{
		scanned: f,
		dry:     true,
	}
	defer op.Close()
	return fn(&op)
}

// DryRun performs a dry run of attempted fixes for each of the file's issues.
func (f File) DryRun(ctx context.Context) ([]Outcome, error) {
	if len(f.Issues) == 0 {
		return nil, nil
	}

	var results []Outcome
	err := f.DryOperation(func(op *Operation) error {
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

// Fix attemtps to fix each of the issues with the given file.
func (f File) Fix(ctx context.Context) ([]Outcome, error) {
	if len(f.Issues) == 0 {
		return nil, nil
	}

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
