package tui

import (
	"errors"
	"io"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	claudepkg "github.com/johnnyfreeman/viewscreen/claude"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"golang.org/x/term"
)

func startClaudeProcess(prompt string) (managedClaudeProcess, error) {
	return claudepkg.Start(prompt, nil)
}

// Run starts the TUI and returns the final rendered content for optional dumping.
func Run() (string, error) {
	// Initialize styles (needed for renderers)
	cfg := config.Get()
	render.NewMarkdownRenderer(cfg.NoColor(), 80)
	resetTerminalModes(os.Stdout)

	var opts []tea.ProgramOption
	width, height := detectTerminalSize(os.Stdout)
	stdinIsTTY := isatty(os.Stdin.Fd())

	// When stdin is not a TTY (e.g., piped input), keyboard input must come
	// from /dev/tty instead of the stream-json pipe.
	if !stdinIsTTY {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			opts = append(opts, tea.WithInput(tty))
			defer tty.Close()
		}
	}

	p := tea.NewProgram(NewModel(
		WithInputReader(streamInputReader(os.Stdin, stdinIsTTY)),
		WithInitialSize(width, height),
		WithAutoExit(cfg.AutoExit),
		WithVerboseParseErrors(cfg.IsVerbose()),
	), opts...)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(Model); ok {
		return m.content.String(), nil
	}
	return "", nil
}

func streamInputReader(stdin io.Reader, stdinIsTTY bool) io.Reader {
	if stdinIsTTY || stdin == nil {
		return strings.NewReader("")
	}
	return stdin
}

// RunWithPrompt spawns claude with the given prompt and runs the TUI on its output.
func RunWithPrompt(prompt string) (string, error) {
	// Initialize styles (needed for renderers)
	cfg := config.Get()
	render.NewMarkdownRenderer(cfg.NoColor(), 80)
	resetTerminalModes(os.Stdout)

	proc, err := startClaudeProcess(prompt)
	if err != nil {
		return "", err
	}
	stdout := proc.Stdout()
	if stdout == nil {
		_ = proc.Kill()
		_ = proc.Wait()
		return "", errors.New("claude stdout unavailable")
	}

	var teaOpts []tea.ProgramOption
	width, height := detectTerminalSize(os.Stdout)
	tty, err := os.Open("/dev/tty")
	if err == nil {
		teaOpts = append(teaOpts, tea.WithInput(tty))
		defer tty.Close()
	}

	model := NewModel(
		WithInputReader(stdout),
		WithClaudeProcess(proc),
		WithClaudeStarter(startClaudeProcess),
		WithPrompt(prompt),
		WithInitialSize(width, height),
		WithAutoExit(cfg.AutoExit),
		WithVerboseParseErrors(cfg.IsVerbose()),
	)

	p := tea.NewProgram(model, teaOpts...)

	finalModel, err := p.Run()
	if err != nil {
		_ = proc.Kill()
		_ = proc.Wait()
		return "", err
	}

	if m, ok := finalModel.(Model); ok {
		// If the user quit before Claude closed stdout, terminate it so the TUI
		// exits immediately instead of waiting for the generation to finish.
		m.stopClaudeProcessIfRunning()
		if m.claudeProcess != nil {
			_ = m.claudeProcess.Wait()
		} else {
			_ = proc.Wait()
		}
		return m.content.String(), nil
	}
	_ = proc.Wait()
	return "", nil
}

// isatty returns true if the file descriptor is a terminal.
func isatty(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
