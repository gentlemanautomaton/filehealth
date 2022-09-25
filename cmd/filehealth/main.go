package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
)

func main() {
	// Capture shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var cli struct {
		Scan ScanCmd `kong:"cmd,help='Scans a set of file paths resursively for issues.'"`
		Fix  FixCmd  `kong:"cmd,help='Scans and optionally fixes files with issues.'"`
	}

	app := kong.Parse(&cli,
		kong.Description("Scans the file system for files with health issues and optionally fixes them."),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.UsageOnError())

	if err := app.Run(ctx); err != nil {
		app.FatalIfErrorf(err)
	}
}
