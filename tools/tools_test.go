package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
)

func init() {
	// Initialize styles with no color for predictable test output
	style.Init(true)
}

func TestRenderToolHeaderToString(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		input          map[string]any
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:     "Bash with command",
			toolName: "Bash",
			input:    map[string]any{"command": "ls -la"},
			wantContains: []string{
				style.Bullet,
				"Bash",
				"ls -la",
			},
		},
		{
			name:     "Read with file path",
			toolName: "Read",
			input:    map[string]any{"file_path": "/path/to/file.go"},
			wantContains: []string{
				style.Bullet,
				"Read",
				"/path/to/file.go",
			},
		},
		{
			name:     "Write with file path",
			toolName: "Write",
			input:    map[string]any{"file_path": "/path/to/new.go"},
			wantContains: []string{
				style.Bullet,
				"Write",
				"/path/to/new.go",
			},
		},
		{
			name:     "Edit with file path",
			toolName: "Edit",
			input:    map[string]any{"file_path": "/path/to/edit.go"},
			wantContains: []string{
				style.Bullet,
				"Edit",
				"/path/to/edit.go",
			},
		},
		{
			name:     "Glob with pattern",
			toolName: "Glob",
			input:    map[string]any{"pattern": "**/*.go"},
			wantContains: []string{
				style.Bullet,
				"Glob",
				"**/*.go",
			},
		},
		{
			name:     "Grep with pattern",
			toolName: "Grep",
			input:    map[string]any{"pattern": "TODO:"},
			wantContains: []string{
				style.Bullet,
				"Grep",
				"TODO:",
			},
		},
		{
			name:     "Task with description",
			toolName: "Task",
			input:    map[string]any{"description": "Explore codebase"},
			wantContains: []string{
				style.Bullet,
				"Task",
				"Explore codebase",
			},
		},
		{
			name:     "WebFetch with url",
			toolName: "WebFetch",
			input:    map[string]any{"url": "https://example.com"},
			wantContains: []string{
				style.Bullet,
				"WebFetch",
				"https://example.com",
			},
		},
		{
			name:     "WebSearch with query",
			toolName: "WebSearch",
			input:    map[string]any{"query": "golang testing"},
			wantContains: []string{
				style.Bullet,
				"WebSearch",
				"golang testing",
			},
		},
		{
			name:     "TodoWrite with todos",
			toolName: "TodoWrite",
			input: map[string]any{
				"todos": []any{
					map[string]any{"content": "task1"},
					map[string]any{"content": "task2"},
				},
			},
			wantContains: []string{
				style.Bullet,
				"TodoWrite",
				"2 items",
			},
		},
		{
			name:     "AskUserQuestion with questions",
			toolName: "AskUserQuestion",
			input: map[string]any{
				"questions": []any{
					map[string]any{"question": "What should I do?"},
					map[string]any{"question": "What else?"},
				},
			},
			wantContains: []string{
				style.Bullet,
				"AskUserQuestion",
				"2 questions",
			},
		},
		{
			name:     "empty input",
			toolName: "Bash",
			input:    map[string]any{},
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
			input:    map[string]any{"key": "value"},
			wantContains: []string{
				style.Bullet,
				"CustomTool",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := RenderToolHeaderToString(tt.toolName, tt.input)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("RenderToolHeaderToString() output missing %q\nGot: %q", want, output)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("RenderToolHeaderToString() output should not contain %q\nGot: %q", notWant, output)
				}
			}

			// Verify output ends with newline
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("RenderToolHeaderToString() output should end with newline, got: %q", output)
			}
		})
	}
}

func TestRenderToolHeaderTruncation(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		input        map[string]any
		wantTrunc    bool
		wantContains string
	}{
		{
			name:      "short command not truncated",
			toolName:  "Bash",
			input:     map[string]any{"command": "ls -la"},
			wantTrunc: false,
		},
		{
			name:     "long command truncated",
			toolName: "Bash",
			input: map[string]any{
				"command": "this is a very long command that exceeds eighty characters and should be truncated with ellipsis at the end",
			},
			wantTrunc:    true,
			wantContains: "...",
		},
		{
			name:     "long file path truncated",
			toolName: "Read",
			input: map[string]any{
				"file_path": "/home/user/some/very/deeply/nested/directory/structure/with/many/levels/that/exceeds/eighty/characters/file.go",
			},
			wantTrunc:    true,
			wantContains: "...",
		},
		{
			name:     "exactly 80 chars not truncated",
			toolName: "Bash",
			input: map[string]any{
				"command": strings.Repeat("a", 80),
			},
			wantTrunc: false,
		},
		{
			name:     "81 chars truncated",
			toolName: "Bash",
			input: map[string]any{
				"command": strings.Repeat("a", 81),
			},
			wantTrunc:    true,
			wantContains: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := RenderToolHeaderToString(tt.toolName, tt.input)

			hasTrunc := strings.Contains(output, "...")
			if hasTrunc != tt.wantTrunc {
				t.Errorf("RenderToolHeaderToString() truncation = %v, want %v\nOutput: %q", hasTrunc, tt.wantTrunc, output)
			}

			if tt.wantContains != "" && !strings.Contains(output, tt.wantContains) {
				t.Errorf("RenderToolHeaderToString() missing %q\nOutput: %q", tt.wantContains, output)
			}
		})
	}
}

func TestRenderToolUseToString(t *testing.T) {
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
				Name:  "EnterPlanMode",
				Input: json.RawMessage(`{}`),
			},
			wantContains: []string{
				style.Bullet,
				"EnterPlanMode",
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
			output, _ := RenderToolUseToString(tt.block)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("RenderToolUseToString() output missing %q\nGot: %q", want, output)
				}
			}

			// Verify output ends with newline
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("RenderToolUseToString() output should end with newline, got: %q", output)
			}
		})
	}
}

func TestRenderNestedToolUseToString(t *testing.T) {
	tests := []struct {
		name         string
		block        types.ContentBlock
		wantContains []string
	}{
		{
			name: "nested tool with valid JSON input",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Grep",
				Input: json.RawMessage(`{"pattern": "TODO"}`),
			},
			wantContains: []string{
				style.NestedPrefix,
				style.Bullet,
				"Grep",
				"TODO",
			},
		},
		{
			name: "nested tool with nil input",
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "SomeTool",
				Input: nil,
			},
			wantContains: []string{
				style.NestedPrefix,
				style.Bullet,
				"SomeTool",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := RenderNestedToolUseToString(tt.block)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("RenderNestedToolUseToString() output missing %q\nGot: %q", want, output)
				}
			}

			if !strings.HasSuffix(output, "\n") {
				t.Errorf("RenderNestedToolUseToString() output should end with newline, got: %q", output)
			}
		})
	}
}

func TestFunctionSignatures(t *testing.T) {
	// Verify that the function signatures match what callers expect
	t.Run("RenderToolUse signature", func(t *testing.T) {
		_ = RenderToolUse // Compile-time check
	})

	t.Run("RenderToolHeader signature", func(t *testing.T) {
		_ = RenderToolHeader // Compile-time check
	})

	t.Run("RenderNestedToolUse signature", func(t *testing.T) {
		_ = RenderNestedToolUse // Compile-time check
	})
}

func TestRenderHeaderTo(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		input        map[string]any
		opts         HeaderOptions
		wantContains []string
	}{
		{
			name:     "default options uses bullet",
			toolName: "Bash",
			input:    map[string]any{"command": "echo hi"},
			opts:     DefaultHeaderOptions(),
			wantContains: []string{
				style.Bullet,
				"Bash",
				"echo hi",
			},
		},
		{
			name:     "custom icon replaces bullet",
			toolName: "Read",
			input:    map[string]any{"file_path": "/path/to/file"},
			opts: HeaderOptions{
				Icon: "◐ ",
			},
			wantContains: []string{
				"◐ ",
				"Read",
				"/path/to/file",
			},
		},
		{
			name:     "prefix is prepended",
			toolName: "Grep",
			input:    map[string]any{"pattern": "TODO"},
			opts: HeaderOptions{
				Prefix: style.NestedPrefix,
			},
			wantContains: []string{
				style.NestedPrefix,
				style.Bullet,
				"Grep",
				"TODO",
			},
		},
		{
			name:     "custom icon with prefix",
			toolName: "Glob",
			input:    map[string]any{"pattern": "**/*.go"},
			opts: HeaderOptions{
				Icon:   "⏳ ",
				Prefix: ">> ",
			},
			wantContains: []string{
				">> ",
				"⏳ ",
				"Glob",
				"**/*.go",
			},
		},
		{
			name:     "empty icon defaults to bullet",
			toolName: "Task",
			input:    map[string]any{"description": "test"},
			opts: HeaderOptions{
				Icon: "",
			},
			wantContains: []string{
				style.Bullet,
				"Task",
				"test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := render.StringOutput()
			RenderHeaderTo(out, tt.toolName, tt.input, tt.opts)
			result := out.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("RenderHeaderTo() output missing %q\nGot: %q", want, result)
				}
			}

			if !strings.HasSuffix(result, "\n") {
				t.Errorf("RenderHeaderTo() output should end with newline, got: %q", result)
			}
		})
	}
}

func TestDefaultHeaderOptions(t *testing.T) {
	opts := DefaultHeaderOptions()

	if opts.Icon != style.Bullet {
		t.Errorf("DefaultHeaderOptions().Icon = %q, want %q", opts.Icon, style.Bullet)
	}

	if opts.Prefix != "" {
		t.Errorf("DefaultHeaderOptions().Prefix = %q, want empty string", opts.Prefix)
	}
}
