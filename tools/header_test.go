package tools

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
)

func TestHeaderRenderer_RenderToString(t *testing.T) {
	tests := []struct {
		name         string
		opts         []HeaderRendererOption
		toolName     string
		input        map[string]any
		wantContains []string
	}{
		{
			name:     "default renderer",
			opts:     nil,
			toolName: "Bash",
			input:    map[string]any{"command": "ls -la"},
			wantContains: []string{
				style.Bullet,
				"Bash",
				"ls -la",
			},
		},
		{
			name:     "with custom icon",
			opts:     []HeaderRendererOption{WithIcon("◐ ")},
			toolName: "Read",
			input:    map[string]any{"file_path": "/path/to/file.go"},
			wantContains: []string{
				"◐ ",
				"Read",
				"/path/to/file.go",
			},
		},
		{
			name:     "with nested prefix",
			opts:     []HeaderRendererOption{WithNested()},
			toolName: "Grep",
			input:    map[string]any{"pattern": "TODO"},
			wantContains: []string{
				style.NestedPrefix,
				style.Bullet,
				"Grep",
				"TODO",
			},
		},
		{
			name:     "with custom prefix",
			opts:     []HeaderRendererOption{WithPrefix(">> ")},
			toolName: "Glob",
			input:    map[string]any{"pattern": "**/*.go"},
			wantContains: []string{
				">> ",
				style.Bullet,
				"Glob",
				"**/*.go",
			},
		},
		{
			name:     "with icon and nested",
			opts:     []HeaderRendererOption{WithIcon("⏳ "), WithNested()},
			toolName: "Task",
			input:    map[string]any{"description": "test task"},
			wantContains: []string{
				style.NestedPrefix,
				"⏳ ",
				"Task",
				"test task",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewHeaderRenderer(tt.opts...)
			result, ctx := r.RenderToString(tt.toolName, tt.input)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("RenderToString() output missing %q\nGot: %q", want, result)
				}
			}

			if !strings.HasSuffix(result, "\n") {
				t.Errorf("RenderToString() output should end with newline, got: %q", result)
			}

			if ctx.ToolName != tt.toolName {
				t.Errorf("RenderToString() context.ToolName = %q, want %q", ctx.ToolName, tt.toolName)
			}
		})
	}
}

func TestHeaderRenderer_RenderBlockToString(t *testing.T) {
	tests := []struct {
		name         string
		opts         []HeaderRendererOption
		block        types.ContentBlock
		wantContains []string
	}{
		{
			name: "valid JSON input",
			opts: nil,
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
			name: "nested with custom icon",
			opts: []HeaderRendererOption{WithIcon("◑ "), WithNested()},
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Read",
				Input: json.RawMessage(`{"file_path": "/path/to/file.go"}`),
			},
			wantContains: []string{
				style.NestedPrefix,
				"◑ ",
				"Read",
				"/path/to/file.go",
			},
		},
		{
			name: "nil input",
			opts: nil,
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "EnterPlanMode",
				Input: nil,
			},
			wantContains: []string{
				style.Bullet,
				"EnterPlanMode",
			},
		},
		{
			name: "invalid JSON input",
			opts: nil,
			block: types.ContentBlock{
				Type:  "tool_use",
				Name:  "Bash",
				Input: json.RawMessage(`{invalid`),
			},
			wantContains: []string{
				style.Bullet,
				"Bash",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewHeaderRenderer(tt.opts...)
			result, ctx := r.RenderBlockToString(tt.block)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("RenderBlockToString() output missing %q\nGot: %q", want, result)
				}
			}

			if !strings.HasSuffix(result, "\n") {
				t.Errorf("RenderBlockToString() output should end with newline, got: %q", result)
			}

			if ctx.ToolName != tt.block.Name {
				t.Errorf("RenderBlockToString() context.ToolName = %q, want %q", ctx.ToolName, tt.block.Name)
			}
		})
	}
}

func TestHeaderRenderer_WithOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewHeaderRenderer(WithOutput(&buf))
	r.Render("Bash", map[string]any{"command": "ls"})

	result := buf.String()
	if !strings.Contains(result, "Bash") {
		t.Errorf("WithOutput() did not write to custom output, got: %q", result)
	}
	if !strings.Contains(result, "ls") {
		t.Errorf("WithOutput() missing command arg, got: %q", result)
	}
}

func TestHeaderRenderer_ToolContext(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		input        map[string]any
		wantFilePath string
	}{
		{
			name:         "Read extracts file path",
			toolName:     "Read",
			input:        map[string]any{"file_path": "/path/to/file.go"},
			wantFilePath: "/path/to/file.go",
		},
		{
			name:         "Edit extracts file path",
			toolName:     "Edit",
			input:        map[string]any{"file_path": "/path/to/edit.go", "old_string": "foo"},
			wantFilePath: "/path/to/edit.go",
		},
		{
			name:         "NotebookEdit extracts notebook path",
			toolName:     "NotebookEdit",
			input:        map[string]any{"notebook_path": "/path/to/notebook.ipynb"},
			wantFilePath: "/path/to/notebook.ipynb",
		},
		{
			name:         "Bash has no file path",
			toolName:     "Bash",
			input:        map[string]any{"command": "ls"},
			wantFilePath: "",
		},
		{
			name:         "nil input has no file path",
			toolName:     "Read",
			input:        nil,
			wantFilePath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewHeaderRenderer()
			_, ctx := r.RenderToString(tt.toolName, tt.input)

			if ctx.FilePath != tt.wantFilePath {
				t.Errorf("RenderToString() context.FilePath = %q, want %q", ctx.FilePath, tt.wantFilePath)
			}
		})
	}
}

func TestParseBlockInput(t *testing.T) {
	tests := []struct {
		name  string
		block types.ContentBlock
		want  map[string]any
	}{
		{
			name: "valid JSON",
			block: types.ContentBlock{
				Input: json.RawMessage(`{"key": "value"}`),
			},
			want: map[string]any{"key": "value"},
		},
		{
			name: "empty input",
			block: types.ContentBlock{
				Input: json.RawMessage{},
			},
			want: nil,
		},
		{
			name: "nil input",
			block: types.ContentBlock{
				Input: nil,
			},
			want: nil,
		},
		{
			name: "invalid JSON",
			block: types.ContentBlock{
				Input: json.RawMessage(`{invalid`),
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBlockInput(tt.block)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ParseBlockInput() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("ParseBlockInput() = nil, want %v", tt.want)
				}
			}
		})
	}
}
