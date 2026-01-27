// Package events provides unified event parsing.
package events

import (
	"encoding/json"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

// re-export event types for convenience
type (
	AssistantEventData = assistant.Event
	UserEventData      = user.Event
	StreamEventData    = stream.Event
	SystemEventData    = system.Event
	ResultEventData    = result.Event
)

// Event represents a parsed event of any type.
type Event interface {
	eventMarker()
}

// SystemEvent wraps a parsed system event.
type SystemEvent struct{ Data system.Event }

func (SystemEvent) eventMarker() {}

// AssistantEvent wraps a parsed assistant event.
type AssistantEvent struct{ Data assistant.Event }

func (AssistantEvent) eventMarker() {}

// UserEvent wraps a parsed user event.
type UserEvent struct{ Data user.Event }

func (UserEvent) eventMarker() {}

// StreamEvent wraps a parsed stream event.
type StreamEvent struct{ Data stream.Event }

func (StreamEvent) eventMarker() {}

// ResultEvent wraps a parsed result event.
type ResultEvent struct{ Data result.Event }

func (ResultEvent) eventMarker() {}

// ParseError represents an error parsing an event.
type ParseError struct {
	Err  error
	Line string
}

func (ParseError) eventMarker() {}

// Parse parses a JSON line into a typed Event.
// Returns nil for empty lines.
func Parse(line string) Event {
	if line == "" {
		return nil
	}

	var base types.BaseEvent
	if err := json.Unmarshal([]byte(line), &base); err != nil {
		return ParseError{Err: err, Line: line}
	}

	switch base.Type {
	case "system":
		var event system.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseError{Err: err, Line: line}
		}
		return SystemEvent{Data: event}

	case "assistant":
		var event assistant.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseError{Err: err, Line: line}
		}
		return AssistantEvent{Data: event}

	case "user":
		var event user.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseError{Err: err, Line: line}
		}
		return UserEvent{Data: event}

	case "stream_event":
		var event stream.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseError{Err: err, Line: line}
		}
		return StreamEvent{Data: event}

	case "result":
		var event result.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return ParseError{Err: err, Line: line}
		}
		return ResultEvent{Data: event}

	default:
		return ParseError{Err: nil, Line: "Unknown event type: " + base.Type}
	}
}
