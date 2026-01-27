package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

// renderToolHeaderTo is the core rendering logic for tool headers.
// It writes to the provided output and uses the optional prefix for nested tools.
func renderToolHeaderTo(out *render.Output, prefix, toolName string, input map[string]any) {
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

	// Set context for syntax highlighting of tool results
	filePath := GetFilePath(toolName, input)
	user.SetToolContext(toolName, filePath)
}

// renderToolUseTo renders a tool use block, delegating to renderToolHeaderTo.
func renderToolUseTo(out *render.Output, prefix string, block types.ContentBlock) {
	if len(block.Input) > 0 {
		var input map[string]any
		if err := json.Unmarshal(block.Input, &input); err == nil {
			renderToolHeaderTo(out, prefix, block.Name, input)
			return
		}
	}
	// Fallback if no input or parse error
	fmt.Fprintln(out, prefix+style.ApplyThemeBoldGradient(style.Bullet+block.Name))
	user.SetToolContext(block.Name, "")
}

// RenderToolUse renders a tool use block header and input to stdout.
func RenderToolUse(block types.ContentBlock) {
	renderToolUseTo(render.WriterOutput(os.Stdout), "", block)
}

// RenderToolUseToString renders a tool use block to a string.
func RenderToolUseToString(block types.ContentBlock) string {
	out := render.StringOutput()
	renderToolUseTo(out, "", block)
	return out.String()
}

// RenderToolHeader renders the tool header in format: ● ToolName args
func RenderToolHeader(toolName string, input map[string]any) {
	renderToolHeaderTo(render.WriterOutput(os.Stdout), "", toolName, input)
}

// RenderToolHeaderToString renders the tool header to a string.
func RenderToolHeaderToString(toolName string, input map[string]any) string {
	out := render.StringOutput()
	renderToolHeaderTo(out, "", toolName, input)
	return out.String()
}

// RenderNestedToolUse renders a tool use block with nested indentation for sub-agent tools.
func RenderNestedToolUse(block types.ContentBlock) {
	renderToolUseTo(render.WriterOutput(os.Stdout), style.NestedPrefix, block)
}

// RenderNestedToolUseToString renders a nested tool use block to a string.
func RenderNestedToolUseToString(block types.ContentBlock) string {
	out := render.StringOutput()
	renderToolUseTo(out, style.NestedPrefix, block)
	return out.String()
}

// RenderNestedToolHeader renders a nested tool header with indentation.
func RenderNestedToolHeader(toolName string, input map[string]any) {
	renderToolHeaderTo(render.WriterOutput(os.Stdout), style.NestedPrefix, toolName, input)
}

// RenderNestedToolHeaderToString renders a nested tool header to a string.
func RenderNestedToolHeaderToString(toolName string, input map[string]any) string {
	out := render.StringOutput()
	renderToolHeaderTo(out, style.NestedPrefix, toolName, input)
	return out.String()
}
