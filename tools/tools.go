package tools

import (
	"encoding/json"
	"fmt"
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

// renderToolHeaderTo is the core rendering logic for tool headers.
// It writes to the provided output and uses the optional prefix for nested tools.
// Returns tool context for use by the caller.
func renderToolHeaderTo(out *render.Output, prefix, toolName string, input map[string]any) ToolContext {
	args := GetToolArg(toolName, input)

	// Truncate long args
	if len(args) > 80 {
		args = args[:77] + "..."
	}

	// Dotted underline for file paths for emphasis
	styledArgs := args
	if IsFilePathTool(toolName) && args != "" {
		styledArgs = style.DottedUnderline(args)
	}

	// Build header: [prefix]● ToolName args
	fmt.Fprint(out, prefix+style.ApplyThemeBoldGradient(style.Bullet+toolName))
	if styledArgs != "" {
		fmt.Fprint(out, " "+style.Muted.Render(styledArgs))
	}
	fmt.Fprintln(out)

	return ToolContext{
		ToolName: toolName,
		FilePath: GetFilePath(toolName, input),
	}
}

// renderToolUseTo renders a tool use block, delegating to renderToolHeaderTo.
// Returns tool context for use by the caller.
func renderToolUseTo(out *render.Output, prefix string, block types.ContentBlock) ToolContext {
	if len(block.Input) > 0 {
		var input map[string]any
		if err := json.Unmarshal(block.Input, &input); err == nil {
			return renderToolHeaderTo(out, prefix, block.Name, input)
		}
	}
	// Fallback if no input or parse error
	fmt.Fprintln(out, prefix+style.ApplyThemeBoldGradient(style.Bullet+block.Name))
	return ToolContext{ToolName: block.Name, FilePath: ""}
}

// RenderToolUse renders a tool use block header and input to stdout.
// Returns tool context for syntax highlighting of tool results.
func RenderToolUse(block types.ContentBlock) ToolContext {
	return renderToolUseTo(render.WriterOutput(os.Stdout), "", block)
}

// RenderToolUseToString renders a tool use block to a string.
// Returns the rendered string and tool context.
func RenderToolUseToString(block types.ContentBlock) (string, ToolContext) {
	out := render.StringOutput()
	ctx := renderToolUseTo(out, "", block)
	return out.String(), ctx
}

// RenderToolHeader renders the tool header in format: ● ToolName args
// Returns tool context for syntax highlighting of tool results.
func RenderToolHeader(toolName string, input map[string]any) ToolContext {
	return renderToolHeaderTo(render.WriterOutput(os.Stdout), "", toolName, input)
}

// RenderToolHeaderToString renders the tool header to a string.
// Returns the rendered string and tool context.
func RenderToolHeaderToString(toolName string, input map[string]any) (string, ToolContext) {
	out := render.StringOutput()
	ctx := renderToolHeaderTo(out, "", toolName, input)
	return out.String(), ctx
}

// RenderNestedToolUse renders a tool use block with nested indentation for sub-agent tools.
// Returns tool context for syntax highlighting of tool results.
func RenderNestedToolUse(block types.ContentBlock) ToolContext {
	return renderToolUseTo(render.WriterOutput(os.Stdout), style.NestedPrefix, block)
}

// RenderNestedToolUseToString renders a nested tool use block to a string.
// Returns the rendered string and tool context.
func RenderNestedToolUseToString(block types.ContentBlock) (string, ToolContext) {
	out := render.StringOutput()
	ctx := renderToolUseTo(out, style.NestedPrefix, block)
	return out.String(), ctx
}
