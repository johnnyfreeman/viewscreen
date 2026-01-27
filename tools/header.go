// Package tools provides tool header rendering and tracking.
package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
)

// HeaderRenderer renders tool headers with configurable formatting.
// It provides a cleaner API than the individual RenderToolUse/RenderToolHeader functions,
// consolidating nested/non-nested and icon variations into a single composable type.
type HeaderRenderer struct {
	output io.Writer
	icon   string
	prefix string
}

// HeaderRendererOption configures a HeaderRenderer.
type HeaderRendererOption func(*HeaderRenderer)

// WithOutput sets the output writer.
func WithOutput(w io.Writer) HeaderRendererOption {
	return func(r *HeaderRenderer) {
		r.output = w
	}
}

// WithIcon sets the icon prefix (e.g., spinner frames).
func WithIcon(icon string) HeaderRendererOption {
	return func(r *HeaderRenderer) {
		r.icon = icon
	}
}

// WithNested adds the nested prefix for sub-agent tools.
func WithNested() HeaderRendererOption {
	return func(r *HeaderRenderer) {
		r.prefix = style.NestedPrefix
	}
}

// WithPrefix sets a custom prefix.
func WithPrefix(prefix string) HeaderRendererOption {
	return func(r *HeaderRenderer) {
		r.prefix = prefix
	}
}

// NewHeaderRenderer creates a HeaderRenderer with the given options.
func NewHeaderRenderer(opts ...HeaderRendererOption) *HeaderRenderer {
	r := &HeaderRenderer{
		output: os.Stdout,
		icon:   style.Bullet,
		prefix: "",
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RenderBlock renders a tool header from a ContentBlock.
// Returns the ToolContext for syntax highlighting of tool results.
func (r *HeaderRenderer) RenderBlock(block types.ContentBlock) ToolContext {
	input := ParseBlockInput(block)
	return r.Render(block.Name, input)
}

// RenderBlockToString renders a tool header from a ContentBlock to a string.
// Returns the rendered string and ToolContext.
func (r *HeaderRenderer) RenderBlockToString(block types.ContentBlock) (string, ToolContext) {
	input := ParseBlockInput(block)
	return r.RenderToString(block.Name, input)
}

// Render renders a tool header with the given name and input.
// Returns the ToolContext for syntax highlighting of tool results.
func (r *HeaderRenderer) Render(toolName string, input map[string]any) ToolContext {
	out := render.WriterOutput(r.output)
	return r.renderTo(out, toolName, input)
}

// RenderToString renders a tool header to a string.
// Returns the rendered string and ToolContext.
func (r *HeaderRenderer) RenderToString(toolName string, input map[string]any) (string, ToolContext) {
	out := render.StringOutput()
	ctx := r.renderTo(out, toolName, input)
	return out.String(), ctx
}

// renderTo is the core rendering logic.
func (r *HeaderRenderer) renderTo(out *render.Output, toolName string, input map[string]any) ToolContext {
	args := GetToolArg(toolName, input)

	// Truncate long args
	if len(args) > 80 {
		args = args[:77] + "..."
	}

	// Build header: [prefix][icon] ToolName args
	// Icon is printed separately since it may already have styling (e.g., spinner)
	icon := r.icon
	if icon == "" {
		icon = style.Bullet
	}
	fmt.Fprint(out, r.prefix+icon+" "+style.ApplyThemeBoldGradient(toolName))

	// Style args: file paths get muted color + dotted underline (combined in single
	// ANSI sequence via Ultraviolet), other args get just muted color
	if args != "" {
		if IsFilePathTool(toolName) {
			fmt.Fprint(out, " "+style.MutedDottedUnderline(args))
		} else {
			fmt.Fprint(out, " "+style.MutedText(args))
		}
	}
	fmt.Fprintln(out)

	return ToolContext{
		ToolName: toolName,
		FilePath: GetFilePath(toolName, input),
	}
}

// RenderBlockToStringWithNesting renders a tool header from a ContentBlock to a string,
// applying the nested prefix if isNested is true. This is a convenience method that
// consolidates the common "if nested then X else Y" pattern.
func (r *HeaderRenderer) RenderBlockToStringWithNesting(block types.ContentBlock, isNested bool) (string, ToolContext) {
	if isNested {
		return NewHeaderRenderer(WithNested()).RenderBlockToString(block)
	}
	return r.RenderBlockToString(block)
}

// ParseBlockInput parses the JSON input from a ContentBlock into a map.
// Returns nil if input is empty or parsing fails.
func ParseBlockInput(block types.ContentBlock) map[string]any {
	if len(block.Input) == 0 {
		return nil
	}
	var input map[string]any
	if err := json.Unmarshal(block.Input, &input); err != nil {
		return nil
	}
	return input
}

// RenderBlockToOutput writes a tool header to a render.Output.
// This is an adapter method for use with assistant.ToolUseRenderer.
func (r *HeaderRenderer) RenderBlockToOutput(out *render.Output, block types.ContentBlock) {
	input := ParseBlockInput(block)
	r.renderTo(out, block.Name, input)
}
