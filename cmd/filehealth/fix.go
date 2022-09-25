package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentlemanautomaton/filehealth"
)

// FixCmd scans a set of files and fixes them.
type FixCmd struct {
	Paths []string `kong:"env='PATHS',name='paths',arg,required,help='Paths to search recursively.'"`
	Batch int      `kong:"env='BATCH',name='batch',help='Maximum number of files to fix at a time.'"`
}

// Run executes the connect command.
func (cmd FixCmd) Run(ctx context.Context) error {
	handlers := buildHandlers()

	for _, path := range cmd.Paths {
		root := filehealth.Dir(filepath.Clean(path))

		// Scan the directory
		files, _, err := collectFiles(ctx, handlers, root, cmd.Batch)
		if err != nil {
			return err
		}

		// Continue on if there's no work to be done
		if len(files) == 0 {
			continue
		}

		// Prompt the user for confirmation of the proposed actions
		filesCount := pluralize(len(files), "file", "files")
		confirmed, err := prompt(fmt.Sprintf("Proceed with fixes affecting %s?", filesCount))
		if err != nil {
			fmt.Printf("Cancelling due to unexpected response: %v\n", err)
			os.Exit(1)
		}

		if !confirmed {
			fmt.Printf("Cancelled.\n")
			return nil
		}

		if err := fixFiles(ctx, files); err != nil {
			return err
		}
	}

	return nil
}

func fixFiles(ctx context.Context, files []filehealth.File) error {
	for f := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		file := &files[f]
		outcomes, err := file.Fix(ctx)
		if err != nil {
			fmt.Printf("FAILED: %s: %v\n", file.Path, err)
		}
		for _, outcome := range outcomes {
			if outcome.Err() != nil {
				fmt.Printf("FAILED: %s: %s\n", file.Path, outcome)
			} else {
				fmt.Printf("FIXED: %s: %s\n", file.Path, outcome)
			}
		}
	}
	return nil
}

func pluralize(v int, singular, plural string) string {
	if v == 1 {
		return fmt.Sprintf("%d %s", v, singular)
	}
	return fmt.Sprintf("%d %s", v, plural)
}
