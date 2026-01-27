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
	fmt.Fprintln(cfg.output, style.ApplyThemeBoldGradient(style.Bullet+block.Name))
	user.SetToolContext(block.Name, "")
}

// RenderToolHeader renders the tool header in format: ● ToolName args
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

	// Build header: ● ToolName args
	fmt.Fprint(cfg.output, style.ApplyThemeBoldGradient(style.Bullet+toolName))
	if styledArgs != "" {
		fmt.Fprint(cfg.output, " "+style.Muted.Render(styledArgs))
	}
	fmt.Fprintln(cfg.output)

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

// RenderToolUseToString renders a tool use block to a string
func RenderToolUseToString(block types.ContentBlock) string {
	if len(block.Input) > 0 {
		var input map[string]interface{}
		if err := json.Unmarshal(block.Input, &input); err == nil {
			return RenderToolHeaderToString(block.Name, input)
		}
	}
	// Fallback if no input or parse error
	return style.ApplyThemeBoldGradient(style.Bullet+block.Name) + "\n"
}

// RenderToolHeaderToString renders the tool header to a string in format: ● ToolName args
func RenderToolHeaderToString(toolName string, input map[string]interface{}) string {
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

	// Build header: ● ToolName args
	var result string
	result = style.ApplyThemeBoldGradient(style.Bullet + toolName)
	if styledArgs != "" {
		result += " " + style.Muted.Render(styledArgs)
	}
	result += "\n"

	// Set context for syntax highlighting of tool results
	filePath := GetFilePath(toolName, input)
	user.SetToolContext(toolName, filePath)

	return result
}

// RenderNestedToolUse renders a tool use block with nested indentation for sub-agent tools
func RenderNestedToolUse(block types.ContentBlock, opts ...ToolRenderOption) {
	cfg := newRenderConfig(opts...)
	if len(block.Input) > 0 {
		var input map[string]interface{}
		if err := json.Unmarshal(block.Input, &input); err == nil {
			RenderNestedToolHeader(block.Name, input, opts...)
			return
		}
	}
	// Fallback if no input or parse error
	fmt.Fprintln(cfg.output, style.NestedPrefix+style.ApplyThemeBoldGradient(style.Bullet+block.Name))
	user.SetToolContext(block.Name, "")
}

// RenderNestedToolHeader renders a nested tool header with indentation
func RenderNestedToolHeader(toolName string, input map[string]interface{}, opts ...ToolRenderOption) {
	cfg := newRenderConfig(opts...)

	args := GetToolArg(toolName, input)

	truncated := false
	if len(args) > 80 {
		args = args[:77]
		truncated = true
	}

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

	// Build nested header: │ ● ToolName args
	fmt.Fprint(cfg.output, style.NestedPrefix+style.ApplyThemeBoldGradient(style.Bullet+toolName))
	if styledArgs != "" {
		fmt.Fprint(cfg.output, " "+style.Muted.Render(styledArgs))
	}
	fmt.Fprintln(cfg.output)

	filePath := GetFilePath(toolName, input)
	user.SetToolContext(toolName, filePath)
}

// RenderNestedToolUseToString renders a nested tool use block to a string
func RenderNestedToolUseToString(block types.ContentBlock) string {
	if len(block.Input) > 0 {
		var input map[string]interface{}
		if err := json.Unmarshal(block.Input, &input); err == nil {
			return RenderNestedToolHeaderToString(block.Name, input)
		}
	}
	// Fallback if no input or parse error
	return style.NestedPrefix + style.ApplyThemeBoldGradient(style.Bullet+block.Name) + "\n"
}

// RenderNestedToolHeaderToString renders a nested tool header to a string
func RenderNestedToolHeaderToString(toolName string, input map[string]interface{}) string {
	args := GetToolArg(toolName, input)

	truncated := false
	if len(args) > 80 {
		args = args[:77]
		truncated = true
	}

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

	// Build nested header: │ ● ToolName args
	var result string
	result = style.NestedPrefix + style.ApplyThemeBoldGradient(style.Bullet+toolName)
	if styledArgs != "" {
		result += " " + style.Muted.Render(styledArgs)
	}
	result += "\n"

	filePath := GetFilePath(toolName, input)
	user.SetToolContext(toolName, filePath)

	return result
}
