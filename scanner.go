package filehealth

import (
	"context"
	"io/fs"
	"time"
)

// Scanner scans a set of files for issues identified by its issue handlers.
type Scanner struct {
	Handlers []IssueHandler
}

// func (s Scanner) ScanFS(ctx context.Context, fsys fs.FS) ([]File, Summary, error) {
// ScanFS causes the scanner to scan the given file system directory.
func (s Scanner) ScanDir(ctx context.Context, root Dir) ([]File, Summary, error) {
	var files []File

	summary := Summary{Start: time.Now()}

	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		summary.Scanned++

		file := File{
			Root:  root,
			Path:  path,
			Index: summary.Scanned,
		}

		// If an error was reported, such as access denied, record it as a
		// scan error and carry on
		if err != nil {
			file.Issues = append(file.Issues, ScanIssue{Err: err})
			files = append(files, file)
			return nil
		}

		// Attempt to collect more information about the file
		info, err := d.Info()
		if err != nil {
			file.Issues = append(file.Issues, ScanIssue{Err: err})
		} else {
			file.Name = info.Name()
			file.Size = info.Size()
			file.Mode = info.Mode()
			file.ModTime = info.ModTime()
		}

		// Ask each of the handlers to examine the file and return a set
		// of issues
		if len(s.Handlers) > 0 {
			exam := Examination{
				root:  file.Root,
				path:  file.Path,
				index: file.Index,
				info:  info,
			}
			for _, h := range s.Handlers {
				file.Issues = append(file.Issues, h.Examine(ctx, &exam)...)
			}
		}

		// If issues were encountered, add the file to the files list.
		if count := len(file.Issues); count > 0 {
			files = append(files, file)
			summary.Matched++
			summary.Issues += count
		}

		return nil
	})

	summary.End = time.Now()

	return files, summary, err
}

// ScanFS scans the given file system for issues.
func ScanDir(ctx context.Context, root Dir, handlers ...IssueHandler) ([]File, Summary, error) {
	return Scanner{Handlers: handlers}.ScanDir(ctx, root)
}
