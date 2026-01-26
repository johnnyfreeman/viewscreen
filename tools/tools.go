package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

// ToolRenderConfig holds configuration for tool rendering
type ToolRenderConfig struct {
	output io.Writer
}

// ToolRenderOption is a functional option for configuring tool rendering
type ToolRenderOption func(*ToolRenderConfig)

// WithOutput sets a custom output writer for tool rendering
func WithOutput(w io.Writer) ToolRenderOption {
	return func(c *ToolRenderConfig) {
		c.output = w
	}
}

// newRenderConfig creates a new render config with the given options
func newRenderConfig(opts ...ToolRenderOption) *ToolRenderConfig {
	c := &ToolRenderConfig{
		output: os.Stdout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// RenderToolUse renders a tool use block header and input
func RenderToolUse(block types.ContentBlock, opts ...ToolRenderOption) {
	cfg := newRenderConfig(opts...)
	if len(block.Input) > 0 {
		var input map[string]interface{}
		if err := json.Unmarshal(block.Input, &input); err == nil {
			RenderToolHeader(block.Name, input, opts...)
			return
		}
	}
	// Fallback if no input or parse error
	fmt.Fprintln(cfg.output, style.ToolHeader.Render(fmt.Sprintf("%s%s()", style.Bullet, block.Name)))
	user.SetToolContext(block.Name, "")
}

// RenderToolHeader renders the tool header in Claude Code format: ● ToolName(args)
func RenderToolHeader(toolName string, input map[string]interface{}, opts ...ToolRenderOption) {
	cfg := newRenderConfig(opts...)

	// Use registry to get the argument string
	args := GetToolArg(toolName, input)

	// Truncate long args
	truncated := false
	if len(args) > 80 {
		args = args[:77]
		truncated = true
	}

	// Dotted underline for file paths for emphasis
	var styledArgs string
	if IsFilePathTool(toolName) && args != "" {
		if truncated {
			args += "..."
		}
		styledArgs = style.DottedUnderline(args)
	} else {
		if truncated {
			args += "..."
		}
		styledArgs = args
	}

	// Build header: ● ToolName(args)
	header := fmt.Sprintf("%s%s(", style.Bullet, toolName)
	fmt.Fprint(cfg.output, style.ToolHeader.Render(header))
	fmt.Fprint(cfg.output, styledArgs)
	fmt.Fprintln(cfg.output, style.ToolHeader.Render(")"))

	// Set context for syntax highlighting of tool results
	filePath := GetFilePath(toolName, input)
	user.SetToolContext(toolName, filePath)
}

// RenderToolUseDefault is a convenience wrapper for RenderToolUse that uses default output
// This matches the ToolUseRenderer function signature expected by other packages
func RenderToolUseDefault(block types.ContentBlock) {
	RenderToolUse(block)
}

// RenderToolHeaderDefault is a convenience wrapper for RenderToolHeader that uses default output
// This matches the ToolHeaderRenderer function signature expected by other packages
func RenderToolHeaderDefault(toolName string, input map[string]interface{}) {
	RenderToolHeader(toolName, input)
}
