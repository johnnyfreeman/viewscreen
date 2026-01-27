package tools

import (
	"os"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
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

// RenderHeaderTo is the core rendering logic for tool headers.
// It writes to the provided output using the specified options.
// Returns tool context for use by the caller.
//
// Deprecated: Use HeaderRenderer for new code. This function is maintained
// for backward compatibility.
func RenderHeaderTo(out *render.Output, toolName string, input map[string]any, opts HeaderOptions) ToolContext {
	r := NewHeaderRenderer(WithIcon(opts.Icon), WithPrefix(opts.Prefix))
	return r.renderTo(out, toolName, input)
}

// RenderToolUse renders a tool use block header and input to stdout.
// Returns tool context for syntax highlighting of tool results.
func RenderToolUse(block types.ContentBlock) ToolContext {
	return NewHeaderRenderer().RenderBlock(block)
}

// RenderToolUseToString renders a tool use block to a string.
// Returns the rendered string and tool context.
func RenderToolUseToString(block types.ContentBlock) (string, ToolContext) {
	return NewHeaderRenderer().RenderBlockToString(block)
}

// RenderToolHeader renders the tool header in format: ● ToolName args
// Returns tool context for syntax highlighting of tool results.
func RenderToolHeader(toolName string, input map[string]any) ToolContext {
	return NewHeaderRenderer().Render(toolName, input)
}

// RenderToolHeaderToString renders the tool header to a string.
// Returns the rendered string and tool context.
func RenderToolHeaderToString(toolName string, input map[string]any) (string, ToolContext) {
	return NewHeaderRenderer().RenderToString(toolName, input)
}

// RenderNestedToolUse renders a tool use block with nested indentation for sub-agent tools.
// Returns tool context for syntax highlighting of tool results.
func RenderNestedToolUse(block types.ContentBlock) ToolContext {
	return NewHeaderRenderer(WithOutput(os.Stdout), WithNested()).RenderBlock(block)
}

// RenderNestedToolUseToString renders a nested tool use block to a string.
// Returns the rendered string and tool context.
func RenderNestedToolUseToString(block types.ContentBlock) (string, ToolContext) {
	return NewHeaderRenderer(WithNested()).RenderBlockToString(block)
}

// RenderHeaderToString renders a tool header to string with the given options.
// Returns the rendered string.
func RenderHeaderToString(toolName string, input map[string]any, opts HeaderOptions) string {
	r := NewHeaderRenderer(WithIcon(opts.Icon), WithPrefix(opts.Prefix))
	str, _ := r.RenderToString(toolName, input)
	return str
}
