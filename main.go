package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/sammccord/flyctl/pkg/iostreams"

	"github.com/sammccord/flyctl/internal/buildinfo"
	"github.com/sammccord/flyctl/internal/cli"
	"github.com/sammccord/flyctl/internal/sentry"
)

func main() {
	os.Exit(run())
}

func run() (exitCode int) {
	ctx, cancel := newContext()
	defer cancel()

	if !buildinfo.IsDev() {
		defer func() {
			if r := recover(); r != nil {
				sentry.Record(r)

				exitCode = 3
			}
		}()
	}

	exitCode = cli.Run(ctx, iostreams.System(), os.Args[1:]...)

	return
}

func newContext() (context.Context, context.CancelFunc) {
	// NOTE: when signal.Notify is called for os.Interrupt it traps both
	// ^C (Control-C) and ^BREAK (Control-Break) on Windows.

	signals := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		signals = append(signals, syscall.SIGTERM)
	}

	return signal.NotifyContext(context.Background(), signals...)
}
