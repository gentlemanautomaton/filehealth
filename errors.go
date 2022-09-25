package filehealth

import "errors"

// ErrFileChanged is returned by some issue handlers when they detect that
// a file changed between the time it was examined and the time that a fix
// was requested.
var ErrFileChanged = errors.New("the file has changed since the file was examined")
