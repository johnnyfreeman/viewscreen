package tools

import (
	"encoding/json"
	"fmt"

	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

// RenderToolUse renders a tool use block header and input
func RenderToolUse(block types.ContentBlock) {
	if len(block.Input) > 0 {
		var input map[string]interface{}
		if err := json.Unmarshal(block.Input, &input); err == nil {
			RenderToolHeader(block.Name, input)
			return
		}
	}
	// Fallback if no input or parse error
	fmt.Println(style.ToolHeader.Render(fmt.Sprintf("%s%s()", style.Bullet, block.Name)))
	user.SetToolContext(block.Name, "")
}

// RenderToolHeader renders the tool header in Claude Code format: ● ToolName(args)
func RenderToolHeader(toolName string, input map[string]interface{}) {
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
	fmt.Print(style.ToolHeader.Render(header))
	fmt.Print(styledArgs)
	fmt.Println(style.ToolHeader.Render(")"))

	// Set context for syntax highlighting of tool results
	filePath := GetFilePath(toolName, input)
	user.SetToolContext(toolName, filePath)
}
