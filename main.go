package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/parser"
	"github.com/johnnyfreeman/viewscreen/tui"
	"golang.org/x/term"
)

// Runner encapsulates application dependencies for testability
type Runner struct {
	errOutput     io.Writer
	parserFactory func() *parser.Parser
	exitFunc      func(int)
	parseFlags    func()
}

// RunnerOption is a functional option for configuring a Runner
type RunnerOption func(*Runner)

// WithErrOutput sets a custom error output writer
func WithErrOutput(w io.Writer) RunnerOption {
	return func(r *Runner) {
		r.errOutput = w
	}
}

// WithParserFactory sets a custom parser factory
func WithParserFactory(f func() *parser.Parser) RunnerOption {
	return func(r *Runner) {
		r.parserFactory = f
	}
}

// WithExitFunc sets a custom exit function (for testing)
func WithExitFunc(f func(int)) RunnerOption {
	return func(r *Runner) {
		r.exitFunc = f
	}
}

// WithParseFlags sets a custom flag parsing function
func WithParseFlags(f func()) RunnerOption {
	return func(r *Runner) {
		r.parseFlags = f
	}
}

// NewRunner creates a new Runner with default options
func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		errOutput:     os.Stderr,
		parserFactory: parser.NewParser,
		exitFunc:      os.Exit,
		parseFlags:    config.ParseFlags,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run executes the application
func (r *Runner) Run() {
	r.parseFlags()

	// Determine if we should use TUI mode
	// TUI mode is used when:
	// 1. --no-tui flag is NOT set, AND
	// 2. stdout is a TTY (interactive terminal)
	useTUI := !config.NoTUI && term.IsTerminal(int(os.Stdout.Fd()))

	if useTUI {
		// Auto-enable --auto-exit when stdin is piped (loop-friendly default).
		// When stdin is a pipe, the user is running something like:
		//   while :; do cat ... | claude ... | viewscreen; done
		// In this case, auto-exit prevents the TUI from blocking after each iteration.
		if !term.IsTerminal(int(os.Stdin.Fd())) && !config.AutoExit {
			config.AutoExit = true
			// Auto-enable dump too (same logic as the flag)
			if !isFlagExplicitlySet("dump") {
				config.Dump = true
			}
		}

		content, err := tui.Run()
		if err != nil {
			fmt.Fprintf(r.errOutput, "TUI error: %v\n", err)
			r.exitFunc(1)
		}
		if config.Dump && content != "" {
			fmt.Print(content)
		}
		return
	}

	// Legacy mode: stream directly to stdout
	p := r.parserFactory()
	if err := p.Run(); err != nil {
		fmt.Fprintf(r.errOutput, "%v\n", err)
		r.exitFunc(1)
	}
}

// isFlagExplicitlySet checks if a flag was explicitly set on the command line.
func isFlagExplicitlySet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	NewRunner().Run()
}
