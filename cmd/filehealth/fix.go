package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gentlemanautomaton/filehealth"
)

// FixCmd scans a set of files and fixes them.
type FixCmd struct {
	Paths       []string             `kong:"env='PATHS',name='paths',arg,required,help='Paths to search recursively.'"`
	Include     []filehealth.Pattern `kong:"env='INCLUDE',name='include',help='Include files matching regular expression patterns.'"`
	Exclude     []filehealth.Pattern `kong:"env='EXCLUDE',name='exclude',help='Exclude files matching regular expression patterns.'"`
	ShowSkipped bool                 `kong:"env='SHOW_SKIPPED',name='skipped',help='Report on skipped files.'"`
	ShowHealthy bool                 `kong:"env='SHOW_HEALTHY',name='healthy',help='Report on healthy files.'"`
	Batch       int                  `kong:"env='BATCH',name='batch',help='Maximum number of files to fix at a time.'"`
	DryRun      bool                 `kong:"env='DRYRUN',name='dry',help='Perform a dry run without modifying files.'"`
}

// Scanner returns a file health scanner configured according to the command.
func (cmd FixCmd) Scanner() filehealth.Scanner {
	return filehealth.Scanner{
		Handlers:    buildHandlers(),
		SendSkipped: cmd.ShowSkipped,
		SendHealthy: cmd.ShowHealthy,
		Include:     cmd.Include,
		Exclude:     cmd.Exclude,
	}
}

// Run executes the connect command.
func (cmd FixCmd) Run(ctx context.Context) error {
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

func (cmd FixCmd) runJob(ctx context.Context, path string) error {
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

	// If no batch was specified, just use a really high value
	batch := cmd.Batch
	if batch <= 0 {
		batch = 1 << 30
	}

	// Scan and fix files in batches
	for done := false; !done; {
		prealloc := batch
		if prealloc > 4096 {
			prealloc = 4096
		}

		files := make([]filehealth.File, 0, prealloc)
		unhealthy := 0
		for i := 0; i < batch; i++ {
			if done = !iter.Scan(ctx); done {
				break
			}

			file := iter.File()
			if desc := file.Description(); desc != "" {
				fmt.Println(file.Description())
			} else {
				fmt.Println(file)
			}
			files = append(files, file)

			if len(file.Issues) > 0 {
				unhealthy++
			}
		}

		// Stop looping if we encountered an error or there were no more
		// files in the last batch
		if done {
			if err := iter.Err(); err != nil {
				return err
			}
			if unhealthy == 0 {
				break
			}
		}

		// If healthy or skipped files are being shown, we can go through an
		// entire batch without having something to fix. If so, ask the user
		// whether they want to continue.
		if unhealthy == 0 {
			filesCount := pluralize(len(files), "file", "files")
			confirmed, err := promptEnter(fmt.Sprintf("No issues found in %s. Continue?", filesCount))
			if err != nil {
				return err
			}
			if !confirmed {
				return context.Canceled
			}
			continue
		}

		// Prompt the user for confirmation of the proposed actions
		filesCount := pluralize(unhealthy, "file", "files")
		confirmed, err := promptYesNo(fmt.Sprintf("Proceed with fixes affecting %s?", filesCount))
		if err != nil {
			return err
		}

		if !confirmed {
			continue
		}

		fixFiles(ctx, files, cmd.DryRun)

		if !done {
			// Print a summary after each batch
			fmt.Printf("----%s (%s)----\n", iter.Stats(), iter.Duration())
		}
	}

	// Ensure the iterator gets closed
	iter.Close()

	// Print a final summary
	fmt.Printf("----%s (%s)----\n", iter.Stats(), iter.Duration())

	// Report whether the job was interrupted
	return iter.Err()
}

func fixFiles(ctx context.Context, files []filehealth.File, dry bool) error {
	for f := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		file := &files[f]
		var outcomes []filehealth.Outcome
		if dry {
			outcomes, _ = file.DryRun(ctx)
		} else {
			outcomes, _ = file.Fix(ctx)
		}
		for i, outcome := range outcomes {
			prefix := ""
			if err := outcome.Err(); err != nil {
				if err == filehealth.ErrDryRun {
					prefix = "DRY RUN"
				} else {
					prefix = "FAILED"
				}
			} else {
				prefix = "FIXED"
			}
			fmt.Printf("%s: [%d.%d] %s: \"%s\": %s\n", prefix, file.Index, i, outcome.Issue().Summary(), file.Path, outcome)
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
