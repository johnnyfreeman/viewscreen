package tui

import (
	"bufio"
	"io"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/johnnyfreeman/viewscreen/events"
	"golang.org/x/term"
)

// ReadStdinLine returns a command that reads the next line from stdin
func ReadStdinLine(scanner *bufio.Scanner) tea.Cmd {
	return func() tea.Msg {
		if scanner.Scan() {
			return RawLineMsg{Line: scanner.Text()}
		}
		return StdinClosedMsg{Err: scanner.Err()}
	}
}

// ParseEvent parses a JSON line and returns the appropriate tea.Msg.
// Returns the events.Event directly since it already implements the tea.Msg interface.
func ParseEvent(line string) tea.Msg {
	return events.Parse(line)
}

// AutoExitTick returns a command that sends an AutoExitTickMsg after 1 second.
func AutoExitTick() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return AutoExitTickMsg{}
	}
}

// resetTerminalModes clears terminal modes that can leak in from a previous TUI
// before Bubble Tea starts reading input.
func resetTerminalModes(w io.Writer) {
	if w == nil {
		return
	}
	_, _ = io.WriteString(w, terminalModeResetSequence())
}

func terminalModeResetSequence() string {
	return ansi.KittyKeyboard(0, 1) +
		ansi.ResetModeBracketedPaste +
		ansi.ResetModeMouseX10 +
		ansi.ResetModeMouseNormal +
		ansi.ResetModeMouseHighlight +
		ansi.ResetModeMouseButtonEvent +
		ansi.ResetModeMouseAnyEvent +
		ansi.ResetModeMouseExtUtf8 +
		ansi.ResetModeMouseExtSgr +
		ansi.ResetModeMouseExtUrxvt +
		ansi.ResetModeMouseExtSgrPixel +
		ansi.ResetModeFocusEvent
}

func detectTerminalSize(f interface {
	Fd() uintptr
}) (int, int) {
	if f == nil || !term.IsTerminal(int(f.Fd())) {
		return defaultInitialWidth, defaultInitialHeight
	}
	width, height, err := term.GetSize(int(f.Fd()))
	if err != nil || width <= 0 || height <= 0 {
		return defaultInitialWidth, defaultInitialHeight
	}
	return width, height
}
