package tui

import (
	"bufio"

	tea "charm.land/bubbletea/v2"
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
