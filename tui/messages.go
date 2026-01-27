package tui

import (
	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/user"
)

// RawLineMsg is sent when a line is read from stdin
type RawLineMsg struct {
	Line string
}

// StdinClosedMsg is sent when stdin is closed
type StdinClosedMsg struct {
	Err error
}

// ParseErrorMsg is sent when there's an error parsing JSON
type ParseErrorMsg struct {
	Err  error
	Line string
}

// SystemEventMsg wraps a parsed system event
type SystemEventMsg struct {
	Event system.Event
}

// AssistantEventMsg wraps a parsed assistant event
type AssistantEventMsg struct {
	Event assistant.Event
}

// UserEventMsg wraps a parsed user event
type UserEventMsg struct {
	Event user.Event
}

// StreamEventMsg wraps a parsed stream event
type StreamEventMsg struct {
	Event stream.Event
}

// ResultEventMsg wraps a parsed result event
type ResultEventMsg struct {
	Event result.Event
}

// WindowSizeMsg is sent when the terminal window size changes
type WindowSizeMsg struct {
	Width  int
	Height int
}
