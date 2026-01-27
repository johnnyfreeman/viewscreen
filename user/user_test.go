package user

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/content"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/tools"
)

func TestToolResultContent_Content(t *testing.T) {
	tests := []struct {
		name     string
		rawJSON  string
		expected string
	}{
		{
			name:     "empty content",
			rawJSON:  "",
			expected: "",
		},
		{
			name:     "simple string",
			rawJSON:  `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "string with newlines",
			rawJSON:  `"line1\nline2\nline3"`,
			expected: "line1\nline2\nline3",
		},
		{
			name:     "string with special characters",
			rawJSON:  `"hello \"world\" with special chars: \t\n"`,
			expected: "hello \"world\" with special chars: \t\n",
		},
		{
			name:     "single text block array",
			rawJSON:  `[{"type": "text", "text": "hello from block"}]`,
			expected: "hello from block",
		},
		{
			name:     "multiple text blocks",
			rawJSON:  `[{"type": "text", "text": "first"}, {"type": "text", "text": "second"}]`,
			expected: "first\nsecond",
		},
		{
			name:     "array with empty text",
			rawJSON:  `[{"type": "text", "text": ""}, {"type": "text", "text": "second"}]`,
			expected: "second",
		},
		{
			name:     "array with non-text types",
			rawJSON:  `[{"type": "image", "data": "base64data"}, {"type": "text", "text": "caption"}]`,
			expected: "caption",
		},
		{
			name:     "array with only non-text types",
			rawJSON:  `[{"type": "image", "data": "base64data"}]`,
			expected: "",
		},
		{
			name:     "empty array",
			rawJSON:  `[]`,
			expected: "",
		},
		{
			name:     "invalid JSON returns raw",
			rawJSON:  `not valid json`,
			expected: "not valid json",
		},
		{
			name:     "number value returns raw",
			rawJSON:  `123`,
			expected: "123",
		},
		{
			name:     "null value returns empty string",
			rawJSON:  `null`,
			expected: "",
		},
		{
			name:     "object value returns raw",
			rawJSON:  `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := ToolResultContent{
				RawContent: json.RawMessage(tt.rawJSON),
			}
			result := content.Content()
			if result != tt.expected {
				t.Errorf("Content() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToolResultContent_ContentWithFields(t *testing.T) {
	// Test that other fields don't affect Content() behavior
	content := ToolResultContent{
		Type:       "tool_result",
		ToolUseID:  "test-123",
		RawContent: json.RawMessage(`"test content"`),
		IsError:    true,
	}

	result := content.Content()
	if result != "test content" {
		t.Errorf("Content() = %q, expected %q", result, "test content")
	}
}

func TestRendererSetToolContext(t *testing.T) {
	r := NewRenderer()

	r.SetToolContext(tools.ToolContext{ToolName: "Read", FilePath: "/path/to/file.go"})

	if r.toolContext.ToolName != "Read" {
		t.Errorf("ToolName = %q, expected %q", r.toolContext.ToolName, "Read")
	}
	if r.toolContext.FilePath != "/path/to/file.go" {
		t.Errorf("FilePath = %q, expected %q", r.toolContext.FilePath, "/path/to/file.go")
	}
}

func TestEditResult_Unmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected EditResult
	}{
		{
			name: "complete edit result",
			jsonData: `{
				"filePath": "/path/to/file.go",
				"oldString": "old code",
				"newString": "new code",
				"structuredPatch": [
					{
						"oldStart": 10,
						"oldLines": 3,
						"newStart": 10,
						"newLines": 5,
						"lines": [" context", "-removed", "+added"]
					}
				]
			}`,
			expected: EditResult{
				FilePath:  "/path/to/file.go",
				OldString: "old code",
				NewString: "new code",
				StructuredPatch: []PatchHunk{
					{
						OldStart: 10,
						OldLines: 3,
						NewStart: 10,
						NewLines: 5,
						Lines:    []string{" context", "-removed", "+added"},
					},
				},
			},
		},
		{
			name:     "minimal edit result",
			jsonData: `{"filePath": "/test.txt"}`,
			expected: EditResult{
				FilePath: "/test.txt",
			},
		},
		{
			name: "edit result with multiple hunks",
			jsonData: `{
				"filePath": "/path/to/file.go",
				"structuredPatch": [
					{"oldStart": 1, "oldLines": 1, "newStart": 1, "newLines": 1, "lines": ["-old"]},
					{"oldStart": 10, "oldLines": 1, "newStart": 10, "newLines": 1, "lines": ["+new"]}
				]
			}`,
			expected: EditResult{
				FilePath: "/path/to/file.go",
				StructuredPatch: []PatchHunk{
					{OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1, Lines: []string{"-old"}},
					{OldStart: 10, OldLines: 1, NewStart: 10, NewLines: 1, Lines: []string{"+new"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result EditResult
			err := json.Unmarshal([]byte(tt.jsonData), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.FilePath != tt.expected.FilePath {
				t.Errorf("FilePath = %q, expected %q", result.FilePath, tt.expected.FilePath)
			}
			if result.OldString != tt.expected.OldString {
				t.Errorf("OldString = %q, expected %q", result.OldString, tt.expected.OldString)
			}
			if result.NewString != tt.expected.NewString {
				t.Errorf("NewString = %q, expected %q", result.NewString, tt.expected.NewString)
			}
			if len(result.StructuredPatch) != len(tt.expected.StructuredPatch) {
				t.Errorf("StructuredPatch length = %d, expected %d", len(result.StructuredPatch), len(tt.expected.StructuredPatch))
			}
			for i, hunk := range result.StructuredPatch {
				expected := tt.expected.StructuredPatch[i]
				if hunk.OldStart != expected.OldStart || hunk.OldLines != expected.OldLines ||
					hunk.NewStart != expected.NewStart || hunk.NewLines != expected.NewLines {
					t.Errorf("Hunk %d mismatch: got %+v, expected %+v", i, hunk, expected)
				}
			}
		})
	}
}

func TestPatchHunk_Unmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected PatchHunk
	}{
		{
			name: "full hunk",
			jsonData: `{
				"oldStart": 5,
				"oldLines": 10,
				"newStart": 7,
				"newLines": 12,
				"lines": [" context", "-removed", "+added", " more context"]
			}`,
			expected: PatchHunk{
				OldStart: 5,
				OldLines: 10,
				NewStart: 7,
				NewLines: 12,
				Lines:    []string{" context", "-removed", "+added", " more context"},
			},
		},
		{
			name:     "empty hunk",
			jsonData: `{}`,
			expected: PatchHunk{},
		},
		{
			name:     "hunk with zero values",
			jsonData: `{"oldStart": 0, "oldLines": 0, "newStart": 0, "newLines": 0, "lines": []}`,
			expected: PatchHunk{
				OldStart: 0,
				OldLines: 0,
				NewStart: 0,
				NewLines: 0,
				Lines:    []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PatchHunk
			err := json.Unmarshal([]byte(tt.jsonData), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.OldStart != tt.expected.OldStart {
				t.Errorf("OldStart = %d, expected %d", result.OldStart, tt.expected.OldStart)
			}
			if result.OldLines != tt.expected.OldLines {
				t.Errorf("OldLines = %d, expected %d", result.OldLines, tt.expected.OldLines)
			}
			if result.NewStart != tt.expected.NewStart {
				t.Errorf("NewStart = %d, expected %d", result.NewStart, tt.expected.NewStart)
			}
			if result.NewLines != tt.expected.NewLines {
				t.Errorf("NewLines = %d, expected %d", result.NewLines, tt.expected.NewLines)
			}
			if len(result.Lines) != len(tt.expected.Lines) {
				t.Errorf("Lines length = %d, expected %d", len(result.Lines), len(tt.expected.Lines))
			}
		})
	}
}

func TestMessage_Unmarshaling(t *testing.T) {
	jsonData := `{
		"role": "user",
		"content": [
			{
				"type": "tool_result",
				"tool_use_id": "tool-123",
				"content": "result content",
				"is_error": false
			}
		]
	}`

	var msg Message
	err := json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if msg.Role != "user" {
		t.Errorf("Role = %q, expected %q", msg.Role, "user")
	}
	if len(msg.Content) != 1 {
		t.Fatalf("Content length = %d, expected 1", len(msg.Content))
	}
	if msg.Content[0].Type != "tool_result" {
		t.Errorf("Content[0].Type = %q, expected %q", msg.Content[0].Type, "tool_result")
	}
	if msg.Content[0].ToolUseID != "tool-123" {
		t.Errorf("Content[0].ToolUseID = %q, expected %q", msg.Content[0].ToolUseID, "tool-123")
	}
}

func TestEvent_Unmarshaling(t *testing.T) {
	jsonData := `{
		"type": "user",
		"message": {
			"role": "user",
			"content": [
				{
					"type": "tool_result",
					"tool_use_id": "tool-456",
					"content": "file contents here",
					"is_error": false
				}
			]
		},
		"tool_use_result": {"filePath": "/test.go", "success": true}
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if event.Type != "user" {
		t.Errorf("Type = %q, expected %q", event.Type, "user")
	}
	if event.Message.Role != "user" {
		t.Errorf("Message.Role = %q, expected %q", event.Message.Role, "user")
	}
	if len(event.Message.Content) != 1 {
		t.Fatalf("Message.Content length = %d, expected 1", len(event.Message.Content))
	}
	if event.ToolUseResult == nil {
		t.Error("ToolUseResult should not be nil")
	}
}

func TestContentBlock_Unmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected content.Block
	}{
		{
			name:     "text block",
			jsonData: `{"type": "text", "text": "hello world"}`,
			expected: content.Block{Type: "text", Text: "hello world"},
		},
		{
			name:     "empty text",
			jsonData: `{"type": "text", "text": ""}`,
			expected: content.Block{Type: "text", Text: ""},
		},
		{
			name:     "image block without text",
			jsonData: `{"type": "image"}`,
			expected: content.Block{Type: "image", Text: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var block content.Block
			err := json.Unmarshal([]byte(tt.jsonData), &block)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			if block.Type != tt.expected.Type {
				t.Errorf("Type = %q, expected %q", block.Type, tt.expected.Type)
			}
			if block.Text != tt.expected.Text {
				t.Errorf("Text = %q, expected %q", block.Text, tt.expected.Text)
			}
		})
	}
}

// Mock implementations for testing

type mockConfigProvider struct {
	verbose   bool
	noColor   bool
	showUsage bool
}

func (m mockConfigProvider) IsVerbose() bool { return m.verbose }
func (m mockConfigProvider) NoColor() bool   { return m.noColor }
func (m mockConfigProvider) ShowUsage() bool { return m.showUsage }

type mockStyleApplier struct{}

func (m mockStyleApplier) ErrorRender(text string) string   { return "[ERROR:" + text + "]" }
func (m mockStyleApplier) MutedRender(text string) string   { return "[MUTED:" + text + "]" }
func (m mockStyleApplier) SuccessRender(text string) string { return "[SUCCESS:" + text + "]" }
func (m mockStyleApplier) WarningRender(text string) string { return "[WARNING:" + text + "]" }
func (m mockStyleApplier) OutputPrefix() string             { return "  ⎿  " }
func (m mockStyleApplier) OutputContinue() string           { return "     " }
func (m mockStyleApplier) Bullet() string                   { return "● " }
func (m mockStyleApplier) LineNumberRender(text string) string    { return "[LN:" + text + "]" }
func (m mockStyleApplier) LineNumberSepRender(text string) string { return "│" }
func (m mockStyleApplier) DiffAddRender(text string) string       { return "[ADD:" + text + "]" }
func (m mockStyleApplier) DiffRemoveRender(text string) string    { return "[REM:" + text + "]" }
func (m mockStyleApplier) DiffAddBg() lipgloss.Color              { return lipgloss.Color("#00ff00") }
func (m mockStyleApplier) DiffRemoveBg() lipgloss.Color           { return lipgloss.Color("#ff0000") }
func (m mockStyleApplier) SessionHeaderRender(text string) string { return "[HEADER:" + text + "]" }
func (m mockStyleApplier) ApplyThemeBoldGradient(text string) string { return "[GRADIENT:" + text + "]" }
func (m mockStyleApplier) ApplySuccessGradient(text string) string   { return "[SUCCESS_GRAD:" + text + "]" }
func (m mockStyleApplier) ApplyErrorGradient(text string) string     { return "[ERROR_GRAD:" + text + "]" }
func (m mockStyleApplier) NoColor() bool                             { return true }

// Ultraviolet-based style methods
func (m mockStyleApplier) UVSuccessText(text string) string     { return "[UV_SUCCESS:" + text + "]" }
func (m mockStyleApplier) UVWarningText(text string) string     { return "[UV_WARNING:" + text + "]" }
func (m mockStyleApplier) UVMutedText(text string) string       { return "[UV_MUTED:" + text + "]" }
func (m mockStyleApplier) UVErrorText(text string) string       { return "[UV_ERROR:" + text + "]" }
func (m mockStyleApplier) UVErrorBoldText(text string) string   { return "[UV_ERROR_BOLD:" + text + "]" }
func (m mockStyleApplier) UVSuccessBoldText(text string) string { return "[UV_SUCCESS_BOLD:" + text + "]" }

type mockCodeHighlighter struct{}

func (m mockCodeHighlighter) Highlight(code, language string) string { return code }
func (m mockCodeHighlighter) HighlightFile(code, filename string) string {
	return code
}
func (m mockCodeHighlighter) HighlightWithBg(code, language string, bgColor lipgloss.Color) string {
	return code
}
func (m mockCodeHighlighter) HighlightFileWithBg(code, filename string, bgColor lipgloss.Color) string {
	return code
}

func TestRenderer_SetToolContext(t *testing.T) {
	r := NewRendererWithOptions(
		WithConfigProvider(mockConfigProvider{noColor: true}),
	)

	r.SetToolContext(tools.ToolContext{ToolName: "Read", FilePath: "/path/to/file.go"})

	if r.toolContext.ToolName != "Read" {
		t.Errorf("ToolName = %q, expected %q", r.toolContext.ToolName, "Read")
	}
	if r.toolContext.FilePath != "/path/to/file.go" {
		t.Errorf("FilePath = %q, expected %q", r.toolContext.FilePath, "/path/to/file.go")
	}
}

func TestRenderer_Render_NonVerbose(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"line1\nline2\nline3"`),
					IsError:    false,
				},
			},
		},
	}

	r.Render(event)
	output := buf.String()

	// Non-verbose mode should show summary
	if !strings.Contains(output, "Read 3 lines") {
		t.Errorf("Expected 'Read 3 lines' in output, got: %q", output)
	}
	if !strings.Contains(output, "[MUTED:") {
		t.Errorf("Expected muted style in output, got: %q", output)
	}
}

func TestRenderer_Render_Verbose(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"line1\nline2\nline3"`),
					IsError:    false,
				},
			},
		},
	}

	r.Render(event)
	output := buf.String()

	// Verbose mode should show content
	if !strings.Contains(output, "line1") {
		t.Errorf("Expected 'line1' in output, got: %q", output)
	}
	if !strings.Contains(output, "line2") {
		t.Errorf("Expected 'line2' in output, got: %q", output)
	}
	if !strings.Contains(output, "line3") {
		t.Errorf("Expected 'line3' in output, got: %q", output)
	}
}

func TestRenderer_Render_Error(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"Something went wrong"`),
					IsError:    true,
				},
			},
		},
	}

	r.Render(event)
	output := buf.String()

	// Should show error styled
	if !strings.Contains(output, "[ERROR:") {
		t.Errorf("Expected error style in output, got: %q", output)
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("Expected error message in output, got: %q", output)
	}
}

func TestRenderer_Render_EmptyContent(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`""`),
					IsError:    false,
				},
			},
		},
	}

	r.Render(event)
	output := buf.String()

	// Empty content should produce no output
	if output != "" {
		t.Errorf("Expected empty output for empty content, got: %q", output)
	}
}

func TestRenderer_Render_MultipleContent(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"first content"`),
					IsError:    false,
				},
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"second content"`),
					IsError:    false,
				},
			},
		},
	}

	r.Render(event)
	output := buf.String()

	// Should have two output prefixes (one for each content block)
	count := strings.Count(output, "  ⎿  ")
	if count != 2 {
		t.Errorf("Expected 2 output prefixes, got %d in: %q", count, output)
	}
}

func TestRenderer_Render_EditResult_Verbose(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	editResult := EditResult{
		FilePath: "/path/to/file.go",
		StructuredPatch: []PatchHunk{
			{
				OldStart: 10,
				OldLines: 2,
				NewStart: 10,
				NewLines: 2,
				Lines:    []string{"-old line", "+new line"},
			},
		},
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Should show diff with styled markers
	if !strings.Contains(output, "[SUCCESS:+]") {
		t.Errorf("Expected success-styled + in output, got: %q", output)
	}
	if !strings.Contains(output, "[ERROR:-]") {
		t.Errorf("Expected error-styled - in output, got: %q", output)
	}
	if !strings.Contains(output, "old line") {
		t.Errorf("Expected 'old line' in output, got: %q", output)
	}
	if !strings.Contains(output, "new line") {
		t.Errorf("Expected 'new line' in output, got: %q", output)
	}
}

func TestRenderer_Render_EditResult_ContextLines(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	editResult := EditResult{
		FilePath: "/path/to/file.txt",
		StructuredPatch: []PatchHunk{
			{
				OldStart: 5,
				OldLines: 3,
				NewStart: 5,
				NewLines: 3,
				Lines:    []string{" context before", "-removed", "+added", " context after"},
			},
		},
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Context lines should not have +/- styling
	if !strings.Contains(output, "context before") {
		t.Errorf("Expected 'context before' in output, got: %q", output)
	}
	if !strings.Contains(output, "context after") {
		t.Errorf("Expected 'context after' in output, got: %q", output)
	}
}

func TestRenderer_Render_EditResult_NonVerbose(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	editResult := EditResult{
		FilePath: "/path/to/file.go",
		StructuredPatch: []PatchHunk{
			{
				OldStart: 10,
				OldLines: 2,
				NewStart: 10,
				NewLines: 2,
				Lines:    []string{"-old line", "+new line"},
			},
		},
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Edit diffs are shown by default (even in non-verbose mode)
	// because developers want to see what changed
	if !strings.Contains(output, "old line") {
		t.Errorf("Expected 'old line' in output, got: %q", output)
	}
	if !strings.Contains(output, "new line") {
		t.Errorf("Expected 'new line' in output, got: %q", output)
	}
}

func TestRenderer_Render_EditResult_EmptyPatch(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	// Edit result without structured patch
	editResult := EditResult{
		FilePath: "/path/to/file.go",
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"fallback content"`),
					IsError:    false,
				},
			},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Should fall back to regular content rendering
	if !strings.Contains(output, "fallback content") {
		t.Errorf("Expected fallback content in output, got: %q", output)
	}
}

func TestRenderer_Render_StripsSystemReminders(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"real content<system-reminder>secret stuff</system-reminder>more content"`),
					IsError:    false,
				},
			},
		},
	}

	r.Render(event)
	output := buf.String()

	// System reminders should be stripped
	if strings.Contains(output, "secret stuff") {
		t.Errorf("Expected system reminder to be stripped, got: %q", output)
	}
	if !strings.Contains(output, "real content") {
		t.Errorf("Expected 'real content' in output, got: %q", output)
	}
}

func TestRenderer_highlightContent_WithToolContext(t *testing.T) {
	highlightFileCalled := false
	var capturedFilename string

	mockHighlighter := &trackingHighlighter{
		highlightFileFunc: func(code, filename string) string {
			highlightFileCalled = true
			capturedFilename = filename
			return code
		},
	}

	r := NewRendererWithOptions(
		WithConfigProvider(mockConfigProvider{noColor: true}),
		WithCodeHighlighter(mockHighlighter),
		WithToolContext(&tools.ToolContext{FilePath: "/path/to/file.go"}),
	)

	r.highlightContent("package main")

	if !highlightFileCalled {
		t.Error("Expected HighlightFile to be called")
	}
	if capturedFilename != "/path/to/file.go" {
		t.Errorf("Expected filename '/path/to/file.go', got %q", capturedFilename)
	}
}

type trackingHighlighter struct {
	highlightFunc     func(code, language string) string
	highlightFileFunc func(code, filename string) string
}

func (t *trackingHighlighter) Highlight(code, language string) string {
	if t.highlightFunc != nil {
		return t.highlightFunc(code, language)
	}
	return code
}

func (t *trackingHighlighter) HighlightFile(code, filename string) string {
	if t.highlightFileFunc != nil {
		return t.highlightFileFunc(code, filename)
	}
	return code
}

func (t *trackingHighlighter) HighlightWithBg(code, language string, bgColor lipgloss.Color) string {
	return code
}

func (t *trackingHighlighter) HighlightFileWithBg(code, filename string, bgColor lipgloss.Color) string {
	return code
}

func TestRenderer_highlightContent_WithUnknownExtension(t *testing.T) {
	highlightFileCalled := false

	mockHighlighter := &trackingHighlighter{
		highlightFileFunc: func(code, filename string) string {
			highlightFileCalled = true
			return code
		},
	}

	r := NewRendererWithOptions(
		WithConfigProvider(mockConfigProvider{noColor: true}),
		WithCodeHighlighter(mockHighlighter),
		WithToolContext(&tools.ToolContext{FilePath: "/path/to/file.unknown"}), // Unknown extension
	)

	r.highlightContent("some content")

	if !highlightFileCalled {
		t.Error("Expected HighlightFile to be called")
	}
}

func TestNewRenderer_Defaults(t *testing.T) {
	r := NewRenderer()

	if r.output == nil {
		t.Error("Expected output to be set")
	}
	if r.config == nil {
		t.Error("Expected config to be set")
	}
	if r.styleApplier == nil {
		t.Error("Expected styleApplier to be set")
	}
	if r.highlighter == nil {
		t.Error("Expected highlighter to be set")
	}
	if r.toolContext == nil {
		t.Error("Expected toolContext to be set")
	}
}

func TestNewRendererWithOptions_AllOptions(t *testing.T) {
	var buf bytes.Buffer
	cc := mockConfigProvider{verbose: true}
	sa := mockStyleApplier{}
	ch := mockCodeHighlighter{}
	tc := &tools.ToolContext{ToolName: "Test", FilePath: "/test"}

	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(cc),
		WithStyleApplier(sa),
		WithCodeHighlighter(ch),
		WithToolContext(tc),
	)

	if r.output != &buf {
		t.Error("Expected output to be set via option")
	}
	if r.config != cc {
		t.Error("Expected config to be set via option")
	}
	if r.toolContext != tc {
		t.Error("Expected toolContext to be set via option")
	}
}

func TestRenderer_Render_EditResult_MultipleHunks(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	editResult := EditResult{
		FilePath: "/path/to/file.go",
		StructuredPatch: []PatchHunk{
			{
				OldStart: 5,
				OldLines: 1,
				NewStart: 5,
				NewLines: 1,
				Lines:    []string{"-first change"},
			},
			{
				OldStart: 20,
				OldLines: 1,
				NewStart: 20,
				NewLines: 1,
				Lines:    []string{"+second change"},
			},
		},
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	if !strings.Contains(output, "first change") {
		t.Errorf("Expected 'first change' in output, got: %q", output)
	}
	if !strings.Contains(output, "second change") {
		t.Errorf("Expected 'second change' in output, got: %q", output)
	}
}

func TestRenderer_Render_EditResult_SkipsEmptyLines(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	editResult := EditResult{
		FilePath: "/path/to/file.go",
		StructuredPatch: []PatchHunk{
			{
				OldStart: 10,
				OldLines: 2,
				NewStart: 10,
				NewLines: 2,
				Lines:    []string{"", "-valid line", "", "+another valid"},
			},
		},
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Should have rendered valid lines
	if !strings.Contains(output, "valid line") {
		t.Errorf("Expected 'valid line' in output, got: %q", output)
	}
	if !strings.Contains(output, "another valid") {
		t.Errorf("Expected 'another valid' in output, got: %q", output)
	}
}

func TestRenderer_Render_EditResult_InvalidJSON(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"fallback content"`),
					IsError:    false,
				},
			},
		},
		ToolUseResult: json.RawMessage(`not valid json`),
	}

	r.Render(event)
	output := buf.String()

	// Should fall back to regular content rendering
	if !strings.Contains(output, "fallback content") {
		t.Errorf("Expected fallback content in output, got: %q", output)
	}
}

func TestRenderer_Render_LineNumbers(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	editResult := EditResult{
		FilePath: "/path/to/file.go",
		StructuredPatch: []PatchHunk{
			{
				OldStart: 100,
				OldLines: 2,
				NewStart: 100,
				NewLines: 2,
				Lines:    []string{"-old at 100", "+new at 100"},
			},
		},
	}
	toolUseResult, _ := json.Marshal(editResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Should have line numbers
	if !strings.Contains(output, "[LN:100]") {
		t.Errorf("Expected line number 100 in output, got: %q", output)
	}
}

func TestRenderer_Render_WriteResult_Create(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	writeResult := WriteResult{
		Type:     "create",
		FilePath: "/path/to/new-file.txt",
		Content:  "line 1\nline 2\nline 3",
	}
	toolUseResult, _ := json.Marshal(writeResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Write results should show "Created (N lines)" instead of "Read N lines"
	if !strings.Contains(output, "Created (3 lines)") {
		t.Errorf("Expected 'Created (3 lines)' in output, got: %q", output)
	}
	// Should not show the misleading "Read" message
	if strings.Contains(output, "Read") {
		t.Errorf("Should not show 'Read' for write results, got: %q", output)
	}
}

func TestRenderer_Render_WriteResult_SingleLine(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	writeResult := WriteResult{
		Type:     "create",
		FilePath: "/path/to/file.txt",
		Content:  "single line content",
	}
	toolUseResult, _ := json.Marshal(writeResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Single line file should show "Created (1 lines)"
	if !strings.Contains(output, "Created (1 lines)") {
		t.Errorf("Expected 'Created (1 lines)' in output, got: %q", output)
	}
}

func TestRenderer_Render_WriteResult_NotCreate(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: true, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	// Write result with different type should fall through
	writeResult := WriteResult{
		Type:     "update", // Not "create"
		FilePath: "/path/to/file.txt",
		Content:  "content",
	}
	toolUseResult, _ := json.Marshal(writeResult)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"fallback content"`),
					IsError:    false,
				},
			},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Should fall back to regular content rendering
	if !strings.Contains(output, "fallback content") {
		t.Errorf("Expected fallback content in output, got: %q", output)
	}
}

func TestWriteResult_Unmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected WriteResult
	}{
		{
			name:     "complete write result",
			jsonData: `{"type": "create", "filePath": "/path/to/file.txt", "content": "hello world"}`,
			expected: WriteResult{
				Type:     "create",
				FilePath: "/path/to/file.txt",
				Content:  "hello world",
			},
		},
		{
			name:     "write result with multiline content",
			jsonData: `{"type": "create", "filePath": "/test.go", "content": "line 1\nline 2\nline 3"}`,
			expected: WriteResult{
				Type:     "create",
				FilePath: "/test.go",
				Content:  "line 1\nline 2\nline 3",
			},
		},
		{
			name:     "minimal write result",
			jsonData: `{"type": "create", "filePath": "/test.txt"}`,
			expected: WriteResult{
				Type:     "create",
				FilePath: "/test.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result WriteResult
			err := json.Unmarshal([]byte(tt.jsonData), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Type = %q, expected %q", result.Type, tt.expected.Type)
			}
			if result.FilePath != tt.expected.FilePath {
				t.Errorf("FilePath = %q, expected %q", result.FilePath, tt.expected.FilePath)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Content = %q, expected %q", result.Content, tt.expected.Content)
			}
		})
	}
}

func TestRendererSetToolContextMultiple(t *testing.T) {
	r := NewRenderer()

	// Set initial context
	r.SetToolContext(tools.ToolContext{ToolName: "Bash", FilePath: "/script.sh"})

	if r.toolContext.ToolName != "Bash" {
		t.Errorf("ToolName = %q, expected %q", r.toolContext.ToolName, "Bash")
	}
	if r.toolContext.FilePath != "/script.sh" {
		t.Errorf("FilePath = %q, expected %q", r.toolContext.FilePath, "/script.sh")
	}

	// Update context
	r.SetToolContext(tools.ToolContext{ToolName: "Read", FilePath: "/file.go"})

	if r.toolContext.ToolName != "Read" {
		t.Errorf("ToolName = %q, expected %q", r.toolContext.ToolName, "Read")
	}
	if r.toolContext.FilePath != "/file.go" {
		t.Errorf("FilePath = %q, expected %q", r.toolContext.FilePath, "/file.go")
	}
}

func TestDefaultConfigProvider(t *testing.T) {
	cp := config.DefaultProvider{}

	// These just verify the methods exist and return bools
	_ = cp.IsVerbose()
	_ = cp.NoColor()
	_ = cp.ShowUsage()
}

func TestDefaultStyleApplier(t *testing.T) {
	sa := render.DefaultStyleApplier{}

	// Verify all methods exist and return strings
	_ = sa.ErrorRender("test")
	_ = sa.MutedRender("test")
	_ = sa.SuccessRender("test")
	_ = sa.WarningRender("test")
	_ = sa.OutputPrefix()
	_ = sa.OutputContinue()
	_ = sa.Bullet()
	_ = sa.LineNumberRender("test")
	_ = sa.LineNumberSepRender("│")
	_ = sa.DiffAddRender("test")
	_ = sa.DiffRemoveRender("test")
	_ = sa.DiffAddBg()
	_ = sa.DiffRemoveBg()
	_ = sa.SessionHeaderRender("test")
	_ = sa.ApplyThemeBoldGradient("test")
	_ = sa.ApplySuccessGradient("test")
	_ = sa.ApplyErrorGradient("test")
	_ = sa.NoColor()
}

func TestRenderer_Render_TodoResult(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	todoResult := TodoResult{
		OldTodos: []Todo{},
		NewTodos: []Todo{
			{Content: "Review code", Status: "completed", ActiveForm: "Reviewing code"},
			{Content: "Write tests", Status: "in_progress", ActiveForm: "Writing tests"},
			{Content: "Update docs", Status: "pending", ActiveForm: "Updating docs"},
		},
	}
	toolUseResult, _ := json.Marshal(todoResult)

	event := Event{
		Message: Message{
			Role:    "user",
			Content: []ToolResultContent{},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Check completed task shows checkmark and content is muted (using UV methods)
	if !strings.Contains(output, "[UV_SUCCESS:✓]") {
		t.Errorf("Expected UV success-styled checkmark for completed task, got: %q", output)
	}
	if !strings.Contains(output, "[UV_MUTED:Review code]") {
		t.Errorf("Expected UV muted content for completed task, got: %q", output)
	}

	// Check in_progress task shows arrow and uses activeForm (using UV methods)
	if !strings.Contains(output, "[UV_WARNING:→]") {
		t.Errorf("Expected UV warning-styled arrow for in_progress task, got: %q", output)
	}
	if !strings.Contains(output, "Writing tests") {
		t.Errorf("Expected activeForm for in_progress task, got: %q", output)
	}

	// Check pending task shows circle and content is muted (using UV methods)
	if !strings.Contains(output, "[UV_MUTED:○]") {
		t.Errorf("Expected UV muted circle for pending task, got: %q", output)
	}
	if !strings.Contains(output, "[UV_MUTED:Update docs]") {
		t.Errorf("Expected UV muted content for pending task, got: %q", output)
	}

	// First line should have output prefix
	if !strings.HasPrefix(output, "  ⎿  ") {
		t.Errorf("Expected output to start with output prefix, got: %q", output)
	}
}

func TestRenderer_Render_TodoResult_EmptyTodos(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigProvider(mockConfigProvider{verbose: false, noColor: true}),
		WithStyleApplier(mockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	// Empty todo result should fall through to regular rendering
	todoResult := TodoResult{
		OldTodos: []Todo{},
		NewTodos: []Todo{},
	}
	toolUseResult, _ := json.Marshal(todoResult)

	event := Event{
		Message: Message{
			Role: "user",
			Content: []ToolResultContent{
				{
					Type:       "tool_result",
					RawContent: json.RawMessage(`"fallback content"`),
					IsError:    false,
				},
			},
		},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	// Should fall back to regular content rendering
	if !strings.Contains(output, "Read 1 lines") {
		t.Errorf("Expected fallback to regular rendering, got: %q", output)
	}
}

func TestTodo_Unmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected Todo
	}{
		{
			name:     "complete todo",
			jsonData: `{"content": "Review code", "status": "completed", "activeForm": "Reviewing code"}`,
			expected: Todo{Content: "Review code", Status: "completed", ActiveForm: "Reviewing code"},
		},
		{
			name:     "pending todo",
			jsonData: `{"content": "Write tests", "status": "pending", "activeForm": "Writing tests"}`,
			expected: Todo{Content: "Write tests", Status: "pending", ActiveForm: "Writing tests"},
		},
		{
			name:     "in_progress todo",
			jsonData: `{"content": "Update docs", "status": "in_progress", "activeForm": "Updating docs"}`,
			expected: Todo{Content: "Update docs", Status: "in_progress", ActiveForm: "Updating docs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Todo
			err := json.Unmarshal([]byte(tt.jsonData), &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			if result.Content != tt.expected.Content {
				t.Errorf("Content = %q, expected %q", result.Content, tt.expected.Content)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status = %q, expected %q", result.Status, tt.expected.Status)
			}
			if result.ActiveForm != tt.expected.ActiveForm {
				t.Errorf("ActiveForm = %q, expected %q", result.ActiveForm, tt.expected.ActiveForm)
			}
		})
	}
}

func TestTodoResult_Unmarshaling(t *testing.T) {
	jsonData := `{
		"oldTodos": [
			{"content": "Old task", "status": "completed", "activeForm": "Old task"}
		],
		"newTodos": [
			{"content": "Task 1", "status": "pending", "activeForm": "Task 1"},
			{"content": "Task 2", "status": "in_progress", "activeForm": "Doing Task 2"}
		]
	}`

	var result TodoResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(result.OldTodos) != 1 {
		t.Errorf("OldTodos length = %d, expected 1", len(result.OldTodos))
	}
	if len(result.NewTodos) != 2 {
		t.Errorf("NewTodos length = %d, expected 2", len(result.NewTodos))
	}
	if result.NewTodos[0].Content != "Task 1" {
		t.Errorf("NewTodos[0].Content = %q, expected %q", result.NewTodos[0].Content, "Task 1")
	}
	if result.NewTodos[1].Status != "in_progress" {
		t.Errorf("NewTodos[1].Status = %q, expected %q", result.NewTodos[1].Status, "in_progress")
	}
}
