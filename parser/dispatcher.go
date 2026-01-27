package parser

import (
	"fmt"
	"io"
)

// EventHandler processes a parsed event of a specific type.
// The line parameter contains the raw JSON for the event.
type EventDispatchHandler func(line []byte) error

// EventDispatcher routes events to their registered handlers by type.
// It follows the same registry pattern as tools/registry.go.
type EventDispatcher struct {
	handlers  map[string]EventDispatchHandler
	errOutput io.Writer
}

// NewEventDispatcher creates a new dispatcher with the given error output.
func NewEventDispatcher(errOutput io.Writer) *EventDispatcher {
	return &EventDispatcher{
		handlers:  make(map[string]EventDispatchHandler),
		errOutput: errOutput,
	}
}

// Register adds a handler for the given event type.
func (d *EventDispatcher) Register(eventType string, handler EventDispatchHandler) {
	d.handlers[eventType] = handler
}

// Dispatch routes an event to its handler based on type.
// Returns true if the event was handled, false if no handler was found.
func (d *EventDispatcher) Dispatch(eventType string, line []byte) bool {
	handler, ok := d.handlers[eventType]
	if !ok {
		return false
	}

	if err := handler(line); err != nil {
		// Log parse errors but don't stop processing
		fmt.Fprintf(d.errOutput, "Error parsing %s event: %v\n", eventType, err)
	}
	return true
}
