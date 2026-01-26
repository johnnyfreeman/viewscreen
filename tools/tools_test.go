package tools

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
)

func init() {
	// Initialize styles with no color for predictable test output
	style.Init(true)
}

func TestRenderToolHeader(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		input          map[string]interface{}
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:     "Bash with command",
			toolName: "Bash",
			input:    map[string]interface{}{"command": "ls -la"},
			wantContains: []string{
				style.Bullet,
				"Bash",
				"ls -la",
			},
		},
		{
			name:     "Read with file path",
			toolName: "Read",
			input:    map[string]interface{}{"file_path": "/path/to/file.go"},
			wantContains: []string{
				style.Bullet,
				"Read",
				"/path/to/file.go",
			},
		},
		{
			name:     "Write with file path",
			toolName: "Write",
			input:    map[string]interface{}{"file_path": "/path/to/new.go"},
			wantContains: []string{
				style.Bullet,
				"Write",
				"/path/to/new.go",
			},
		},
		{
			name:     "Edit with file path",
			toolName: "Edit",
			input:    map[string]interface{}{"file_path": "/path/to/edit.go"},
			wantContains: []string{
				style.Bullet,
				"Edit",
				"/path/to/edit.go",
			},
		},
		{
			name:     "Glob with pattern",
			toolName: "Glob",
			input:    map[string]interface{}{"pattern": "**/*.go"},
			wantContains: []string{
				style.Bullet,
				"Glob",
				"**/*.go",
			},
		},
		{
			name:     "Grep with pattern",
			toolName: "Grep",
			input:    map[string]interface{}{"pattern": "TODO:"},
			wantContains: []string{
				style.Bullet,
				"Grep",
				"TODO:",
			},
		},
		{
			name:     "Task with description",
			toolName: "Task",
			input:    map[string]interface{}{"description": "Explore codebase"},
			wantContains: []string{
				style.Bullet,
				"Task",
				"Explore codebase",
			},
		},
		{
			name:     "WebFetch with url",
			toolName: "WebFetch",
			input:    map[string]interface{}{"url": "https://example.com"},
			wantContains: []string{
				style.Bullet,
				"WebFetch",
				"https://example.com",
			},
		},
		{
			name:     "WebSearch with query",
			toolName: "WebSearch",
			input:    map[string]interface{}{"query": "golang testing"},
			wantContains: []string{
				style.Bullet,
				"WebSearch",
				"golang testing",
			},
		},
		{
			name:     "TodoWrite with todos",
			toolName: "TodoWrite",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{"content": "task1"},
					map[string]interface{}{"content": "task2"},
				},
			},
			wantContains: []string{
				style.Bullet,
				"TodoWrite",
				"2 items",
			},
		},
		{
			name:     "AskUser with question",
			toolName: "AskUser",
			input:    map[string]interface{}{"question": "What should I do?"},
			wantContains: []string{
				style.Bullet,
				"AskUser",
				"What should I do?",
			},
		},
		{
			name:     "empty input",
			toolName: "Bash",
			input:    map[string]interface{}{},
			wantContains: []string{
				style.Bullet,
				"Bash",
			},
		},
		{
			name:     "nil input",
			toolName: "Read",
			input:    nil,
			wantContains: []string{
				style.Bullet,
				"Read",
			},
		},
		{
			name:     "unknown tool",
			toolName: "CustomTool",
			input:    map[string]interface{}{"key": "value"},
			wantContains: []string{
				style.Bullet,
				"CustomTool",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderToolHeader(tt.toolName, tt.input, WithOutput(&buf))

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("RenderToolHeader() output missing %q\nGot: %q", want, output)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("RenderToolHeader() output should not contain %q\nGot: %q", notWant, output)
				}
			}

			// Verify output ends with newline
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("RenderToolHeader() output should end with newline, got: %q", output)
			}
		})
	}
}

func TestRenderToolHeaderTruncation(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		input        map[string]interface{}
		wantTrunc    bool
		wantContains string
	}{
		{
			name:      "short command not truncated",
			toolName:  "Bash",
			input:     map[string]interface{}{"command": "ls -la"},
			wantTrunc: false,
		},
		{
			name:     "long command truncated",
			toolName: "Bash",
			input: map[string]interface{}{
				"command": "this is a very long command that exceeds eighty characters and should be truncated with ellipsis at the end",
			},
			wantTrunc:    true,
			wantContains: "...",
		},
		{
			name:     "long file path truncated",
			toolName: "Read",
			input: map[string]interface{}{
				"file_path": "/home/user/some/very/deeply/nested/directory/structure/with/many/levels/that/exceeds/eighty/characters/file.go",
			},
			wantTrunc:    true,
			wantContains: "...",
		},
		{
			name:     "exactly 80 chars not truncated",
			toolName: "Bash",
			input: map[string]interface{}{
				"command": strings.Repeat("a", 80),
			},
			wantTrunc: false,
		},
		{
			name:     "81 chars truncated",
			toolName: "Bash",
			input: map[string]interface{}{
				"command": strings.Repeat("a", 81),
			},
			wantTrunc:    true,
			wantContains: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderToolHeader(tt.toolName, tt.input, WithOutput(&buf))

			output := buf.String()

			hasTrunc := strings.Contains(output, "...")
			if hasTrunc != tt.wantTrunc {
				t.Errorf("RenderToolHeader() truncation = %v, want %v\nOutput: %q", hasTrunc, tt.wantTrunc, output)
			}

			if tt.wantContains != "" && !strings.Contains(output, tt.wantContains) {
				t.Errorf("RenderToolHeader() missing %q\nOutput: %q", tt.wantContains, output)
			}
		})
	}
}

func TestRenderToolUse(t *testing.T) {
	tests := []struct {
		name         string
		block        types.ContentBlock
		wantContains []string
	}{
		{
			name: "tool with valid JSON input",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Bash",
				Input: json.RawMessage(`{"command": "echo hello"}`),
			},
			wantContains: []string{
				style.Bullet,
				"Bash",
				"echo hello",
			},
		},
		{
			name: "tool with empty input",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "TodoRead",
				Input: json.RawMessage(`{}`),
			},
			wantContains: []string{
				style.Bullet,
				"TodoRead",
			},
		},
		{
			name: "tool with nil input",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "SomeTool",
				Input: nil,
			},
			wantContains: []string{
				style.Bullet,
				"SomeTool",
			},
		},
		{
			name: "tool with invalid JSON input",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Bash",
				Input: json.RawMessage(`{invalid json`),
			},
			wantContains: []string{
				style.Bullet,
				"Bash",
			},
		},
		{
			name: "Read tool with file path",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Read",
				Input: json.RawMessage(`{"file_path": "/path/to/file.go"}`),
			},
			wantContains: []string{
				style.Bullet,
				"Read",
				"/path/to/file.go",
			},
		},
		{
			name: "Edit tool with file path",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Edit",
				Input: json.RawMessage(`{"file_path": "/path/to/edit.go", "old_string": "foo", "new_string": "bar"}`),
			},
			wantContains: []string{
				style.Bullet,
				"Edit",
				"/path/to/edit.go",
			},
		},
		{
			name: "Glob tool with pattern",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Glob",
				Input: json.RawMessage(`{"pattern": "**/*.ts"}`),
			},
			wantContains: []string{
				style.Bullet,
				"Glob",
				"**/*.ts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			RenderToolUse(tt.block, WithOutput(&buf))

			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("RenderToolUse() output missing %q\nGot: %q", want, output)
				}
			}

			// Verify output ends with newline
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("RenderToolUse() output should end with newline, got: %q", output)
			}
		})
	}
}

func TestRenderToolUseDefault(t *testing.T) {
	// Test that the default wrapper function exists and has correct signature
	// We verify the function type matches what's expected by callers
	var fn func(types.ContentBlock) = RenderToolUseDefault
	if fn == nil {
		t.Error("RenderToolUseDefault should not be nil")
	}
}

func TestRenderToolHeaderDefault(t *testing.T) {
	// Test that the default wrapper function exists and has correct signature
	// We verify the function type matches what's expected by callers
	var fn func(string, map[string]interface{}) = RenderToolHeaderDefault
	if fn == nil {
		t.Error("RenderToolHeaderDefault should not be nil")
	}
}

func TestWithOutput(t *testing.T) {
	// Test that WithOutput properly sets the writer
	var buf bytes.Buffer
	cfg := newRenderConfig(WithOutput(&buf))

	if cfg.output != &buf {
		t.Errorf("WithOutput() did not set the output writer correctly")
	}
}

func TestNewRenderConfig(t *testing.T) {
	tests := []struct {
		name       string
		opts       []ToolRenderOption
		wantStdout bool
	}{
		{
			name:       "default config uses stdout",
			opts:       nil,
			wantStdout: true,
		},
		{
			name:       "empty options uses stdout",
			opts:       []ToolRenderOption{},
			wantStdout: true,
		},
		{
			name: "custom output overrides stdout",
			opts: []ToolRenderOption{
				WithOutput(&bytes.Buffer{}),
			},
			wantStdout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newRenderConfig(tt.opts...)

			if cfg.output == nil {
				t.Error("newRenderConfig() returned nil output")
			}

			// Can't easily compare to os.Stdout, but we can verify it's not nil
			// and that custom output is different
			if !tt.wantStdout {
				// If we don't want stdout, the output should be the buffer we passed
				if _, ok := cfg.output.(*bytes.Buffer); !ok {
					t.Error("newRenderConfig() did not use custom output")
				}
			}
		})
	}
}

func TestToolRenderConfigChaining(t *testing.T) {
	// Test that multiple options are applied in order
	var buf1, buf2 bytes.Buffer

	cfg := newRenderConfig(
		WithOutput(&buf1),
		WithOutput(&buf2), // Should override buf1
	)

	if cfg.output != &buf2 {
		t.Error("Multiple WithOutput options should apply in order, last one wins")
	}
}
