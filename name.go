package filehealth

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// NameHandler handles file name issues.
type NameHandler struct {
	TrimSpace bool

	// TODO: Add support for invalid character handling
}

// Name returns the name of the handler.
func (h NameHandler) Name() string {
	return "File Name Issue Handler"
}

// Examine checks the file under examination for issues. It returns nil if no
// issues are identified.
func (h NameHandler) Examine(ctx context.Context, exam *Examination) []Issue {
	info := exam.FileInfo()
	if info == nil {
		return nil
	}

	originalName := info.Name()

	newName := h.clean(originalName)
	if originalName == newName {
		return nil
	}

	return []Issue{
		NameIssue{
			OriginalName: originalName,
			NewName:      newName,
		},
	}
}

func (h NameHandler) clean(name string) string {
	return strings.TrimSpace(name)
}

// NameIssue describes a file name issue.
type NameIssue struct {
	OriginalName string
	NewName      string

	NameHandler
}

// Handler returns the Handler that's responsible for handling the name issue.
func (issue NameIssue) Handler() IssueHandler {
	return issue.NameHandler
}

// Resolution returns a string describing a proposed resolution to the issue.
func (issue NameIssue) Description() string {
	return "leading or trailing space"
}

// Resolution returns a string describing a proposed resolution to the issue.
func (issue NameIssue) Resolution() string {
	return fmt.Sprintf("\"%s\" → \"%s\"", issue.OriginalName, issue.NewName)
}

// FileOpenFlags returns the set of file permission flags required to fix
// the issue.
func (issue NameIssue) FileOpenFlags() int {
	// https://stackoverflow.com/questions/6007463/which-permissions-are-needed-to-delete-a-file-in-windows
	// https://learn.microsoft.com/en-us/windows/win32/fileio/file-security-and-access-rights
	// https://learn.microsoft.com/en-us/windows/win32/secauthz/standard-access-rights
	// https://learn.microsoft.com/en-us/windows/win32/secauthz/access-mask

	//return os.O_RDWR | windows.DELETE
	//return os.O_RDWR
	return 0
}

// Fix attempts to correct the issue a file.
func (issue NameIssue) Fix(ctx context.Context, op *Operation) Outcome {
	outcome := NameOutcome{
		issue: issue,
	}
	outcome.err = op.WithFile(func(f fs.File) error {
		// Ensure the file hasn't changed since it was scanned
		if changed, err := op.FileChanged(); err != nil {
			return err
		} else if changed {
			return ErrFileChanged
		}

		// From (absolute)
		from, err := filepath.Abs(path.Join(string(op.Root()), op.OriginalPath()))
		if err != nil {
			return err
		}
		outcome.OldFilePath = from

		originalDir, _ := path.Split(op.OriginalPath())

		// To (absolute)
		to, err := filepath.Abs(path.Join(string(op.Root()), originalDir, issue.NewName))
		if err != nil {
			return err
		}
		outcome.NewFilePath = to

		// Make sure that a file with that name doesn't already exist
		if _, err := os.Stat(to); err == nil {
			return os.ErrExist
		} else if !os.IsNotExist(err) {
			return err
		}

		// Close open file handles so they don't interfere with the move
		op.Close()

		// Rename the file
		return os.Rename(from, to)

		//return nil

		/*
			// Ensure that it's an operating system file
			file, ok := f.(*os.File)
			if !ok {
				return errors.New("file is not an operating system file")
			}

			//file2FD, err := fileapi.ReOpenFile(syscall.Handle(file.Fd()), windows.DELETE, syscall.FILE_SHARE_DELETE, 0)

			// Update the file name
			update := fileapi.RenameInfo{
				ReplaceIfExists: true,
				FileName:        issue.NewName,
			}

			// https://stackoverflow.com/questions/36450222/moving-a-file-using-setfileinformationbyhandle
			// https://stackoverflow.com/questions/36217150/deleting-a-file-based-on-disk-id
			// https://www.youtube.com/watch?v=uhRWMGBjlO8

			return fileapi.SetFileInformationByHandle(syscall.Handle(file.Fd()), update)
		*/
	})
	return outcome
}

// NameOutcome records the outcome of an attempted fix for a file name issue.
type NameOutcome struct {
	OldFilePath string
	NewFilePath string

	issue NameIssue
	err   error
}

// Issue returns the issue this outcome pertains to.
func (n NameOutcome) Issue() Issue {
	return n.issue
}

// String returns a string representation of the issue.
func (n NameOutcome) String() string {
	resolution := fmt.Sprintf("name change: \"%s\" → \"%s\"", n.OldFilePath, n.NewFilePath)
	if n.err != nil {
		resolution += ": " + n.err.Error()
	}
	return resolution
}

// Err returns an error if one was encountered during the operation.
func (n NameOutcome) Err() error {
	return n.err
}
