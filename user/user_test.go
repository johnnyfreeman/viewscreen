package user

import (
	"encoding/json"
	"testing"
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
