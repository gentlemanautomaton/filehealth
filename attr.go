package filehealth

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"github.com/gentlemanautomaton/volmgmt/fileapi"
	"github.com/gentlemanautomaton/volmgmt/fileattr"
)

// AttrHandler handles file attribute issues.
type AttrHandler struct {
	Unwanted fileattr.Value
}

// Name returns the name of the handler.
func (h AttrHandler) Name() string {
	return "File Attribute Issue Handler"
}

// Examine checks the file under examination for issues. It returns nil if no
// issues are identified.
func (h AttrHandler) Examine(ctx context.Context, exam *Examination) []Issue {
	info := exam.FileInfo()
	if info == nil {
		return nil
	}

	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return nil
	}

	var (
		original = fileattr.Value(data.FileAttributes)
		matched  fileattr.Value
	)

	for i := 0; i < 32; i++ {
		flag := fileattr.Value(1 << uint32(i))
		if h.Unwanted.Match(flag) && original.Match(flag) {
			matched |= flag
		}
	}

	if matched == 0 {
		return nil
	}

	return []Issue{
		AttrIssue{
			Original: original,
			Matched:  matched,
		},
	}
}

// NameIssue describes a file name issue.
type AttrIssue struct {
	Original fileattr.Value
	Matched  fileattr.Value

	AttrHandler
}

// Handler returns the Handler that's responsible for handling the name issue.
func (issue AttrIssue) Handler() IssueHandler {
	return issue.AttrHandler
}

// Resolution returns a string describing a proposed resolution to the issue.
func (issue AttrIssue) Description() string {
	return fmt.Sprintf("unwanted attributes: %s", issue.Matched.Join(",", fileattr.FormatCode))
}

// Resolution returns a string describing a proposed resolution to the issue.
func (issue AttrIssue) Resolution() string {
	updated := issue.Original &^ issue.Matched
	before := issue.Original.Join(",", fileattr.FormatCode)
	after := updated.Join(",", fileattr.FormatCode)
	return fmt.Sprintf("%s → %s", before, after)

}

// FileOpenFlags returns the set of file permission flags required to fix
// the issue.
func (issue AttrIssue) FileOpenFlags() int {
	return os.O_RDWR
}

// Fix attempts to correct the issue with the file.
func (issue AttrIssue) Fix(ctx context.Context, op *Operation) Outcome {
	outcome := AttrOutcome{
		issue: issue,
	}
	outcome.err = op.WithFile(func(f fs.File) error {
		// Ensure the file hasn't changed since it was scanned
		if changed, err := op.FileChanged(); err != nil {
			return err
		} else if changed {
			return ErrFileChanged
		}

		// Ensure that it's an operating system file
		file, ok := f.(*os.File)
		if !ok {
			return errors.New("file is not an operating system file")
		}

		// Get current file attributes
		info, err := file.Stat()
		if err != nil {
			return err
		}

		data, ok := info.Sys().(*syscall.Win32FileAttributeData)
		if !ok {
			return errors.New("file is not a windows operating system file")
		}

		attrs := fileattr.Value(data.FileAttributes)

		// Update the attributes
		update := fileapi.BasicInfo{
			FileAttributes: attrs &^ issue.Matched,
		}

		outcome.OldAttributes = attrs
		outcome.NewAttributes = update.FileAttributes

		return fileapi.SetFileInformationByHandle(syscall.Handle(file.Fd()), update)
	})
	return outcome
}

// AttrOutcome records the outcome of an attempted fix for a file name issue.
type AttrOutcome struct {
	OldAttributes fileattr.Value
	NewAttributes fileattr.Value

	issue AttrIssue
	err   error
}

// Issue returns the issue this outcome pertains to.
func (outcome AttrOutcome) Issue() Issue {
	return outcome.issue
}

// String returns a string representation of the issue.
func (outcome AttrOutcome) String() string {
	before := outcome.OldAttributes.Join(",", fileattr.FormatCode)
	after := outcome.NewAttributes.Join(",", fileattr.FormatCode)
	resolution := fmt.Sprintf("attribute change: %s → %s", before, after)
	if outcome.err != nil {
		resolution += ": " + outcome.err.Error()
	}
	return resolution
}

// Err returns an error if one was encountered during the operation.
func (outcome AttrOutcome) Err() error {
	return outcome.err
}
