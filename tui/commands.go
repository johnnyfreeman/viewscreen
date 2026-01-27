package tui

import (
	"bufio"
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
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

// ParseEvent parses a JSON line and returns the appropriate message
func ParseEvent(line string) tea.Msg {
	if line == "" {
		return nil
	}

	// Parse base event to determine type
	var base types.BaseEvent
	if err := json.Unmarshal([]byte(line), &base); err != nil {
		return ParseErrorMsg{Err: err, Line: line}
	}

	switch base.Type {
	case "system":
		var event system.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseErrorMsg{Err: err, Line: line}
		}
		return SystemEventMsg{Event: event}

	case "assistant":
		var event assistant.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseErrorMsg{Err: err, Line: line}
		}
		return AssistantEventMsg{Event: event}

	case "user":
		var event user.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseErrorMsg{Err: err, Line: line}
		}
		return UserEventMsg{Event: event}

	case "stream_event":
		var event stream.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseErrorMsg{Err: err, Line: line}
		}
		return StreamEventMsg{Event: event}

	case "result":
		var event result.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseErrorMsg{Err: err, Line: line}
		}
		return ResultEventMsg{Event: event}

	default:
		return ParseErrorMsg{Err: nil, Line: "Unknown event type: " + base.Type}
	}
}
