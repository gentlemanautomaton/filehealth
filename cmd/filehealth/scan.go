package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gentlemanautomaton/filehealth"
)

// ScanCmd scans a set of files without modifying them.
type ScanCmd struct {
	Paths       []string             `kong:"env='PATHS',name='paths',arg,required,help='Paths to search recursively.'"`
	Include     []filehealth.Pattern `kong:"env='INCLUDE',name='include',help='Include files matching regular expression pattern.'"`
	Exclude     []filehealth.Pattern `kong:"env='EXCLUDE',name='exclude',help='Exclude files matching regular expression pattern.'"`
	ShowSkipped bool                 `kong:"env='SHOW_SKIPPED',name='skipped',help='Report on skipped files.'"`
	ShowHealthy bool                 `kong:"env='SHOW_HEALTHY',name='healthy',help='Report on healthy files.'"`
}

// Scanner returns a file health scanner configured according to the command.
func (cmd ScanCmd) Scanner() filehealth.Scanner {
	return filehealth.Scanner{
		Handlers:    buildHandlers(),
		SendSkipped: cmd.ShowSkipped,
		SendHealthy: cmd.ShowHealthy,
		Include:     cmd.Include,
		Exclude:     cmd.Exclude,
	}
}

// Run executes the connect command.
func (cmd ScanCmd) Run(ctx context.Context) error {
	// Scan each of the provided paths
	for _, path := range cmd.Paths {
		if err := cmd.runJob(ctx, path); err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return err
		}
	}

	return nil
}

func (cmd ScanCmd) runJob(ctx context.Context, path string) error {
	// Prepare a scanner with the desired configuration
	scanner := cmd.Scanner()

	// Start a job
	root := filehealth.Dir(filepath.Clean(path))
	iter := scanner.ScanDir(root)

	// Print the root directory
	if abs, err := filepath.Abs(string(root)); err != nil {
		fmt.Printf("----%s----\n", root)
	} else {
		fmt.Printf("----%s----\n", abs)
	}

	// Process each scanned file
	for iter.Scan(ctx) {
		file := iter.File()
		if desc := file.Description(); desc != "" {
			fmt.Println(file.Description())
		} else {
			fmt.Println(file)
		}
	}

	// Ensure the iterator gets closed
	iter.Close()

	// Print a summary
	fmt.Printf("----%s (%s)----\n", iter.Stats(), iter.Duration())

	// Report whether the job was interrupted
	return iter.Err()
}
