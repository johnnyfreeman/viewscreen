package user

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

func TestSetToolContext(t *testing.T) {
	// Reset after test
	originalToolName := lastToolName
	originalToolPath := lastToolPath
	defer func() {
		lastToolName = originalToolName
		lastToolPath = originalToolPath
	}()

	SetToolContext("Read", "/path/to/file.go")

	if lastToolName != "Read" {
		t.Errorf("lastToolName = %q, expected %q", lastToolName, "Read")
	}
	if lastToolPath != "/path/to/file.go" {
		t.Errorf("lastToolPath = %q, expected %q", lastToolPath, "/path/to/file.go")
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
		expected ContentBlock
	}{
		{
			name:     "text block",
			jsonData: `{"type": "text", "text": "hello world"}`,
			expected: ContentBlock{Type: "text", Text: "hello world"},
		},
		{
			name:     "empty text",
			jsonData: `{"type": "text", "text": ""}`,
			expected: ContentBlock{Type: "text", Text: ""},
		},
		{
			name:     "image block without text",
			jsonData: `{"type": "image"}`,
			expected: ContentBlock{Type: "image", Text: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var block ContentBlock
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

type mockConfigChecker struct {
	verbose bool
	noColor bool
}

func (m mockConfigChecker) IsVerbose() bool { return m.verbose }
func (m mockConfigChecker) NoColor() bool   { return m.noColor }

type mockStyleApplier struct{}

func (m mockStyleApplier) ErrorRender(text string) string         { return "[ERROR:" + text + "]" }
func (m mockStyleApplier) MutedRender(text string) string         { return "[MUTED:" + text + "]" }
func (m mockStyleApplier) SuccessRender(text string) string       { return "[SUCCESS:" + text + "]" }
func (m mockStyleApplier) OutputPrefix() string                   { return "  ⎿  " }
func (m mockStyleApplier) OutputContinue() string                 { return "     " }
func (m mockStyleApplier) LineNumberRender(text string) string    { return "[LN:" + text + "]" }
func (m mockStyleApplier) LineNumberSepRender(text string) string { return "│" }
func (m mockStyleApplier) DiffAddRender(text string) string       { return "[ADD:" + text + "]" }
func (m mockStyleApplier) DiffRemoveRender(text string) string    { return "[REM:" + text + "]" }
func (m mockStyleApplier) DiffAddBg() lipgloss.Color              { return lipgloss.Color("#00ff00") }
func (m mockStyleApplier) DiffRemoveBg() lipgloss.Color           { return lipgloss.Color("#ff0000") }

type mockCodeHighlighter struct{}

func (m mockCodeHighlighter) Highlight(code, language string) string { return code }
func (m mockCodeHighlighter) HighlightFile(code, filename string) string {
	return code
}
func (m mockCodeHighlighter) HighlightWithBg(code, language string, bgColor lipgloss.Color) string {
	return code
}

func TestRenderer_SetToolContext(t *testing.T) {
	r := NewRendererWithOptions(
		WithConfigChecker(mockConfigChecker{noColor: true}),
	)

	r.SetToolContext("Read", "/path/to/file.go")

	if r.toolContext.ToolName != "Read" {
		t.Errorf("ToolName = %q, expected %q", r.toolContext.ToolName, "Read")
	}
	if r.toolContext.ToolPath != "/path/to/file.go" {
		t.Errorf("ToolPath = %q, expected %q", r.toolContext.ToolPath, "/path/to/file.go")
	}
}

func TestRenderer_Render_NonVerbose(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigChecker(mockConfigChecker{verbose: false, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: false, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: false, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: false, noColor: true}),
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

	// Non-verbose mode should NOT render diff (tryRenderEditResult returns early)
	if strings.Contains(output, "old line") || strings.Contains(output, "new line") {
		t.Errorf("Expected no diff output in non-verbose mode, got: %q", output)
	}
}

func TestRenderer_Render_EditResult_EmptyPatch(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
	highlightCalled := false
	var capturedLang string

	mockHighlighter := &trackingHighlighter{
		highlightFunc: func(code, language string) string {
			highlightCalled = true
			capturedLang = language
			return code
		},
	}

	r := NewRendererWithOptions(
		WithConfigChecker(mockConfigChecker{noColor: true}),
		WithCodeHighlighter(mockHighlighter),
		WithToolContext(&ToolContext{ToolPath: "/path/to/file.go"}),
	)

	r.highlightContent("package main")

	if !highlightCalled {
		t.Error("Expected Highlight to be called")
	}
	if capturedLang != "go" {
		t.Errorf("Expected language 'go', got %q", capturedLang)
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

func TestRenderer_highlightContent_FallsBackToHighlightFile(t *testing.T) {
	highlightFileCalled := false

	mockHighlighter := &trackingHighlighter{
		highlightFileFunc: func(code, filename string) string {
			highlightFileCalled = true
			return code
		},
	}

	r := NewRendererWithOptions(
		WithConfigChecker(mockConfigChecker{noColor: true}),
		WithCodeHighlighter(mockHighlighter),
		WithToolContext(&ToolContext{ToolPath: "/path/to/file.unknown"}), // Unknown extension
	)

	r.highlightContent("some content")

	if !highlightFileCalled {
		t.Error("Expected HighlightFile to be called as fallback")
	}
}

func TestNewRenderer_Defaults(t *testing.T) {
	r := NewRenderer()

	if r.output == nil {
		t.Error("Expected output to be set")
	}
	if r.configChecker == nil {
		t.Error("Expected configChecker to be set")
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
	cc := mockConfigChecker{verbose: true}
	sa := mockStyleApplier{}
	ch := mockCodeHighlighter{}
	tc := &ToolContext{ToolName: "Test", ToolPath: "/test"}

	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigChecker(cc),
		WithStyleApplier(sa),
		WithCodeHighlighter(ch),
		WithToolContext(tc),
	)

	if r.output != &buf {
		t.Error("Expected output to be set via option")
	}
	if r.configChecker != cc {
		t.Error("Expected configChecker to be set via option")
	}
	if r.toolContext != tc {
		t.Error("Expected toolContext to be set via option")
	}
}

func TestRenderer_Render_EditResult_MultipleHunks(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererWithOptions(
		WithOutput(&buf),
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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
		WithConfigChecker(mockConfigChecker{verbose: true, noColor: true}),
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

func TestPackageLevelSetToolContext(t *testing.T) {
	// Reset package state after test
	originalToolName := lastToolName
	originalToolPath := lastToolPath
	originalRenderer := defaultRenderer
	defer func() {
		lastToolName = originalToolName
		lastToolPath = originalToolPath
		defaultRenderer = originalRenderer
	}()

	// Test without default renderer
	defaultRenderer = nil
	SetToolContext("Bash", "/script.sh")

	if lastToolName != "Bash" {
		t.Errorf("lastToolName = %q, expected %q", lastToolName, "Bash")
	}
	if lastToolPath != "/script.sh" {
		t.Errorf("lastToolPath = %q, expected %q", lastToolPath, "/script.sh")
	}

	// Test with default renderer
	defaultRenderer = NewRenderer()
	SetToolContext("Read", "/file.go")

	if defaultRenderer.toolContext.ToolName != "Read" {
		t.Errorf("renderer ToolName = %q, expected %q", defaultRenderer.toolContext.ToolName, "Read")
	}
}

func TestDefaultConfigChecker(t *testing.T) {
	cc := DefaultConfigChecker{}

	// These just verify the methods exist and return bools
	_ = cc.IsVerbose()
	_ = cc.NoColor()
}

func TestDefaultStyleApplier(t *testing.T) {
	sa := DefaultStyleApplier{}

	// Verify all methods exist and return strings
	_ = sa.ErrorRender("test")
	_ = sa.MutedRender("test")
	_ = sa.SuccessRender("test")
	_ = sa.OutputPrefix()
	_ = sa.OutputContinue()
	_ = sa.LineNumberRender("test")
	_ = sa.LineNumberSepRender("│")
	_ = sa.DiffAddRender("test")
	_ = sa.DiffRemoveRender("test")
	_ = sa.DiffAddBg()
	_ = sa.DiffRemoveBg()
}
