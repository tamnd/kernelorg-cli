// Command kernel is a CLI for browsing Linux kernel releases from kernel.org.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/tamnd/kernelorg-cli/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.Root()
	err := fang.Execute(ctx, root,
		fang.WithVersion(cli.Version),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	)
	if err == nil {
		return
	}

	var ee *cli.ExitError
	if errors.As(err, &ee) {
		if ee.Err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "kernel:", ee.Err)
		}
		os.Exit(ee.Code)
	}
	_, _ = fmt.Fprintln(os.Stderr, "kernel:", err)
	os.Exit(1)
}
