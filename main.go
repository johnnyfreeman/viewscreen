package main

import (
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
		if err := tui.Run(); err != nil {
			fmt.Fprintf(r.errOutput, "TUI error: %v\n", err)
			r.exitFunc(1)
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

func main() {
	NewRunner().Run()
}
