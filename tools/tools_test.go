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

func TestHeaderRenderer_Truncation(t *testing.T) {
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
			output, _ := NewHeaderRenderer().RenderToString(tt.toolName, tt.input)

			hasTrunc := strings.Contains(output, "...")
			if hasTrunc != tt.wantTrunc {
				t.Errorf("RenderToString() truncation = %v, want %v\nOutput: %q", hasTrunc, tt.wantTrunc, output)
			}

			if tt.wantContains != "" && !strings.Contains(output, tt.wantContains) {
				t.Errorf("RenderToString() missing %q\nOutput: %q", tt.wantContains, output)
			}
		})
	}
}

func TestHeaderRenderer_RenderBlockToStringWithNesting(t *testing.T) {
	block := types.ContentBlock{
		Type:  "tool_use",
		Name:  "Read",
		Input: json.RawMessage(`{"file_path": "/test/file.go"}`),
	}

	t.Run("not nested", func(t *testing.T) {
		output, ctx := NewHeaderRenderer().RenderBlockToStringWithNesting(block, false)

		if strings.Contains(output, style.NestedPrefix) {
			t.Errorf("expected no nested prefix, got: %q", output)
		}
		if !strings.Contains(output, "Read") {
			t.Errorf("expected tool name in output, got: %q", output)
		}
		if ctx.ToolName != "Read" {
			t.Errorf("expected ToolName=Read, got: %q", ctx.ToolName)
		}
	})

	t.Run("nested", func(t *testing.T) {
		output, ctx := NewHeaderRenderer().RenderBlockToStringWithNesting(block, true)

		if !strings.Contains(output, style.NestedPrefix) {
			t.Errorf("expected nested prefix, got: %q", output)
		}
		if !strings.Contains(output, "Read") {
			t.Errorf("expected tool name in output, got: %q", output)
		}
		if ctx.ToolName != "Read" {
			t.Errorf("expected ToolName=Read, got: %q", ctx.ToolName)
		}
	})
}

func TestHeaderRenderer_renderTo(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		input        map[string]any
		opts         []HeaderRendererOption
		wantContains []string
	}{
		{
			name:     "default options uses bullet",
			toolName: "Bash",
			input:    map[string]any{"command": "echo hi"},
			opts:     nil,
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
			opts:     []HeaderRendererOption{WithIcon("◐ ")},
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
			opts:     []HeaderRendererOption{WithPrefix(style.NestedPrefix)},
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
			opts:     []HeaderRendererOption{WithIcon("⏳ "), WithPrefix(">> ")},
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
			opts:     []HeaderRendererOption{WithIcon("")},
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
			r := NewHeaderRenderer(tt.opts...)
			r.renderTo(out, tt.toolName, tt.input)
			result := out.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("renderTo() output missing %q\nGot: %q", want, result)
				}
			}

			if !strings.HasSuffix(result, "\n") {
				t.Errorf("renderTo() output should end with newline, got: %q", result)
			}
		})
	}
}

func TestHeaderRenderer_ToolArgVariants(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		input        map[string]any
		wantContains []string
	}{
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
			output, _ := NewHeaderRenderer().RenderToString(tt.toolName, tt.input)

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("RenderToString() output missing %q\nGot: %q", want, output)
				}
			}

			if !strings.HasSuffix(output, "\n") {
				t.Errorf("RenderToString() output should end with newline, got: %q", output)
			}
		})
	}
}
