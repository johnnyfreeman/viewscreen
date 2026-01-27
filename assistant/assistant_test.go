package assistant

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/types"
)

// mockMarkdownRenderer is a test double for MarkdownRendererInterface
type mockMarkdownRenderer struct {
	renderCalls []string
	returnValue string
}

func (m *mockMarkdownRenderer) Render(content string) string {
	m.renderCalls = append(m.renderCalls, content)
	if m.returnValue != "" {
		return m.returnValue
	}
	return content
}

// mockToolUseRenderer tracks tool use render calls
type mockToolUseRenderer struct {
	calls []types.ContentBlock
}

func (m *mockToolUseRenderer) render(out *render.Output, block types.ContentBlock) {
	m.calls = append(m.calls, block)
}

func TestNewRenderer(t *testing.T) {
	r := NewRenderer()

	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	if r.output == nil {
		t.Error("expected output to be non-nil")
	}

	if r.markdownRenderer == nil {
		t.Error("expected markdownRenderer to be non-nil")
	}

	if r.toolUseRenderer == nil {
		t.Error("expected toolUseRenderer to be non-nil")
	}
}

func TestNewRendererWithOptions(t *testing.T) {
	t.Run("with custom output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		r := NewRendererWithOptions(WithOutput(buf))

		if r.output != buf {
			t.Error("expected custom output writer")
		}
	})

	t.Run("with custom markdown renderer", func(t *testing.T) {
		mock := &mockMarkdownRenderer{}
		r := NewRendererWithOptions(WithMarkdownRenderer(mock))

		if r.markdownRenderer != mock {
			t.Error("expected custom markdown renderer")
		}
	})

	t.Run("with custom tool use renderer", func(t *testing.T) {
		called := false
		custom := func(out *render.Output, block types.ContentBlock) {
			called = true
		}
		r := NewRendererWithOptions(WithToolUseRenderer(custom))

		r.toolUseRenderer(render.StringOutput(), types.ContentBlock{})
		if !called {
			t.Error("expected custom tool use renderer to be called")
		}
	})

	t.Run("with multiple options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		mock := &mockMarkdownRenderer{}

		r := NewRendererWithOptions(
			WithOutput(buf),
			WithMarkdownRenderer(mock),
		)

		if r.output != buf {
			t.Error("expected custom output writer")
		}
		if r.markdownRenderer != mock {
			t.Error("expected custom markdown renderer")
		}
	})
}

func TestRenderer_Render_Error(t *testing.T) {
	output := &bytes.Buffer{}
	r := NewRendererWithOptions(WithOutput(output))

	event := Event{
		Error: "Something went wrong",
		Message: Message{
			Content: []types.ContentBlock{},
		},
	}

	r.Render(event, false, false)

	result := output.String()
	if result == "" {
		t.Error("expected error output, got empty string")
	}

	// Check that error message is rendered
	if !bytes.Contains(output.Bytes(), []byte("Error")) {
		t.Error("expected 'Error' header in output")
	}
	if !bytes.Contains(output.Bytes(), []byte("Something went wrong")) {
		t.Error("expected error message in output")
	}
}

func TestRenderer_Render_TextBlock_NotStreaming(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{returnValue: "rendered markdown\n"}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "text", Text: "Hello World"},
			},
		},
	}

	// Not streaming (inTextBlock = false)
	r.Render(event, false, false)

	if len(markdown.renderCalls) != 1 {
		t.Errorf("expected markdown.Render() to be called once, got %d", len(markdown.renderCalls))
	}

	if markdown.renderCalls[0] != "Hello World" {
		t.Errorf("expected markdown.Render() to be called with 'Hello World', got %q", markdown.renderCalls[0])
	}

	if output.String() != "rendered markdown\n" {
		t.Errorf("expected output 'rendered markdown\\n', got %q", output.String())
	}
}

func TestRenderer_Render_TextBlock_AlreadyStreaming(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "text", Text: "Hello World"},
			},
		},
	}

	// Already streaming (inTextBlock = true), should skip rendering
	r.Render(event, true, false)

	if len(markdown.renderCalls) != 0 {
		t.Error("expected markdown.Render() not to be called when already streaming")
	}

	if output.String() != "" {
		t.Errorf("expected no output when already streaming, got %q", output.String())
	}
}

func TestRenderer_Render_TextBlock_AddsNewline(t *testing.T) {
	output := &bytes.Buffer{}
	// Return value without trailing newline
	markdown := &mockMarkdownRenderer{returnValue: "no trailing newline"}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "text", Text: "test"},
			},
		},
	}

	r.Render(event, false, false)

	// Should add newline if rendered content doesn't end with one
	if !bytes.HasSuffix(output.Bytes(), []byte("\n")) {
		t.Error("expected output to end with newline")
	}
}

func TestRenderer_Render_TextBlock_PreservesNewline(t *testing.T) {
	output := &bytes.Buffer{}
	// Return value with trailing newline
	markdown := &mockMarkdownRenderer{returnValue: "has trailing newline\n"}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "text", Text: "test"},
			},
		},
	}

	r.Render(event, false, false)

	// Should not add extra newline
	if output.String() != "has trailing newline\n" {
		t.Errorf("expected 'has trailing newline\\n', got %q", output.String())
	}
}

func TestRenderer_Render_ToolUseBlock_NotStreaming(t *testing.T) {
	toolRenderer := &mockToolUseRenderer{}

	r := NewRendererWithOptions(
		WithToolUseRenderer(toolRenderer.render),
	)

	inputJSON, _ := json.Marshal(map[string]string{"file_path": "/test.go"})
	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{
					Type:  "tool_use",
					ID:    "tool_123",
					Name:  "Read",
					Input: inputJSON,
				},
			},
		},
	}

	// Not streaming (inToolUseBlock = false)
	r.Render(event, false, false)

	if len(toolRenderer.calls) != 1 {
		t.Errorf("expected tool renderer to be called once, got %d", len(toolRenderer.calls))
	}

	if toolRenderer.calls[0].Name != "Read" {
		t.Errorf("expected tool name 'Read', got %q", toolRenderer.calls[0].Name)
	}

	if toolRenderer.calls[0].ID != "tool_123" {
		t.Errorf("expected tool ID 'tool_123', got %q", toolRenderer.calls[0].ID)
	}
}

func TestRenderer_Render_ToolUseBlock_AlreadyStreaming(t *testing.T) {
	toolRenderer := &mockToolUseRenderer{}

	r := NewRendererWithOptions(
		WithToolUseRenderer(toolRenderer.render),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{
					Type: "tool_use",
					Name: "Read",
				},
			},
		},
	}

	// Already streaming (inToolUseBlock = true), should skip rendering
	r.Render(event, false, true)

	if len(toolRenderer.calls) != 0 {
		t.Error("expected tool renderer not to be called when already streaming")
	}
}

func TestRenderer_Render_MultipleBlocks(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{returnValue: "text\n"}
	toolRenderer := &mockToolUseRenderer{}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
		WithToolUseRenderer(toolRenderer.render),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "text", Text: "First message"},
				{Type: "tool_use", Name: "Read"},
				{Type: "text", Text: "Second message"},
			},
		},
	}

	r.Render(event, false, false)

	if len(markdown.renderCalls) != 2 {
		t.Errorf("expected markdown.Render() to be called twice, got %d", len(markdown.renderCalls))
	}

	if len(toolRenderer.calls) != 1 {
		t.Errorf("expected tool renderer to be called once, got %d", len(toolRenderer.calls))
	}
}

func TestRenderer_Render_UnknownBlockType(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{}
	toolRenderer := &mockToolUseRenderer{}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
		WithToolUseRenderer(toolRenderer.render),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "unknown_type"},
			},
		},
	}

	// Should not panic and should not call any renderer
	r.Render(event, false, false)

	if len(markdown.renderCalls) != 0 {
		t.Error("expected markdown.Render() not to be called for unknown type")
	}
	if len(toolRenderer.calls) != 0 {
		t.Error("expected tool renderer not to be called for unknown type")
	}
}

func TestRenderer_Render_EmptyContent(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	event := Event{
		Message: Message{
			Content: []types.ContentBlock{},
		},
	}

	// Should not panic
	r.Render(event, false, false)

	if output.String() != "" {
		t.Errorf("expected no output for empty content, got %q", output.String())
	}
}

func TestRenderer_Render_ErrorWithContent(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{returnValue: "text\n"}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	event := Event{
		Error: "An error occurred",
		Message: Message{
			Content: []types.ContentBlock{
				{Type: "text", Text: "Some text"},
			},
		},
	}

	r.Render(event, false, false)

	// Should render both error and text content
	result := output.String()
	if !bytes.Contains(output.Bytes(), []byte("Error")) {
		t.Error("expected error in output")
	}
	if !bytes.Contains(output.Bytes(), []byte("text")) {
		t.Error("expected text content in output")
	}
	if result == "" {
		t.Error("expected non-empty output")
	}
}

func TestEvent_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantErr  bool
		validate func(t *testing.T, e Event)
	}{
		{
			name: "basic event with text content",
			json: `{
				"type": "assistant",
				"message": {
					"id": "msg_123",
					"type": "message",
					"role": "assistant",
					"model": "claude-3",
					"content": [
						{"type": "text", "text": "Hello!"}
					]
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if e.Message.ID != "msg_123" {
					t.Errorf("expected ID 'msg_123', got %q", e.Message.ID)
				}
				if e.Message.Model != "claude-3" {
					t.Errorf("expected Model 'claude-3', got %q", e.Message.Model)
				}
				if len(e.Message.Content) != 1 {
					t.Fatalf("expected 1 content block, got %d", len(e.Message.Content))
				}
				if e.Message.Content[0].Type != "text" {
					t.Errorf("expected content type 'text', got %q", e.Message.Content[0].Type)
				}
				if e.Message.Content[0].Text != "Hello!" {
					t.Errorf("expected text 'Hello!', got %q", e.Message.Content[0].Text)
				}
			},
		},
		{
			name: "event with error",
			json: `{
				"type": "assistant",
				"error": "rate_limit_exceeded",
				"message": {
					"content": []
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if e.Error != "rate_limit_exceeded" {
					t.Errorf("expected error 'rate_limit_exceeded', got %q", e.Error)
				}
			},
		},
		{
			name: "event with tool_use content",
			json: `{
				"type": "assistant",
				"message": {
					"content": [
						{
							"type": "tool_use",
							"id": "tool_abc",
							"name": "Read",
							"input": {"file_path": "/test.go"}
						}
					]
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if len(e.Message.Content) != 1 {
					t.Fatalf("expected 1 content block, got %d", len(e.Message.Content))
				}
				block := e.Message.Content[0]
				if block.Type != "tool_use" {
					t.Errorf("expected type 'tool_use', got %q", block.Type)
				}
				if block.Name != "Read" {
					t.Errorf("expected name 'Read', got %q", block.Name)
				}
				if block.ID != "tool_abc" {
					t.Errorf("expected ID 'tool_abc', got %q", block.ID)
				}
			},
		},
		{
			name: "event with stop_reason",
			json: `{
				"type": "assistant",
				"message": {
					"content": [],
					"stop_reason": "end_turn"
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if e.Message.StopReason == nil {
					t.Fatal("expected stop_reason to be non-nil")
				}
				if *e.Message.StopReason != "end_turn" {
					t.Errorf("expected stop_reason 'end_turn', got %q", *e.Message.StopReason)
				}
			},
		},
		{
			name: "event with usage",
			json: `{
				"type": "assistant",
				"message": {
					"content": [],
					"usage": {
						"input_tokens": 100,
						"output_tokens": 50
					}
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if e.Message.Usage == nil {
					t.Fatal("expected usage to be non-nil")
				}
				if e.Message.Usage.InputTokens != 100 {
					t.Errorf("expected input_tokens 100, got %d", e.Message.Usage.InputTokens)
				}
				if e.Message.Usage.OutputTokens != 50 {
					t.Errorf("expected output_tokens 50, got %d", e.Message.Usage.OutputTokens)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event Event
			err := json.Unmarshal([]byte(tt.json), &event)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, event)
			}
		})
	}
}

func TestMessage_JSONUnmarshal(t *testing.T) {
	input := `{
		"id": "msg_abc",
		"type": "message",
		"role": "assistant",
		"model": "claude-3-opus",
		"content": [
			{"type": "text", "text": "Hello"},
			{"type": "tool_use", "name": "Read"}
		],
		"stop_reason": "tool_use"
	}`

	var msg Message
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg.ID != "msg_abc" {
		t.Errorf("ID: got %q, want 'msg_abc'", msg.ID)
	}
	if msg.Type != "message" {
		t.Errorf("Type: got %q, want 'message'", msg.Type)
	}
	if msg.Role != "assistant" {
		t.Errorf("Role: got %q, want 'assistant'", msg.Role)
	}
	if msg.Model != "claude-3-opus" {
		t.Errorf("Model: got %q, want 'claude-3-opus'", msg.Model)
	}
	if len(msg.Content) != 2 {
		t.Errorf("Content length: got %d, want 2", len(msg.Content))
	}
	if msg.StopReason == nil || *msg.StopReason != "tool_use" {
		t.Errorf("StopReason: got %v, want 'tool_use'", msg.StopReason)
	}
}

