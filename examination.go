package filehealth

import (
	"io/fs"
)

// ExaminationFunc is a function that runs within the context of an
// examination.
type ExaminationFunc func(*Examination) error

// Operation is an operation on a file that has been scanned.
type Examination struct {
	root  Dir
	path  string
	index int
	info  fs.FileInfo
}

// Path returns the path of the file within its file system.
func (op *Examination) Path() string {
	return op.path
}

// Index returns the index of the file when it was scanned.
func (op *Examination) Index() int {
	return op.index
}

// Index returns the index of the file when it was scanned.
func (op *Examination) FileInfo() fs.FileInfo {
	return op.info
}
