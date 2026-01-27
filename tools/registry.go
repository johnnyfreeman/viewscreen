package tools

import (
	"encoding/json"
	"fmt"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/types"
)

// ToolDefinition describes a tool's metadata and rendering behavior.
// This consolidates all tool-specific information in one place.
type ToolDefinition struct {
	// Name is the tool identifier (e.g., "Read", "Bash")
	Name string

	// HeaderField is the input field to display in the tool header.
	// Empty for tools that don't show arguments.
	HeaderField string

	// FilePathField is the input field containing a file path (for syntax highlighting).
	// Empty for tools that don't operate on files.
	FilePathField string

	// FilePathFallback is tried if FilePathField is not found in the input.
	// Used by NotebookEdit which accepts both notebook_path and file_path.
	FilePathFallback string

	// CountField is the input field containing an array to count.
	// Used with Singular/Plural for "N items" style headers.
	CountField string
	Singular   string
	Plural     string
}

// RenderHeader returns the argument string to display in the tool header.
func (d ToolDefinition) RenderHeader(input map[string]interface{}) string {
	// Count-based rendering (e.g., "3 items")
	if d.CountField != "" {
		if arr, ok := input[d.CountField].([]interface{}); ok {
			if len(arr) == 1 {
				return fmt.Sprintf("1 %s", d.Singular)
			}
			return fmt.Sprintf("%d %s", len(arr), d.Plural)
		}
		return ""
	}

	// Field-based rendering
	if d.HeaderField != "" {
		if val, ok := input[d.HeaderField].(string); ok {
			return val
		}
	}
	return ""
}

// GetFilePath extracts the file path from the input if this tool operates on files.
func (d ToolDefinition) GetFilePath(input map[string]interface{}) string {
	if d.FilePathField == "" {
		return ""
	}
	if path, ok := input[d.FilePathField].(string); ok {
		return path
	}
	// Try fallback field if primary field not found
	if d.FilePathFallback != "" {
		if path, ok := input[d.FilePathFallback].(string); ok {
			return path
		}
	}
	return ""
}

// IsFilePathTool returns true if this tool operates on a file path.
func (d ToolDefinition) IsFilePathTool() bool {
	return d.FilePathField != ""
}

// definitions holds all tool definitions, keyed by name.
// This is the single source of truth for tool metadata.
var definitions = map[string]ToolDefinition{}

// RegisterDefinition adds a tool definition to the registry.
func RegisterDefinition(def ToolDefinition) {
	definitions[def.Name] = def
}

// GetDefinition returns the definition for a tool, or an empty definition if not found.
func GetDefinition(name string) (ToolDefinition, bool) {
	def, ok := definitions[name]
	return def, ok
}

// GetToolArg returns the display argument for a tool using the registry.
// Falls back to JSON preview for unknown tools in verbose mode.
func GetToolArg(toolName string, input map[string]interface{}) string {
	return GetToolArgWithConfig(toolName, input, config.DefaultProvider{})
}

// GetToolArgWithConfig returns the display argument for a tool using the provided config.
// Falls back to JSON preview for unknown tools in verbose mode.
func GetToolArgWithConfig(toolName string, input map[string]interface{}, cfg config.Provider) string {
	if def, ok := GetDefinition(toolName); ok {
		return def.RenderHeader(input)
	}

	// Fallback: show compact JSON for unknown tools in verbose mode
	if cfg.IsVerbose() {
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
	if def, ok := definitions[toolName]; ok {
		return def.GetFilePath(input)
	}
	return ""
}

// IsFilePathTool returns true if the tool's argument is a file path
func IsFilePathTool(toolName string) bool {
	if def, ok := definitions[toolName]; ok {
		return def.IsFilePathTool()
	}
	return false
}

// builtinTools declares all built-in tool definitions.
// Each definition consolidates the tool's header rendering, file path extraction,
// and other metadata in one place.
var builtinTools = []ToolDefinition{
	// File operations - tools that operate on file paths
	{Name: "Read", HeaderField: "file_path", FilePathField: "file_path"},
	{Name: "Write", HeaderField: "file_path", FilePathField: "file_path"},
	{Name: "Edit", HeaderField: "file_path", FilePathField: "file_path"},
	{Name: "NotebookEdit", HeaderField: "notebook_path", FilePathField: "notebook_path", FilePathFallback: "file_path"},

	// Simple field extractors - display a single string field
	{Name: "Bash", HeaderField: "command"},
	{Name: "Glob", HeaderField: "pattern"},
	{Name: "Grep", HeaderField: "pattern"},
	{Name: "Task", HeaderField: "description"},
	{Name: "WebFetch", HeaderField: "url"},
	{Name: "WebSearch", HeaderField: "query"},
	{Name: "Skill", HeaderField: "skill"},
	{Name: "TaskOutput", HeaderField: "task_id"},
	{Name: "TaskStop", HeaderField: "task_id"},
	{Name: "ToolSearch", HeaderField: "query"},

	// Array counters - count items with singular/plural formatting
	{Name: "TodoWrite", CountField: "todos", Singular: "item", Plural: "items"},
	{Name: "AskUserQuestion", CountField: "questions", Singular: "question", Plural: "questions"},

	// No-op renderers - tools with no meaningful arguments to display
	{Name: "EnterPlanMode"},
	{Name: "ExitPlanMode"},
}

func init() {
	for _, def := range builtinTools {
		RegisterDefinition(def)
	}
}
