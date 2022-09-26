package filehealth

import (
	"context"
	"errors"
	"io"
	"time"
)

type fileIterUpdate struct {
	updated   time.Time
	file      File
	stats     JobStats
	streamErr error
}

// ErrScanCancelled is reported by FileIter.Err() if it's been cancelled.
var ErrScanCancelled = errors.New("the scan has been cancelled")

// FileIter is a file iterator returned by a Scanner. It's used to step
// through the results of a Scanner as they're produced.
type FileIter struct {
	start  time.Time
	ch     <-chan fileIterUpdate
	cancel context.CancelFunc

	end   time.Time
	file  File
	stats JobStats
	err   error
}

// Scan waits for the next file to become available from the iterator.
// It returns false if the context is cancelled, the scanner encounters an
// error, or the end of the stream is reached.
//
// When Scan returns false, check iter.Err() for a non-nil error to
// understand the cause.
func (iter *FileIter) Scan(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		// This scan call has been cancelled, but that doesn't necessarily
		// mean the entire job has been cancelled. We record the context
		// error in the iterator here so this condition isn't confused with
		// job completion, but we don't issue an iter.cancel().
		iter.err = ctx.Err()
		return false
	case update, ok := <-iter.ch:
		// Check for an end-of-stream condition, indicated by channel closure
		if !ok {
			iter.cancel()
			return false
		}

		// Stats are always updated, even if there's a stream error
		iter.stats = update.stats
		iter.end = update.updated

		// Files are only updated when there's no stream error
		if update.streamErr == nil {
			iter.file = update.file
			return true
		}

		// Ignore io.EOF, which is the job's way of telling us it's
		// wrapping up and is sending us completion stats
		if update.streamErr != io.EOF {
			iter.err = update.streamErr
		}

		// Stream errors immediately precede channel closure, so record the
		// error and wrap up
		iter.cancel()

		// Ranging here isn't necessary, but we do so out of caution and as a
		// nice way to make sure the channel has been drained
		for update := range iter.ch {
			iter.stats = update.stats
		}

		return false
	}
}

// File returns the most recently matched file. It is updated each time
// Scan() returns true. Scan() must be called at least once before calling
// this funcion.
func (iter *FileIter) File() File {
	return iter.file
}

// Err returns a non-nil error if the iterator's job encountered an error and
// stopped. It should be called after Scan() returns false. It returns nil
// if the job completed successfully.
func (iter *FileIter) Err() error {
	return iter.err
}

// Stats returns the statistics for the iterator's job.
func (iter *FileIter) Stats() JobStats {
	return iter.stats
}

// Duration returns the duration of the iterator's job.
func (iter *FileIter) Duration() time.Duration {
	return iter.end.Sub(iter.start)
}

// Close causes the iterator's job to stop running. It always returns nil.
func (iter *FileIter) Close() error {
	// Request cancellation
	iter.cancel()

	// Collect stats as we wait for the job to wrap up
	for update := range iter.ch {
		iter.stats = update.stats
		iter.end = update.updated
		if update.streamErr != nil {
			iter.err = update.streamErr
		}
	}

	return nil
}
