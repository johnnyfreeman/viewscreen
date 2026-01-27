package tools

import (
	"testing"
)

// mockConfigProvider implements config.Provider for testing
type mockConfigProvider struct {
	verbose   bool
	noColor   bool
	showUsage bool
}

func (m mockConfigProvider) IsVerbose() bool { return m.verbose }
func (m mockConfigProvider) NoColor() bool   { return m.noColor }
func (m mockConfigProvider) ShowUsage() bool { return m.showUsage }

func TestGetDefinition(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		wantOK   bool
	}{
		{name: "Bash definition exists", toolName: "Bash", wantOK: true},
		{name: "Read definition exists", toolName: "Read", wantOK: true},
		{name: "Write definition exists", toolName: "Write", wantOK: true},
		{name: "Edit definition exists", toolName: "Edit", wantOK: true},
		{name: "Glob definition exists", toolName: "Glob", wantOK: true},
		{name: "Grep definition exists", toolName: "Grep", wantOK: true},
		{name: "Task definition exists", toolName: "Task", wantOK: true},
		{name: "WebFetch definition exists", toolName: "WebFetch", wantOK: true},
		{name: "WebSearch definition exists", toolName: "WebSearch", wantOK: true},
		{name: "TodoWrite definition exists", toolName: "TodoWrite", wantOK: true},
		{name: "AskUserQuestion definition exists", toolName: "AskUserQuestion", wantOK: true},
		{name: "Skill definition exists", toolName: "Skill", wantOK: true},
		{name: "NotebookEdit definition exists", toolName: "NotebookEdit", wantOK: true},
		{name: "TaskOutput definition exists", toolName: "TaskOutput", wantOK: true},
		{name: "TaskStop definition exists", toolName: "TaskStop", wantOK: true},
		{name: "EnterPlanMode definition exists", toolName: "EnterPlanMode", wantOK: true},
		{name: "ExitPlanMode definition exists", toolName: "ExitPlanMode", wantOK: true},
		{name: "ToolSearch definition exists", toolName: "ToolSearch", wantOK: true},
		{name: "unknown tool not found", toolName: "UnknownTool", wantOK: false},
		{name: "empty string not found", toolName: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := GetDefinition(tt.toolName)
			if ok != tt.wantOK {
				t.Errorf("GetDefinition(%q) ok = %v, want %v", tt.toolName, ok, tt.wantOK)
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
		// AskUserQuestion
		{
			name:     "AskUserQuestion with single question",
			toolName: "AskUserQuestion",
			input: map[string]interface{}{
				"questions": []interface{}{
					map[string]interface{}{"question": "What is your name?"},
				},
			},
			expected: "1 question",
		},
		{
			name:     "AskUserQuestion with multiple questions",
			toolName: "AskUserQuestion",
			input: map[string]interface{}{
				"questions": []interface{}{
					map[string]interface{}{"question": "Q1"},
					map[string]interface{}{"question": "Q2"},
					map[string]interface{}{"question": "Q3"},
				},
			},
			expected: "3 questions",
		},
		{
			name:     "AskUserQuestion without questions",
			toolName: "AskUserQuestion",
			input:    map[string]interface{}{},
			expected: "",
		},
		// Skill
		{
			name:     "Skill with skill name",
			toolName: "Skill",
			input:    map[string]interface{}{"skill": "commit"},
			expected: "commit",
		},
		// NotebookEdit
		{
			name:     "NotebookEdit with notebook_path",
			toolName: "NotebookEdit",
			input:    map[string]interface{}{"notebook_path": "/path/to/notebook.ipynb"},
			expected: "/path/to/notebook.ipynb",
		},
		// TaskOutput
		{
			name:     "TaskOutput with task_id",
			toolName: "TaskOutput",
			input:    map[string]interface{}{"task_id": "abc123"},
			expected: "abc123",
		},
		// TaskStop
		{
			name:     "TaskStop with task_id",
			toolName: "TaskStop",
			input:    map[string]interface{}{"task_id": "xyz789"},
			expected: "xyz789",
		},
		// EnterPlanMode
		{
			name:     "EnterPlanMode always empty",
			toolName: "EnterPlanMode",
			input:    map[string]interface{}{},
			expected: "",
		},
		// ExitPlanMode
		{
			name:     "ExitPlanMode always empty",
			toolName: "ExitPlanMode",
			input:    map[string]interface{}{},
			expected: "",
		},
		// ToolSearch
		{
			name:     "ToolSearch with query",
			toolName: "ToolSearch",
			input:    map[string]interface{}{"query": "weather"},
			expected: "weather",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, ok := GetDefinition(tt.toolName)
			if !ok {
				t.Fatalf("GetDefinition(%q) not found", tt.toolName)
			}
			result := def.RenderHeader(tt.input)
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
			cfg := mockConfigProvider{verbose: tt.verbose}
			result := GetToolArgWithConfig(tt.toolName, tt.input, cfg)
			if result != tt.expected {
				t.Errorf("GetToolArgWithConfig(%q, %v) = %q, expected %q", tt.toolName, tt.input, result, tt.expected)
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


func TestToolDefinition_RenderHeader(t *testing.T) {
	tests := []struct {
		name     string
		def      ToolDefinition
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "field extractor with value",
			def:      ToolDefinition{Name: "Test", HeaderField: "cmd"},
			input:    map[string]interface{}{"cmd": "echo hello"},
			expected: "echo hello",
		},
		{
			name:     "field extractor without value",
			def:      ToolDefinition{Name: "Test", HeaderField: "cmd"},
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "field extractor with non-string value",
			def:      ToolDefinition{Name: "Test", HeaderField: "cmd"},
			input:    map[string]interface{}{"cmd": 123},
			expected: "",
		},
		{
			name:     "count field with single item",
			def:      ToolDefinition{Name: "Test", CountField: "items", Singular: "item", Plural: "items"},
			input:    map[string]interface{}{"items": []interface{}{"a"}},
			expected: "1 item",
		},
		{
			name:     "count field with multiple items",
			def:      ToolDefinition{Name: "Test", CountField: "items", Singular: "item", Plural: "items"},
			input:    map[string]interface{}{"items": []interface{}{"a", "b", "c"}},
			expected: "3 items",
		},
		{
			name:     "count field with empty array",
			def:      ToolDefinition{Name: "Test", CountField: "items", Singular: "item", Plural: "items"},
			input:    map[string]interface{}{"items": []interface{}{}},
			expected: "0 items",
		},
		{
			name:     "count field missing",
			def:      ToolDefinition{Name: "Test", CountField: "items", Singular: "item", Plural: "items"},
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "no-op definition",
			def:      ToolDefinition{Name: "Test"},
			input:    map[string]interface{}{"anything": "value"},
			expected: "",
		},
		{
			name:     "nil input",
			def:      ToolDefinition{Name: "Test", HeaderField: "cmd"},
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.def.RenderHeader(tt.input)
			if result != tt.expected {
				t.Errorf("RenderHeader() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToolDefinition_GetFilePath(t *testing.T) {
	tests := []struct {
		name     string
		def      ToolDefinition
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "with file_path field",
			def:      ToolDefinition{Name: "Test", FilePathField: "file_path"},
			input:    map[string]interface{}{"file_path": "/path/to/file"},
			expected: "/path/to/file",
		},
		{
			name:     "with file_path missing",
			def:      ToolDefinition{Name: "Test", FilePathField: "file_path"},
			input:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "with fallback used",
			def:      ToolDefinition{Name: "Test", FilePathField: "notebook_path", FilePathFallback: "file_path"},
			input:    map[string]interface{}{"file_path": "/fallback/path"},
			expected: "/fallback/path",
		},
		{
			name:     "primary takes precedence over fallback",
			def:      ToolDefinition{Name: "Test", FilePathField: "notebook_path", FilePathFallback: "file_path"},
			input:    map[string]interface{}{"notebook_path": "/primary", "file_path": "/fallback"},
			expected: "/primary",
		},
		{
			name:     "no file path field defined",
			def:      ToolDefinition{Name: "Test"},
			input:    map[string]interface{}{"file_path": "/path"},
			expected: "",
		},
		{
			name:     "nil input",
			def:      ToolDefinition{Name: "Test", FilePathField: "file_path"},
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.def.GetFilePath(tt.input)
			if result != tt.expected {
				t.Errorf("GetFilePath() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToolDefinition_IsFilePathTool(t *testing.T) {
	tests := []struct {
		name     string
		def      ToolDefinition
		expected bool
	}{
		{
			name:     "with file path field",
			def:      ToolDefinition{Name: "Read", FilePathField: "file_path"},
			expected: true,
		},
		{
			name:     "without file path field",
			def:      ToolDefinition{Name: "Bash", HeaderField: "command"},
			expected: false,
		},
		{
			name:     "empty definition",
			def:      ToolDefinition{Name: "Test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.def.IsFilePathTool()
			if result != tt.expected {
				t.Errorf("IsFilePathTool() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

