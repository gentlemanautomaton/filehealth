package filehealth

import (
	"io/fs"
	"os"
	"path"
	"runtime"
	"strings"
)

// Dir is a file directory path accessible via operating system API acalls.
type Dir string

// Open opens the named file.
func (dir Dir) Open(name string) (fs.File, error) {
	return os.DirFS(string(dir)).Open(name)
}

// Open opens the named file with the given flags and mode.
func (dir Dir) OpenFile(name string, flag int, mode fs.FileMode) (fs.File, error) {
	if !fs.ValidPath(name) || runtime.GOOS == "windows" && strings.ContainsAny(name, `\:`) {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrInvalid}
	}

	/*
		// Special handling when DELETE permissions have been requested for
		// deletion and file rename support. The Go standard library doesn't
		// support this flag.
		if flag&windows.DELETE != 0 {
			f, err := fileapi.OpenFile(string(dir)+"/"+name, flag, mode)
			if err != nil {
				return nil, err // nil fs.File
			}
			return f, nil
		}
	*/

	f, err := os.OpenFile(string(dir)+"/"+name, flag, mode)
	if err != nil {
		return nil, err // nil fs.File
	}

	return f, nil
}

// Stat returns a FileInfo describing the file.
func (dir Dir) Stat(name string) (fs.FileInfo, error) {
	return os.DirFS(string(dir)).(fs.StatFS).Stat(name)
}

// FilePath returns the full path of the given file name by joining it
// with dir.
func (dir Dir) FilePath(name string) string {
	return path.Join(string(dir), name)
}
