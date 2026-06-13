// Package cli builds the kernel command tree on top of the kernelorg library.
package cli

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tamnd/kernelorg-cli/kernelorg"
)

// Build metadata, injected via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// exit codes.
const (
	exitError  = 1
	exitUsage  = 2
	exitNoData = 3
)

// ExitError carries a process exit code up to main.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit %d", e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

func codeError(code int, err error) error { return &ExitError{Code: code, Err: err} }

// App holds shared state threaded through every command.
type App struct {
	cfg    kernelorg.Config
	client *kernelorg.Client

	// global output flags
	format   string
	fields   []string
	noHeader bool
	template string
	limit    int
	quiet    bool
}

// Root builds the root command and its subtree.
func Root() *cobra.Command {
	app := &App{cfg: kernelorg.DefaultConfig()}

	root := &cobra.Command{
		Use:   "kernel",
		Short: "Browse Linux kernel releases from kernel.org",
		Long: `kernel reads Linux kernel release data from the official kernel.org JSON API
and returns rich, structured records as table, JSON, JSONL, CSV, TSV, or URLs.

kernel is an independent tool and is not affiliated with, endorsed by,
or sponsored by The Linux Foundation or kernel.org.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return app.setup()
		},
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&app.format, "format", "f", "", "output: table|json|jsonl|csv|tsv|url|raw (default: table on a TTY, jsonl when piped)")
	pf.StringSliceVar(&app.fields, "fields", nil, "comma-separated columns to include")
	pf.BoolVar(&app.noHeader, "no-header", false, "omit the header row in table/csv/tsv")
	pf.StringVar(&app.template, "template", "", "Go text/template applied per record")
	pf.IntVarP(&app.limit, "limit", "n", 0, "limit number of results (0 = no limit)")
	pf.BoolVarP(&app.quiet, "quiet", "q", false, "suppress progress messages on stderr")

	pf.DurationVar(&app.cfg.Rate, "rate", kernelorg.DefaultConfig().Rate, "minimum spacing between requests")
	pf.DurationVar(&app.cfg.Timeout, "timeout", kernelorg.DefaultConfig().Timeout, "per-request timeout")
	pf.IntVar(&app.cfg.Retries, "retries", kernelorg.DefaultConfig().Retries, "retry attempts on 429/5xx")

	root.AddCommand(
		app.releasesCmd(),
		app.latestCmd(),
		app.releaseCmd(),
		app.longtermCmd(),
		newVersionCmd(),
	)
	return root
}

// setup resolves output format default and constructs the shared client.
func (a *App) setup() error {
	if a.format == "" {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			a.format = string(FormatTable)
		} else {
			a.format = string(FormatJSONL)
		}
	}
	a.client = kernelorg.NewClient(a.cfg)
	return nil
}

// render writes records using the resolved global flags.
func (a *App) render(records any) error {
	r := NewRenderer(os.Stdout, Format(a.format), a.fields, a.noHeader, a.template)
	return r.Render(records)
}

// renderOrEmpty renders records, mapping an empty result to exit code 3.
func (a *App) renderOrEmpty(records any, n int) error {
	if err := a.render(records); err != nil {
		return err
	}
	if n == 0 {
		return codeError(exitNoData, nil)
	}
	return nil
}

// applyLimit trims a slice to at most a.limit items (0 = no limit).
func applyLimit[T any](items []T, limit int) []T {
	if limit <= 0 || limit >= len(items) {
		return items
	}
	return items[:limit]
}

// progressf prints a progress line to stderr unless --quiet.
func (a *App) progressf(format string, args ...any) {
	if a.quiet {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// mapFetchErr converts a library error into the right exit code.
func mapFetchErr(err error) error {
	switch {
	case err == nil:
		return nil
	case isNotFound(err):
		return codeError(exitNoData, err)
	default:
		return codeError(exitError, err)
	}
}
