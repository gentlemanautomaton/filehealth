package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gentlemanautomaton/filehealth"
)

// ScanCmd scans a set of files without modifying them.
type ScanCmd struct {
	Paths []string `kong:"env='PATHS',name='paths',arg,required,help='Paths to search recursively.'"`
}

// Run executes the connect command.
func (cmd ScanCmd) Run(ctx context.Context) error {
	handlers := buildHandlers()

	for _, path := range cmd.Paths {
		root := filehealth.Dir(filepath.Clean(path))
		if _, _, err := collectFiles(ctx, handlers, root, 0); err != nil {
			return err
		}
	}

	return nil
}

func collectFiles(ctx context.Context, handlers []filehealth.IssueHandler, root filehealth.Dir, batch int) ([]filehealth.File, filehealth.Summary, error) {
	fmt.Printf("----%s----\n", root)
	files, summary, err := filehealth.ScanDir(ctx, root, handlers...)
	if err != nil {
		return files, summary, err
	}
	if batch > 0 && batch < len(files) {
		files = files[:batch]
	}
	for f := range files {
		file := &files[f]
		if desc := file.Description(); desc != "" {
			fmt.Println(file.Description())
		}
	}
	fmt.Printf("----%s----\n", summary)
	return files, summary, nil
}
