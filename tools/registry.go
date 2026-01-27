package tools

import (
	"encoding/json"
	"fmt"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/types"
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

// GetToolArgFromBlock extracts the tool argument from a ContentBlock.
// Convenience wrapper for use with types.ContentBlock.
func GetToolArgFromBlock(block types.ContentBlock) string {
	if len(block.Input) == 0 {
		return ""
	}
	var input map[string]interface{}
	if err := json.Unmarshal(block.Input, &input); err != nil {
		return ""
	}
	return GetToolArg(block.Name, input)
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

// fieldExtractor returns a renderer that extracts a string field by name.
func fieldExtractor(fieldName string) ToolRendererFunc {
	return func(input map[string]interface{}) string {
		if val, ok := input[fieldName].(string); ok {
			return val
		}
		return ""
	}
}

// arrayCounter returns a renderer that counts items in an array field.
// Uses singular form for count of 1, plural otherwise.
func arrayCounter(fieldName, singular, plural string) ToolRendererFunc {
	return func(input map[string]interface{}) string {
		if arr, ok := input[fieldName].([]interface{}); ok {
			if len(arr) == 1 {
				return fmt.Sprintf("1 %s", singular)
			}
			return fmt.Sprintf("%d %s", len(arr), plural)
		}
		return ""
	}
}

// noOpRenderer always returns empty string.
var noOpRenderer = ToolRendererFunc(func(input map[string]interface{}) string {
	return ""
})

// toolDefinitions declares all built-in tool renderers declaratively.
// Each tool maps to either a field extractor, array counter, or no-op renderer.
var toolDefinitions = map[string]ToolRendererFunc{
	// Field extractors - extract a single string field
	"Bash":         fieldExtractor("command"),
	"Read":         fieldExtractor("file_path"),
	"Write":        fieldExtractor("file_path"),
	"Edit":         fieldExtractor("file_path"),
	"Glob":         fieldExtractor("pattern"),
	"Grep":         fieldExtractor("pattern"),
	"Task":         fieldExtractor("description"),
	"WebFetch":     fieldExtractor("url"),
	"WebSearch":    fieldExtractor("query"),
	"Skill":        fieldExtractor("skill"),
	"NotebookEdit": fieldExtractor("notebook_path"),
	"TaskOutput":   fieldExtractor("task_id"),
	"TaskStop":     fieldExtractor("task_id"),
	"ToolSearch":   fieldExtractor("query"),

	// Array counters - count items with singular/plural formatting
	"TodoWrite":       arrayCounter("todos", "item", "items"),
	"AskUserQuestion": arrayCounter("questions", "question", "questions"),

	// No-op renderers - tools with no meaningful arguments to display
	"EnterPlanMode": noOpRenderer,
	"ExitPlanMode":  noOpRenderer,
}

func init() {
	for name, renderer := range toolDefinitions {
		Register(name, renderer)
	}
}
