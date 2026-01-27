package tui

import (
	"bufio"

	tea "github.com/charmbracelet/bubbletea"
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
// It delegates to the events package for parsing and converts to TUI messages.
func ParseEvent(line string) tea.Msg {
	parsed := events.Parse(line)
	if parsed == nil {
		return nil
	}

	switch e := parsed.(type) {
	case events.SystemEvent:
		return SystemEventMsg{Event: e.Data}
	case events.AssistantEvent:
		return AssistantEventMsg{Event: e.Data}
	case events.UserEvent:
		return UserEventMsg{Event: e.Data}
	case events.StreamEvent:
		return StreamEventMsg{Event: e.Data}
	case events.ResultEvent:
		return ResultEventMsg{Event: e.Data}
	case events.ParseError:
		return ParseErrorMsg{Err: e.Err, Line: e.Line}
	default:
		return nil
	}
}
