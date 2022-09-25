package filehealth

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"time"

	"github.com/gentlemanautomaton/volmgmt/fileapi"
)

const timeFormat = "2006-01-02 15:04:05 MST"

// TimeHandler handles file timestamp issues.
type TimeHandler struct {
	// Min is the minimum timestamp permitted. Optional.
	Min time.Time

	// Max is the maximum timestamp permitted. Optional.
	Max time.Time

	// Reference is used to compensate for the passage of time during
	// long-running operations. Optional.
	//
	// When specified, the current time will be compared to the reference
	// time, and the difference will be added to Min and Max.
	Reference time.Time

	// Lenience is used to compensate for inaccurate and unsynchronized.
	// clocks. Optional.
	//
	// Timestamps that are close to Min or Max will be accepted if the
	// delta is less than lenience.
	Lenience time.Duration
}

// Name returns the name of the handler.
func (h TimeHandler) Name() string {
	return "File Timestamp Issue Handler"
}

// Examine checks the file under examination for issues. It returns nil if no
// issues are identified.
func (h TimeHandler) Examine(ctx context.Context, exam *Examination) []Issue {
	info := exam.FileInfo()
	if info == nil {
		return nil
	}

	// Attempt to read the windows file time attributes directly
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)

	// Fall back to the modification time only, if necessary
	if !ok {
		if mt := info.ModTime(); h.afterMax(mt) || h.beforeMin(mt) {
			return []Issue{TimeIssue{
				Type:        FileTimeLastWrite,
				Time:        mt,
				TimeHandler: h,
			}}
		}
		return nil
	}

	// Build a list of file timestamp issues
	var issues []Issue

	// Creation
	if ct := filetimeToTime(data.CreationTime); !ct.IsZero() && (h.afterMax(ct) || h.beforeMin(ct)) {
		issues = append(issues, TimeIssue{
			Type:        FileTimeCreation,
			Time:        ct,
			TimeHandler: h,
		})
	}

	// Access
	if at := filetimeToTime(data.LastAccessTime); !at.IsZero() && (h.afterMax(at) || h.beforeMin(at)) {
		issues = append(issues, TimeIssue{
			Type:        FileTimeAccess,
			Time:        at,
			TimeHandler: h,
		})
	}

	// LastWrite
	if wt := filetimeToTime(data.LastWriteTime); !wt.IsZero() && (h.afterMax(wt) || h.beforeMin(wt)) {
		issues = append(issues, TimeIssue{
			Type:        FileTimeLastWrite,
			Time:        wt,
			TimeHandler: h,
		})
	}

	// NOTE: The last change time is not provided by the
	//       Win32FileAttributeData structure, sadly.

	// For the difference between last write and change times, see the
	// article by Raymond Chen, titled "What's the difference between
	// LastWriteTime and ChangeTime in FILE_BASIC_INFO?"
	//
	// https://devblogs.microsoft.com/oldnewthing/20100709-00/?p=13463#:~:text=The%20LastWriteTime%20covers%20writes%20to,.)%20or%20renaming%20the%20file.

	return issues
}

// adjustedMax returns the max time plus elapsed time since reference.
func (h TimeHandler) adjustedMax() time.Time {
	if h.Max.IsZero() || h.Reference.IsZero() {
		return h.Max
	}
	return h.Max.Add(time.Since(h.Reference))
}

// adjustedMin returns the min time plus elapsed time since reference.
func (h TimeHandler) adjustedMin() time.Time {
	if h.Min.IsZero() || h.Reference.IsZero() {
		return h.Min
	}
	return h.Min.Add(time.Since(h.Reference))
}

// afterMax return true if t is after the adjusted max time.
func (h TimeHandler) afterMax(t time.Time) bool {
	if h.Max.IsZero() {
		return false
	}
	if h.Reference.IsZero() {
		return h.Max.Add(h.Lenience).Before(t)
	}
	return h.Max.Add(time.Since(h.Reference) + h.Lenience).Before(t)
}

// beforeMin return true if t is before the adjusted min time.
func (h TimeHandler) beforeMin(t time.Time) bool {
	if h.Min.IsZero() {
		return false
	}
	if h.Reference.IsZero() {
		return h.Min.Add(-h.Lenience).After(t)
	}
	return h.Min.Add(time.Since(h.Reference) - h.Lenience).After(t)
}

// NewTime returns the given time, constrained to the bounds of Min and Max.
func (h TimeHandler) NewTime(t time.Time) time.Time {
	if h.afterMax(t) {
		return h.adjustedMax()
	}
	if h.beforeMin(t) {
		return h.adjustedMin()
	}
	return t
}

// FileTimeType identifies a type of file timestamp.
type FileTimeType int

// Timestamp types.
const (
	FileTimeCreation FileTimeType = iota
	FileTimeAccess
	FileTimeLastWrite
	FileTimeChange
)

// String returns a string representation of the timestamp type.
func (t FileTimeType) String() string {
	switch t {
	case FileTimeCreation:
		return "creation time"
	case FileTimeAccess:
		return "access time"
	case FileTimeLastWrite:
		return "mod time"
	case FileTimeChange:
		return "change time"
	default:
		return fmt.Sprintf("unknown time field %d", t)
	}
}

// TimeIssue describes a file modification time issue.
type TimeIssue struct {
	Type FileTimeType
	Time time.Time

	TimeHandler
}

// Handler returns the Handler that's responsible for handling the name issue.
func (issue TimeIssue) Handler() IssueHandler {
	return issue.TimeHandler
}

// Resolution returns a string describing a proposed resolution to the issue.
func (issue TimeIssue) Description() string {
	return issue.Type.String()
}

// Resolution returns a string describing a proposed resolution to the issue.
func (issue TimeIssue) Resolution() string {
	proposed := issue.NewTime(issue.Time)
	if proposed.Equal(issue.Time) {
		return ""
	}
	return fmt.Sprintf("%s → %s", issue.Time.Format(timeFormat), proposed.Format(timeFormat))
}

// FileOpenFlags returns the set of file permission flags required to fix
// the issue.
func (issue TimeIssue) FileOpenFlags() int {
	return os.O_RDWR
}

// Fix attempts to correct the issue a file.
func (issue TimeIssue) Fix(ctx context.Context, op *Operation) Outcome {
	result := TimeOutcome{
		issue: issue,
	}
	result.err = op.WithFile(func(f fs.File) error {
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

		// Try to get the current values
		var current fileapi.BasicInfo
		if err := fileapi.GetFileInformationByHandleEx(syscall.Handle(file.Fd()), &current); err != nil {
			return err
		}

		// Update the field with an issue
		var update fileapi.BasicInfo

		switch issue.Type {
		case FileTimeCreation:
			update.CreationTime = issue.NewTime(current.CreationTime)
			result.OldTime, result.NewTime = current.CreationTime, update.CreationTime
		case FileTimeAccess:
			update.LastAccessTime = issue.NewTime(current.LastAccessTime)
			result.OldTime, result.NewTime = current.LastAccessTime, update.LastAccessTime
		case FileTimeLastWrite, FileTimeChange:
			update.LastWriteTime = issue.NewTime(current.LastWriteTime)
			update.ChangeTime = issue.NewTime(current.ChangeTime)
			result.OldTime, result.NewTime = current.LastWriteTime, update.LastWriteTime
		}

		return fileapi.SetFileInformationByHandle(syscall.Handle(file.Fd()), update)
	})
	return result
}

// TimeOutcome records the outcome of an attempted fix for a file name issue.
type TimeOutcome struct {
	OldTime time.Time
	NewTime time.Time

	issue TimeIssue
	err   error
}

// Issue returns the issue this outcome pertains to.
func (outcome TimeOutcome) Issue() Issue {
	return outcome.issue
}

// String returns a string representation of the issue.
func (outcome TimeOutcome) String() string {
	resolution := fmt.Sprintf("%s: %s → %s", outcome.issue.Type, outcome.OldTime.Format(timeFormat), outcome.NewTime.Format(timeFormat))
	if outcome.err != nil {
		resolution += ": " + outcome.err.Error()
	}
	return resolution
}

// Err returns an error if one was encountered during the operation.
func (outcome TimeOutcome) Err() error {
	return outcome.err
}

func filetimeToTime(ft syscall.Filetime) time.Time {
	if ft.LowDateTime == 0 && ft.HighDateTime == 0 {
		return time.Time{}
	}
	return time.Unix(0, ft.Nanoseconds())
}
