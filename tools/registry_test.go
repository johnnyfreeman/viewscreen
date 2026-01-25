package tools

import (
	"testing"

	"github.com/johnnyfreeman/viewscreen/config"
)

func TestGetRenderer(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		wantNil  bool
	}{
		{name: "Bash renderer exists", toolName: "Bash", wantNil: false},
		{name: "Read renderer exists", toolName: "Read", wantNil: false},
		{name: "Write renderer exists", toolName: "Write", wantNil: false},
		{name: "Edit renderer exists", toolName: "Edit", wantNil: false},
		{name: "Glob renderer exists", toolName: "Glob", wantNil: false},
		{name: "Grep renderer exists", toolName: "Grep", wantNil: false},
		{name: "Task renderer exists", toolName: "Task", wantNil: false},
		{name: "WebFetch renderer exists", toolName: "WebFetch", wantNil: false},
		{name: "WebSearch renderer exists", toolName: "WebSearch", wantNil: false},
		{name: "TodoRead renderer exists", toolName: "TodoRead", wantNil: false},
		{name: "TodoWrite renderer exists", toolName: "TodoWrite", wantNil: false},
		{name: "AskUser renderer exists", toolName: "AskUser", wantNil: false},
		{name: "unknown tool returns nil", toolName: "UnknownTool", wantNil: true},
		{name: "empty string returns nil", toolName: "", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := GetRenderer(tt.toolName)
			if (r == nil) != tt.wantNil {
				t.Errorf("GetRenderer(%q) = %v, wantNil = %v", tt.toolName, r, tt.wantNil)
			}
		})
	}
}

func TestToolRenderers(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]interface{}
		expected string
	}{
		// Bash
		{
			name:     "Bash with command",
			toolName: "Bash",
			input:    map[string]interface{}{"command": "ls -la"},
			expected: "ls -la",
		},
		{
			name:     "Bash without command",
			toolName: "Bash",
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "Bash with non-string command",
			toolName: "Bash",
			input:    map[string]interface{}{"command": 123},
			expected: "",
		},
		// Read
		{
			name:     "Read with file_path",
			toolName: "Read",
			input:    map[string]interface{}{"file_path": "/path/to/file.go"},
			expected: "/path/to/file.go",
		},
		{
			name:     "Read without file_path",
			toolName: "Read",
			input:    map[string]interface{}{},
			expected: "",
		},
		// Write
		{
			name:     "Write with file_path",
			toolName: "Write",
			input:    map[string]interface{}{"file_path": "/path/to/new.go"},
			expected: "/path/to/new.go",
		},
		// Edit
		{
			name:     "Edit with file_path",
			toolName: "Edit",
			input:    map[string]interface{}{"file_path": "/path/to/edit.go"},
			expected: "/path/to/edit.go",
		},
		// Glob
		{
			name:     "Glob with pattern",
			toolName: "Glob",
			input:    map[string]interface{}{"pattern": "**/*.go"},
			expected: "**/*.go",
		},
		{
			name:     "Glob without pattern",
			toolName: "Glob",
			input:    map[string]interface{}{},
			expected: "",
		},
		// Grep
		{
			name:     "Grep with pattern",
			toolName: "Grep",
			input:    map[string]interface{}{"pattern": "TODO:"},
			expected: "TODO:",
		},
		// Task
		{
			name:     "Task with description",
			toolName: "Task",
			input:    map[string]interface{}{"description": "Explore codebase"},
			expected: "Explore codebase",
		},
		{
			name:     "Task without description",
			toolName: "Task",
			input:    map[string]interface{}{"prompt": "some prompt"},
			expected: "",
		},
		// WebFetch
		{
			name:     "WebFetch with url",
			toolName: "WebFetch",
			input:    map[string]interface{}{"url": "https://example.com"},
			expected: "https://example.com",
		},
		// WebSearch
		{
			name:     "WebSearch with query",
			toolName: "WebSearch",
			input:    map[string]interface{}{"query": "golang testing"},
			expected: "golang testing",
		},
		// TodoRead
		{
			name:     "TodoRead always empty",
			toolName: "TodoRead",
			input:    map[string]interface{}{},
			expected: "",
		},
		// TodoWrite
		{
			name:     "TodoWrite with todos",
			toolName: "TodoWrite",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{"content": "task1"},
					map[string]interface{}{"content": "task2"},
					map[string]interface{}{"content": "task3"},
				},
			},
			expected: "3 items",
		},
		{
			name:     "TodoWrite without todos",
			toolName: "TodoWrite",
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "TodoWrite with non-array todos",
			toolName: "TodoWrite",
			input:    map[string]interface{}{"todos": "not an array"},
			expected: "",
		},
		// AskUser
		{
			name:     "AskUser with question",
			toolName: "AskUser",
			input:    map[string]interface{}{"question": "What is your name?"},
			expected: "What is your name?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := GetRenderer(tt.toolName)
			if r == nil {
				t.Fatalf("GetRenderer(%q) returned nil", tt.toolName)
			}
			result := r.RenderHeader(tt.input)
			if result != tt.expected {
				t.Errorf("RenderHeader() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetToolArg(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]interface{}
		verbose  bool
		expected string
	}{
		{
			name:     "known tool returns header",
			toolName: "Bash",
			input:    map[string]interface{}{"command": "echo hello"},
			verbose:  false,
			expected: "echo hello",
		},
		{
			name:     "unknown tool non-verbose returns empty",
			toolName: "CustomTool",
			input:    map[string]interface{}{"key": "value"},
			verbose:  false,
			expected: "",
		},
		{
			name:     "unknown tool verbose returns JSON",
			toolName: "CustomTool",
			input:    map[string]interface{}{"key": "value"},
			verbose:  true,
			expected: `{"key":"value"}`,
		},
		{
			name:     "unknown tool verbose truncates long JSON",
			toolName: "CustomTool",
			input: map[string]interface{}{
				"very_long_key_name": "this is a very long value that will make the JSON exceed one hundred characters when serialized",
			},
			verbose:  true,
			expected: `{"very_long_key_name":"this is a very long value that will make the JSON exceed one hundred characte...`,
		},
		{
			name:     "nil input non-verbose",
			toolName: "UnknownTool",
			input:    nil,
			verbose:  false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore verbose setting
			oldVerbose := config.Verbose
			config.Verbose = tt.verbose
			defer func() { config.Verbose = oldVerbose }()

			result := GetToolArg(tt.toolName, tt.input)
			if result != tt.expected {
				t.Errorf("GetToolArg(%q, %v) = %q, expected %q", tt.toolName, tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetFilePath(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "Read with file_path",
			toolName: "Read",
			input:    map[string]interface{}{"file_path": "/home/user/file.go"},
			expected: "/home/user/file.go",
		},
		{
			name:     "Write with file_path",
			toolName: "Write",
			input:    map[string]interface{}{"file_path": "/path/to/write.txt"},
			expected: "/path/to/write.txt",
		},
		{
			name:     "Edit with file_path",
			toolName: "Edit",
			input:    map[string]interface{}{"file_path": "/edit/this.js"},
			expected: "/edit/this.js",
		},
		{
			name:     "NotebookEdit with notebook_path",
			toolName: "NotebookEdit",
			input:    map[string]interface{}{"notebook_path": "/notebooks/analysis.ipynb"},
			expected: "/notebooks/analysis.ipynb",
		},
		{
			name:     "NotebookEdit with file_path fallback",
			toolName: "NotebookEdit",
			input:    map[string]interface{}{"file_path": "/notebooks/other.ipynb"},
			expected: "/notebooks/other.ipynb",
		},
		{
			name:     "Read without file_path",
			toolName: "Read",
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "Bash does not return file_path",
			toolName: "Bash",
			input:    map[string]interface{}{"file_path": "/some/path"},
			expected: "",
		},
		{
			name:     "Grep does not return file_path",
			toolName: "Grep",
			input:    map[string]interface{}{"file_path": "/some/path"},
			expected: "",
		},
		{
			name:     "unknown tool returns empty",
			toolName: "CustomTool",
			input:    map[string]interface{}{"file_path": "/some/path"},
			expected: "",
		},
		{
			name:     "nil input",
			toolName: "Read",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilePath(tt.toolName, tt.input)
			if result != tt.expected {
				t.Errorf("GetFilePath(%q, %v) = %q, expected %q", tt.toolName, tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsFilePathTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		expected bool
	}{
		{name: "Read is file path tool", toolName: "Read", expected: true},
		{name: "Write is file path tool", toolName: "Write", expected: true},
		{name: "Edit is file path tool", toolName: "Edit", expected: true},
		{name: "NotebookEdit is file path tool", toolName: "NotebookEdit", expected: true},
		{name: "Bash is not file path tool", toolName: "Bash", expected: false},
		{name: "Glob is not file path tool", toolName: "Glob", expected: false},
		{name: "Grep is not file path tool", toolName: "Grep", expected: false},
		{name: "Task is not file path tool", toolName: "Task", expected: false},
		{name: "WebFetch is not file path tool", toolName: "WebFetch", expected: false},
		{name: "unknown is not file path tool", toolName: "Unknown", expected: false},
		{name: "empty is not file path tool", toolName: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFilePathTool(tt.toolName)
			if result != tt.expected {
				t.Errorf("IsFilePathTool(%q) = %v, expected %v", tt.toolName, result, tt.expected)
			}
		})
	}
}

func TestRegisterAndRegisterFunc(t *testing.T) {
	// Test registering a custom renderer
	customName := "TestCustomTool"
	customResult := "custom-result"

	// Ensure it doesn't exist first
	if r := GetRenderer(customName); r != nil {
		t.Fatalf("Expected %q to not exist in registry", customName)
	}

	// Register using RegisterFunc
	RegisterFunc(customName, func(input map[string]interface{}) string {
		return customResult
	})

	// Verify it's registered
	r := GetRenderer(customName)
	if r == nil {
		t.Fatalf("Expected %q to exist in registry after registration", customName)
	}

	result := r.RenderHeader(nil)
	if result != customResult {
		t.Errorf("Custom renderer returned %q, expected %q", result, customResult)
	}

	// Test overwriting with Register
	newResult := "new-result"
	Register(customName, ToolRendererFunc(func(input map[string]interface{}) string {
		return newResult
	}))

	r = GetRenderer(customName)
	result = r.RenderHeader(nil)
	if result != newResult {
		t.Errorf("Overwritten renderer returned %q, expected %q", result, newResult)
	}

	// Clean up - remove from registry to avoid affecting other tests
	delete(registry, customName)
}

func TestToolRendererFuncInterface(t *testing.T) {
	// Test that ToolRendererFunc properly implements ToolRenderer interface
	var renderer ToolRenderer = ToolRendererFunc(func(input map[string]interface{}) string {
		if v, ok := input["test"].(string); ok {
			return v
		}
		return "default"
	})

	// Test with valid input
	result := renderer.RenderHeader(map[string]interface{}{"test": "value"})
	if result != "value" {
		t.Errorf("Expected 'value', got %q", result)
	}

	// Test with nil input
	result = renderer.RenderHeader(nil)
	if result != "default" {
		t.Errorf("Expected 'default', got %q", result)
	}
}
