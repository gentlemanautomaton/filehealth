package filehealth

import (
	"io/fs"
	"os"
	"time"
)

// OperationFunc is a function that runs within the context of an operation.
type OperationFunc func(*Operation) error

// FileFunc is a function that can operate on a file.
type FileFunc func(fs.File) error

// Operation is an operation on a file that has been scanned.
type Operation struct {
	scanned File
	dry     bool

	file fs.File

	checkedForChange bool
	changed          bool
	changedErr       error
}

// Root returns the root directory to which the files's paths is
// relative.
func (op *Operation) Root() Dir {
	return op.scanned.Root
}

// Index returns the index of the file when it was examined.
func (op *Operation) Index() int {
	return op.scanned.Index
}

// DryRun returns true if the operation is a dry run.
func (op *Operation) DryRun() bool {
	return op.dry
}

// OriginalName returns the name of the file when it was examined.
func (op *Operation) OriginalName() string {
	return op.scanned.Name
}

// OriginalPath returns the path of the file within its file system.
func (op *Operation) OriginalPath() string {
	return op.scanned.Path
}

// OriginalSize returns the size of the file at the time it was examined.
func (op *Operation) OriginalSize() int64 {
	return op.scanned.Size
}

// OriginalMode returns the mode of the file at the time it was examined.
func (op *Operation) OriginalMode() fs.FileMode {
	return op.scanned.Mode
}

// OriginalModTime returns the modification time of the file at the time it
// was examined.
func (op *Operation) OriginalModTime() time.Time {
	return op.scanned.ModTime
}

// FileChanged reports whether the file's basic attributes were changed
// between the time it was scanned and the first time this function is
// called on the operation.
func (op *Operation) FileChanged() (bool, error) {
	if !op.checkedForChange {
		op.checkedForChange = true

		fi, err := op.fileInfo()
		if err != nil {
			op.changed = true
			op.changedErr = err
		} else {
			op.changed = func() bool {
				if fi.Name() != op.scanned.Name {
					return true
				}
				if fi.Mode() != op.scanned.Mode {
					return true
				}
				if !op.scanned.Mode.IsDir() {
					if fi.Size() != op.scanned.Size {
						return true
					}
				}
				if !fi.ModTime().Equal(op.scanned.ModTime) {
					return true
				}
				return false
			}()
		}
	}

	return op.changed, op.changedErr
}

// WithFile opens the operation's file and invokes the given function on it.
//
// If the file cannot be opened, an error is returned. Otherwise, the result
// of fn() is returned.
//
// If called more than once, the same fs.File will be returned to each call.
// Callers should not use this function to read from or write to the file,
// because the file's position will not be reset between calls.
func (op *Operation) WithFile(fn FileFunc) error {
	if op.file == nil {
		var err error
		if op.file, err = op.open(); err != nil {
			return err
		}
	}

	return fn(op.file)
}

// WithFileExclusive opens the operation's file and invokes the given function
// on it.
//
// If the file cannot be opened, an error is returned. Otherwise, the result
// of fn() is returned.
//
// A unique fs.File will be returned for each call. It is safe to read from and
// write to the file.
func (op *Operation) WithFileExclusive(fn FileFunc) error {
	file, err := op.open()
	if err != nil {
		return err
	}
	defer file.Close()

	return fn(file)
}

func (op *Operation) fileInfo() (fs.FileInfo, error) {
	if op.file == nil {
		return op.scanned.Root.Stat(op.scanned.Path)
	}

	var fi fs.FileInfo
	err := op.WithFile(func(f fs.File) error {
		var err error
		fi, err = f.Stat()
		return err
	})

	return fi, err
}

func (op *Operation) open() (fs.File, error) {
	var flags int
	for i := range op.scanned.Issues {
		flags |= op.scanned.Issues[i].FileOpenFlags()
	}

	var mode fs.FileMode
	if flags == 0 {
		flags = os.O_RDONLY
	} else {
		mode = 0666
	}

	return op.scanned.Root.OpenFile(op.scanned.Path, flags, mode)
}

// Close closes any file handles that the operation may have open.
func (op *Operation) Close() error {
	if op.file == nil {
		return nil
	}
	err := op.file.Close()
	op.file = nil
	return err
}
