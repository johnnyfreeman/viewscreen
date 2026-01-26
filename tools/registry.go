package tools

import (
	"encoding/json"
	"fmt"

	"github.com/johnnyfreeman/viewscreen/config"
)

// ToolRenderer defines the interface for rendering tool headers
type ToolRenderer interface {
	// RenderHeader returns the argument string to display in the tool header.
	// Returns empty string if no specific argument should be shown.
	RenderHeader(input map[string]interface{}) string
}

// ToolRendererFunc is a function adapter for ToolRenderer
type ToolRendererFunc func(input map[string]interface{}) string

func (f ToolRendererFunc) RenderHeader(input map[string]interface{}) string {
	return f(input)
}

// registry holds tool-specific renderers
var registry = map[string]ToolRenderer{}

// Register adds a tool renderer to the registry
func Register(name string, r ToolRenderer) {
	registry[name] = r
}

// RegisterFunc is a convenience method to register a function as a renderer
func RegisterFunc(name string, f func(input map[string]interface{}) string) {
	registry[name] = ToolRendererFunc(f)
}

// GetRenderer returns the renderer for a tool, or nil if not found
func GetRenderer(name string) ToolRenderer {
	return registry[name]
}

// GetToolArg returns the display argument for a tool using the registry.
// Falls back to JSON preview for unknown tools in verbose mode.
func GetToolArg(toolName string, input map[string]interface{}) string {
	if r := GetRenderer(toolName); r != nil {
		return r.RenderHeader(input)
	}

	// Fallback: show compact JSON for unknown tools in verbose mode
	if config.Verbose {
		if data, err := json.Marshal(input); err == nil {
			s := string(data)
			if len(s) > 100 {
				s = s[:100] + "..."
			}
			return s
		}
	}
	return ""
}

// GetFilePath extracts the file path from tool input if present.
// Used for syntax highlighting context.
func GetFilePath(toolName string, input map[string]interface{}) string {
	switch toolName {
	case "Read", "Write", "Edit", "NotebookEdit":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
		if path, ok := input["notebook_path"].(string); ok {
			return path
		}
	}
	return ""
}

// filePathTools lists tools whose primary argument is a file path
var filePathTools = map[string]bool{
	"Read":         true,
	"Write":        true,
	"Edit":         true,
	"NotebookEdit": true,
}

// IsFilePathTool returns true if the tool's argument is a file path
func IsFilePathTool(toolName string) bool {
	return filePathTools[toolName]
}

func init() {
	// Register built-in tool renderers
	RegisterFunc("Bash", func(input map[string]interface{}) string {
		if cmd, ok := input["command"].(string); ok {
			return cmd
		}
		return ""
	})

	RegisterFunc("Read", func(input map[string]interface{}) string {
		if path, ok := input["file_path"].(string); ok {
			return path
		}
		return ""
	})

	RegisterFunc("Write", func(input map[string]interface{}) string {
		if path, ok := input["file_path"].(string); ok {
			return path
		}
		return ""
	})

	RegisterFunc("Edit", func(input map[string]interface{}) string {
		if path, ok := input["file_path"].(string); ok {
			return path
		}
		return ""
	})

	RegisterFunc("Glob", func(input map[string]interface{}) string {
		if pattern, ok := input["pattern"].(string); ok {
			return pattern
		}
		return ""
	})

	RegisterFunc("Grep", func(input map[string]interface{}) string {
		if pattern, ok := input["pattern"].(string); ok {
			return pattern
		}
		return ""
	})

	RegisterFunc("Task", func(input map[string]interface{}) string {
		if desc, ok := input["description"].(string); ok {
			return desc
		}
		return ""
	})

	RegisterFunc("WebFetch", func(input map[string]interface{}) string {
		if url, ok := input["url"].(string); ok {
			return url
		}
		return ""
	})

	RegisterFunc("WebSearch", func(input map[string]interface{}) string {
		if query, ok := input["query"].(string); ok {
			return query
		}
		return ""
	})

	RegisterFunc("TodoWrite", func(input map[string]interface{}) string {
		if todos, ok := input["todos"].([]interface{}); ok {
			return fmt.Sprintf("%d items", len(todos))
		}
		return ""
	})

	// AskUserQuestion displays the number of questions being asked
	RegisterFunc("AskUserQuestion", func(input map[string]interface{}) string {
		if questions, ok := input["questions"].([]interface{}); ok {
			if len(questions) == 1 {
				return "1 question"
			}
			return fmt.Sprintf("%d questions", len(questions))
		}
		return ""
	})

	RegisterFunc("Skill", func(input map[string]interface{}) string {
		if skill, ok := input["skill"].(string); ok {
			return skill
		}
		return ""
	})

	// NotebookEdit displays the notebook path
	RegisterFunc("NotebookEdit", func(input map[string]interface{}) string {
		if path, ok := input["notebook_path"].(string); ok {
			return path
		}
		return ""
	})

	// TaskOutput displays the task ID being checked
	RegisterFunc("TaskOutput", func(input map[string]interface{}) string {
		if taskID, ok := input["task_id"].(string); ok {
			return taskID
		}
		return ""
	})

	// TaskStop displays the task ID being stopped
	RegisterFunc("TaskStop", func(input map[string]interface{}) string {
		if taskID, ok := input["task_id"].(string); ok {
			return taskID
		}
		return ""
	})

	// EnterPlanMode has no meaningful arguments to display
	RegisterFunc("EnterPlanMode", func(input map[string]interface{}) string {
		return ""
	})

	// ExitPlanMode has no meaningful arguments to display
	RegisterFunc("ExitPlanMode", func(input map[string]interface{}) string {
		return ""
	})

	// ToolSearch displays the search query
	RegisterFunc("ToolSearch", func(input map[string]interface{}) string {
		if query, ok := input["query"].(string); ok {
			return query
		}
		return ""
	})
}
