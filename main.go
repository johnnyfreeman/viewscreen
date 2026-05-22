package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/agent"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/parser"
	"github.com/johnnyfreeman/viewscreen/tui"
	"golang.org/x/term"
)

// stdinReader is used for reading stdin as prompt text.
// It can be overridden in tests.
var stdinReader io.Reader = os.Stdin

// Runner encapsulates application dependencies for testability
type Runner struct {
	errOutput     io.Writer
	parserFactory func() *parser.Parser
	promptStarter func(string, io.Reader) (promptProcess, error)
	exitFunc      func(int)
	configOpts    []config.Option
}

type promptProcess interface {
	Stdout() io.ReadCloser
	Wait() error
	Kill() error
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

// WithPromptStarter sets a custom prompt process starter.
func WithPromptStarter(f func(string, io.Reader) (promptProcess, error)) RunnerOption {
	return func(r *Runner) {
		r.promptStarter = f
	}
}

// WithExitFunc sets a custom exit function (for testing)
func WithExitFunc(f func(int)) RunnerOption {
	return func(r *Runner) {
		r.exitFunc = f
	}
}

// WithConfigOpts sets custom config parse options (for testing)
func WithConfigOpts(opts ...config.Option) RunnerOption {
	return func(r *Runner) {
		r.configOpts = opts
	}
}

// NewRunner creates a new Runner with default options
func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		errOutput:     os.Stderr,
		parserFactory: parser.NewParser,
		promptStarter: func(prompt string, stdin io.Reader) (promptProcess, error) {
			return agent.Start(config.Get().Agent, prompt, stdin)
		},
		exitFunc:   os.Exit,
		configOpts: []config.Option{config.WithArgs(os.Args[1:])},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run executes the application
func (r *Runner) Run() {
	cfg, err := config.Parse(r.configOpts...)
	if err != nil {
		fmt.Fprintf(r.errOutput, "%v\n", err)
		r.exitFunc(1)
		return
	}

	// Resolve prompt: positional args take priority, then -p reads stdin
	prompt := cfg.Prompt
	if prompt == "" && cfg.PromptMode {
		data, err := io.ReadAll(stdinReader)
		if err == nil {
			prompt = string(data)
		}
	}

	// Determine if we should use TUI mode
	// TUI mode is used when:
	// 1. --no-tui flag is NOT set, AND
	// 2. stdout is a TTY (interactive terminal)
	useTUI := !cfg.NoTUI && term.IsTerminal(int(os.Stdout.Fd()))

	// If we have a prompt, spawn claude and stream its JSON output either into
	// the TUI or directly through the legacy renderer when stdout is not a TTY.
	if prompt != "" {
		if useTUI {
			content, err := tui.RunWithPrompt(prompt)
			if err != nil {
				fmt.Fprintf(r.errOutput, "TUI error: %v\n", err)
				r.exitFunc(1)
			}
			if cfg.Dump && content != "" {
				fmt.Print(content)
			}
			return
		}

		if err := r.runPromptLegacy(prompt); err != nil {
			fmt.Fprintf(r.errOutput, "%v\n", err)
			r.exitFunc(1)
		}
		return
	}

	if useTUI {
		// Auto-enable --auto-exit when stdin is piped (loop-friendly default).
		// When stdin is a pipe, the user is running something like:
		//   while :; do cat ... | claude ... | viewscreen; done
		// In this case, auto-exit prevents the TUI from blocking after each iteration.
		if !term.IsTerminal(int(os.Stdin.Fd())) && !cfg.AutoExit {
			cfg.AutoExit = true
			// Auto-enable dump too (same logic as the flag)
			if !cfg.Dump {
				cfg.Dump = true
			}
		}

		content, err := tui.Run()
		if err != nil {
			fmt.Fprintf(r.errOutput, "TUI error: %v\n", err)
			r.exitFunc(1)
		}
		if cfg.Dump && content != "" {
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

func (r *Runner) runPromptLegacy(prompt string) error {
	proc, err := r.promptStarter(prompt, nil)
	if err != nil {
		return err
	}

	stdout := proc.Stdout()
	if stdout == nil {
		_ = proc.Kill()
		_ = proc.Wait()
		return errors.New("agent stdout unavailable")
	}

	p := parser.NewParserWithOptions(parser.WithInput(stdout))
	parseErr := p.Run()
	waitErr := proc.Wait()
	if parseErr != nil {
		return parseErr
	}
	if waitErr != nil {
		return waitErr
	}
	return nil
}

func main() {
	NewRunner().Run()
}
