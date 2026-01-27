package tools

import (
	"github.com/johnnyfreeman/viewscreen/style"
)

// ToolContext holds information about a tool use for syntax highlighting.
type ToolContext struct {
	ToolName string
	FilePath string
}

// HeaderOptions configures how a tool header is rendered.
type HeaderOptions struct {
	// Icon is the prefix icon (default: style.Bullet "● ")
	Icon string
	// Prefix is prepended to the header (e.g., style.NestedPrefix for sub-agents)
	Prefix string
}

// DefaultHeaderOptions returns the default options for tool header rendering.
func DefaultHeaderOptions() HeaderOptions {
	return HeaderOptions{
		Icon:   style.Bullet,
		Prefix: "",
	}
}
