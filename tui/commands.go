package tui

import (
	"bufio"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/johnnyfreeman/viewscreen/events"
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

// ResetTerminalMouseModes clears any mouse tracking mode left behind by this
// process or another TUI. Viewscreen does not need click tracking, and stale
// mouse sequences can otherwise be decoded as keyboard text.
func ResetTerminalMouseModes() tea.Cmd {
	return tea.Raw(
		ansi.ResetModeMouseX10 +
			ansi.ResetModeMouseNormal +
			ansi.ResetModeMouseHighlight +
			ansi.ResetModeMouseButtonEvent +
			ansi.ResetModeMouseAnyEvent +
			ansi.ResetModeMouseExtUtf8 +
			ansi.ResetModeMouseExtSgr +
			ansi.ResetModeMouseExtUrxvt +
			ansi.ResetModeMouseExtSgrPixel +
			ansi.ResetModeFocusEvent,
	)
}
