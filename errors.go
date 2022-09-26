package filehealth

import "errors"

// ErrFileChanged is returned by some issue handlers when they detect that
// a file changed between the time it was examined and the time that a fix
// was requested.
var ErrFileChanged = errors.New("the file has changed since the file was examined")

// ErrDryRun is reported as the outcome for fixes when an operation was
// created as a dry run.
var ErrDryRun = errors.New("dry run")
