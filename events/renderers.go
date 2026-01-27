package events

import (
	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/user"
)

// RendererSet holds the complete set of renderers needed to process events.
// This consolidates the renderer initialization that was previously duplicated
// between the parser and TUI packages.
type RendererSet struct {
	System       *system.Renderer
	Assistant    *assistant.Renderer
	User         *user.Renderer
	Result       *result.Renderer
	Stream       *stream.Renderer
	PendingTools *tools.ToolUseTracker
}

// NewRendererSet creates a new RendererSet with default renderers.
func NewRendererSet() *RendererSet {
	return &RendererSet{
		System:       system.NewRenderer(),
		Assistant:    assistant.NewRenderer(),
		User:         user.NewRenderer(),
		Result:       result.NewRenderer(),
		Stream:       stream.NewRenderer(),
		PendingTools: tools.NewToolUseTracker(),
	}
}
