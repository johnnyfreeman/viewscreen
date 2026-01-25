package stream

import (
	"bytes"
	"encoding/json"
	"testing"

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

// mockIndicator is a test double for IndicatorInterface
type mockIndicator struct {
	showCalls  int
	clearCalls int
}

func (m *mockIndicator) Show() {
	m.showCalls++
}

func (m *mockIndicator) Clear() {
	m.clearCalls++
}

// mockToolHeaderRenderer tracks tool header render calls
type mockToolHeaderRenderer struct {
	calls []struct {
		toolName string
		input    map[string]interface{}
	}
}

func (m *mockToolHeaderRenderer) render(toolName string, input map[string]interface{}) {
	m.calls = append(m.calls, struct {
		toolName string
		input    map[string]interface{}
	}{toolName, input})
}

// Helper to create a content block JSON
func makeContentBlock(blockType, name string) json.RawMessage {
	block := types.ContentBlock{
		Type: blockType,
		Name: name,
	}
	data, _ := json.Marshal(block)
	return data
}

// Helper to create a text delta JSON
func makeTextDelta(text string) json.RawMessage {
	delta := TextDelta{
		Type: "text_delta",
		Text: text,
	}
	data, _ := json.Marshal(delta)
	return data
}

// Helper to create an input JSON delta
func makeInputJSONDelta(partialJSON string) json.RawMessage {
	delta := InputJSONDelta{
		Type:        "input_json_delta",
		PartialJSON: partialJSON,
	}
	data, _ := json.Marshal(delta)
	return data
}

func TestNewRenderer(t *testing.T) {
	r := NewRenderer()

	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	if r.CurrentBlockIndex != -1 {
		t.Errorf("expected CurrentBlockIndex to be -1, got %d", r.CurrentBlockIndex)
	}

	if r.markdownRenderer == nil {
		t.Error("expected markdownRenderer to be non-nil")
	}

	if r.indicator == nil {
		t.Error("expected indicator to be non-nil")
	}

	if r.toolHeaderRender == nil {
		t.Error("expected toolHeaderRender to be non-nil")
	}

	if r.output == nil {
		t.Error("expected output to be non-nil")
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

	t.Run("with custom indicator", func(t *testing.T) {
		mock := &mockIndicator{}
		r := NewRendererWithOptions(WithIndicator(mock))

		if r.indicator != mock {
			t.Error("expected custom indicator")
		}
	})

	t.Run("with custom tool header renderer", func(t *testing.T) {
		called := false
		custom := func(toolName string, input map[string]interface{}) {
			called = true
		}
		r := NewRendererWithOptions(WithToolHeaderRenderer(custom))

		r.toolHeaderRender("Test", nil)
		if !called {
			t.Error("expected custom tool header renderer to be called")
		}
	})
}

func TestRenderer_Render_MessageStart(t *testing.T) {
	r := NewRendererWithOptions()

	event := Event{
		Event: EventData{
			Type: "message_start",
		},
	}

	// Should not panic or change state
	r.Render(event)

	if r.InTextBlock {
		t.Error("InTextBlock should remain false")
	}
	if r.InToolUseBlock {
		t.Error("InToolUseBlock should remain false")
	}
}

func TestRenderer_Render_ContentBlockStart_Text(t *testing.T) {
	r := NewRendererWithOptions()

	event := Event{
		Event: EventData{
			Type:         "content_block_start",
			Index:        0,
			ContentBlock: makeContentBlock("text", ""),
		},
	}

	r.Render(event)

	if !r.InTextBlock {
		t.Error("expected InTextBlock to be true")
	}
	if r.InToolUseBlock {
		t.Error("expected InToolUseBlock to be false")
	}
	if r.CurrentBlockIndex != 0 {
		t.Errorf("expected CurrentBlockIndex to be 0, got %d", r.CurrentBlockIndex)
	}
	if r.CurrentBlockType != "text" {
		t.Errorf("expected CurrentBlockType to be 'text', got %s", r.CurrentBlockType)
	}
}

func TestRenderer_Render_ContentBlockStart_ToolUse(t *testing.T) {
	r := NewRendererWithOptions()

	event := Event{
		Event: EventData{
			Type:         "content_block_start",
			Index:        0,
			ContentBlock: makeContentBlock("tool_use", "Read"),
		},
	}

	r.Render(event)

	if r.InTextBlock {
		t.Error("expected InTextBlock to be false")
	}
	if !r.InToolUseBlock {
		t.Error("expected InToolUseBlock to be true")
	}
	if r.toolName != "Read" {
		t.Errorf("expected toolName to be 'Read', got %s", r.toolName)
	}
}

func TestRenderer_Render_ContentBlockStart_InvalidJSON(t *testing.T) {
	r := NewRendererWithOptions()

	event := Event{
		Event: EventData{
			Type:         "content_block_start",
			Index:        0,
			ContentBlock: json.RawMessage(`invalid json`),
		},
	}

	// Should not panic
	r.Render(event)

	if r.InTextBlock {
		t.Error("InTextBlock should remain false with invalid JSON")
	}
}

func TestRenderer_Render_ContentBlockDelta_TextDelta(t *testing.T) {
	indicator := &mockIndicator{}
	r := NewRendererWithOptions(WithIndicator(indicator))

	// First start a text block
	r.InTextBlock = true
	r.CurrentBlockIndex = 0

	event := Event{
		Event: EventData{
			Type:  "content_block_delta",
			Index: 0,
			Delta: makeTextDelta("Hello "),
		},
	}

	r.Render(event)

	if indicator.showCalls != 1 {
		t.Errorf("expected indicator.Show() to be called once, got %d", indicator.showCalls)
	}
	if r.textBuffer.String() != "Hello " {
		t.Errorf("expected textBuffer to contain 'Hello ', got %q", r.textBuffer.String())
	}

	// Send another delta
	event.Event.Delta = makeTextDelta("World!")
	r.Render(event)

	if indicator.showCalls != 2 {
		t.Errorf("expected indicator.Show() to be called twice, got %d", indicator.showCalls)
	}
	if r.textBuffer.String() != "Hello World!" {
		t.Errorf("expected textBuffer to contain 'Hello World!', got %q", r.textBuffer.String())
	}
}

func TestRenderer_Render_ContentBlockDelta_InputJSONDelta(t *testing.T) {
	r := NewRendererWithOptions()

	// Start a tool use block
	r.InToolUseBlock = true
	r.CurrentBlockIndex = 0
	r.toolName = "Read"

	event := Event{
		Event: EventData{
			Type:  "content_block_delta",
			Index: 0,
			Delta: makeInputJSONDelta(`{"file_`),
		},
	}

	r.Render(event)

	if r.toolInput.String() != `{"file_` {
		t.Errorf("expected toolInput to contain partial JSON, got %q", r.toolInput.String())
	}

	// Send more JSON
	event.Event.Delta = makeInputJSONDelta(`path": "/test.go"}`)
	r.Render(event)

	expected := `{"file_path": "/test.go"}`
	if r.toolInput.String() != expected {
		t.Errorf("expected toolInput to contain %q, got %q", expected, r.toolInput.String())
	}
}

func TestRenderer_Render_ContentBlockStop_TextBlock(t *testing.T) {
	output := &bytes.Buffer{}
	indicator := &mockIndicator{}
	markdown := &mockMarkdownRenderer{returnValue: "**rendered**\n"}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithIndicator(indicator),
		WithMarkdownRenderer(markdown),
	)

	// Setup: text block with buffered content
	r.InTextBlock = true
	r.CurrentBlockIndex = 0
	r.textBuffer.WriteString("Hello World")

	event := Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 0,
		},
	}

	r.Render(event)

	if indicator.clearCalls != 1 {
		t.Errorf("expected indicator.Clear() to be called once, got %d", indicator.clearCalls)
	}

	if len(markdown.renderCalls) != 1 {
		t.Errorf("expected markdown.Render() to be called once, got %d", len(markdown.renderCalls))
	}

	if markdown.renderCalls[0] != "Hello World" {
		t.Errorf("expected markdown.Render() to be called with 'Hello World', got %q", markdown.renderCalls[0])
	}

	if output.String() != "**rendered**\n" {
		t.Errorf("expected output to contain rendered markdown, got %q", output.String())
	}
}

func TestRenderer_Render_ContentBlockStop_TextBlock_Empty(t *testing.T) {
	output := &bytes.Buffer{}
	markdown := &mockMarkdownRenderer{}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithMarkdownRenderer(markdown),
	)

	// Setup: text block with no content
	r.InTextBlock = true
	r.CurrentBlockIndex = 0

	event := Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 0,
		},
	}

	r.Render(event)

	// Should not render empty text
	if len(markdown.renderCalls) != 0 {
		t.Error("expected markdown.Render() not to be called for empty text")
	}
}

func TestRenderer_Render_ContentBlockStop_ToolUseBlock(t *testing.T) {
	toolRenderer := &mockToolHeaderRenderer{}

	r := NewRendererWithOptions(
		WithToolHeaderRenderer(toolRenderer.render),
	)

	// Setup: tool use block with accumulated input
	r.InToolUseBlock = true
	r.CurrentBlockIndex = 0
	r.toolName = "Read"
	r.toolInput.WriteString(`{"file_path": "/test.go"}`)

	event := Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 0,
		},
	}

	r.Render(event)

	if len(toolRenderer.calls) != 1 {
		t.Errorf("expected tool header renderer to be called once, got %d", len(toolRenderer.calls))
	}

	if toolRenderer.calls[0].toolName != "Read" {
		t.Errorf("expected tool name 'Read', got %q", toolRenderer.calls[0].toolName)
	}

	if toolRenderer.calls[0].input["file_path"] != "/test.go" {
		t.Errorf("expected file_path '/test.go', got %v", toolRenderer.calls[0].input["file_path"])
	}
}

func TestRenderer_Render_ContentBlockStop_ToolUseBlock_InvalidJSON(t *testing.T) {
	output := &bytes.Buffer{}
	toolRenderer := &mockToolHeaderRenderer{}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithToolHeaderRenderer(toolRenderer.render),
	)

	// Setup: tool use block with invalid JSON
	r.InToolUseBlock = true
	r.CurrentBlockIndex = 0
	r.toolName = "Read"
	r.toolInput.WriteString(`invalid json`)

	event := Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 0,
		},
	}

	r.Render(event)

	// Should use fallback rendering (directly to output)
	if len(toolRenderer.calls) != 0 {
		t.Error("expected tool header renderer not to be called for invalid JSON")
	}

	// Output should contain fallback format
	if output.Len() == 0 {
		t.Error("expected fallback output for invalid JSON")
	}
}

func TestRenderer_Render_ContentBlockStop_WrongIndex(t *testing.T) {
	markdown := &mockMarkdownRenderer{}

	r := NewRendererWithOptions(
		WithMarkdownRenderer(markdown),
	)

	// Setup: text block with content but wrong index
	r.InTextBlock = true
	r.CurrentBlockIndex = 0
	r.textBuffer.WriteString("Hello World")

	event := Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 1, // Different from CurrentBlockIndex
		},
	}

	r.Render(event)

	// Should not render because indices don't match
	if len(markdown.renderCalls) != 0 {
		t.Error("expected markdown.Render() not to be called for mismatched index")
	}
}

func TestRenderer_Render_MessageDelta(t *testing.T) {
	r := NewRendererWithOptions()

	event := Event{
		Event: EventData{
			Type: "message_delta",
		},
	}

	// Should not change any state
	r.Render(event)

	if r.InTextBlock {
		t.Error("InTextBlock should remain false")
	}
}

func TestRenderer_Render_MessageStop(t *testing.T) {
	r := NewRendererWithOptions()
	r.CurrentBlockIndex = 5

	event := Event{
		Event: EventData{
			Type: "message_stop",
		},
	}

	r.Render(event)

	if r.CurrentBlockIndex != -1 {
		t.Errorf("expected CurrentBlockIndex to be reset to -1, got %d", r.CurrentBlockIndex)
	}
}

func TestRenderer_GetBufferedText(t *testing.T) {
	r := NewRendererWithOptions()

	if r.GetBufferedText() != "" {
		t.Error("expected empty buffer initially")
	}

	r.textBuffer.WriteString("test content")

	if r.GetBufferedText() != "test content" {
		t.Errorf("expected 'test content', got %q", r.GetBufferedText())
	}
}

func TestRenderer_ResetBlockState(t *testing.T) {
	r := NewRendererWithOptions()
	r.InTextBlock = true
	r.InToolUseBlock = true

	r.ResetBlockState()

	if r.InTextBlock {
		t.Error("expected InTextBlock to be false after reset")
	}
	if r.InToolUseBlock {
		t.Error("expected InToolUseBlock to be false after reset")
	}
}

func TestRenderer_FullTextBlockFlow(t *testing.T) {
	output := &bytes.Buffer{}
	indicator := &mockIndicator{}
	markdown := &mockMarkdownRenderer{returnValue: "rendered text\n"}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithIndicator(indicator),
		WithMarkdownRenderer(markdown),
	)

	// 1. Content block start
	r.Render(Event{
		Event: EventData{
			Type:         "content_block_start",
			Index:        0,
			ContentBlock: makeContentBlock("text", ""),
		},
	})

	if !r.InTextBlock {
		t.Error("expected InTextBlock after start")
	}

	// 2. Multiple deltas
	r.Render(Event{
		Event: EventData{
			Type:  "content_block_delta",
			Index: 0,
			Delta: makeTextDelta("Hello "),
		},
	})

	r.Render(Event{
		Event: EventData{
			Type:  "content_block_delta",
			Index: 0,
			Delta: makeTextDelta("World!"),
		},
	})

	if indicator.showCalls != 2 {
		t.Errorf("expected 2 show calls, got %d", indicator.showCalls)
	}

	// 3. Content block stop
	r.Render(Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 0,
		},
	})

	if indicator.clearCalls != 1 {
		t.Errorf("expected 1 clear call, got %d", indicator.clearCalls)
	}

	if len(markdown.renderCalls) != 1 {
		t.Errorf("expected 1 render call, got %d", len(markdown.renderCalls))
	}

	if markdown.renderCalls[0] != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", markdown.renderCalls[0])
	}

	if output.String() != "rendered text\n" {
		t.Errorf("expected output 'rendered text\\n', got %q", output.String())
	}
}

func TestRenderer_FullToolUseBlockFlow(t *testing.T) {
	toolRenderer := &mockToolHeaderRenderer{}

	r := NewRendererWithOptions(
		WithToolHeaderRenderer(toolRenderer.render),
	)

	// 1. Content block start
	r.Render(Event{
		Event: EventData{
			Type:         "content_block_start",
			Index:        0,
			ContentBlock: makeContentBlock("tool_use", "Bash"),
		},
	})

	if !r.InToolUseBlock {
		t.Error("expected InToolUseBlock after start")
	}
	if r.toolName != "Bash" {
		t.Errorf("expected toolName 'Bash', got %q", r.toolName)
	}

	// 2. Multiple input JSON deltas
	r.Render(Event{
		Event: EventData{
			Type:  "content_block_delta",
			Index: 0,
			Delta: makeInputJSONDelta(`{"command": `),
		},
	})

	r.Render(Event{
		Event: EventData{
			Type:  "content_block_delta",
			Index: 0,
			Delta: makeInputJSONDelta(`"ls -la"}`),
		},
	})

	// 3. Content block stop
	r.Render(Event{
		Event: EventData{
			Type:  "content_block_stop",
			Index: 0,
		},
	})

	if len(toolRenderer.calls) != 1 {
		t.Errorf("expected 1 tool render call, got %d", len(toolRenderer.calls))
	}

	if toolRenderer.calls[0].toolName != "Bash" {
		t.Errorf("expected tool name 'Bash', got %q", toolRenderer.calls[0].toolName)
	}

	if toolRenderer.calls[0].input["command"] != "ls -la" {
		t.Errorf("expected command 'ls -la', got %v", toolRenderer.calls[0].input["command"])
	}
}

func TestEventData_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected EventData
	}{
		{
			name: "message_start",
			json: `{"type": "message_start"}`,
			expected: EventData{
				Type: "message_start",
			},
		},
		{
			name: "content_block_start with index",
			json: `{"type": "content_block_start", "index": 5}`,
			expected: EventData{
				Type:  "content_block_start",
				Index: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got EventData
			if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if got.Type != tt.expected.Type {
				t.Errorf("Type: got %q, want %q", got.Type, tt.expected.Type)
			}
			if got.Index != tt.expected.Index {
				t.Errorf("Index: got %d, want %d", got.Index, tt.expected.Index)
			}
		})
	}
}

func TestTextDelta_JSONUnmarshal(t *testing.T) {
	input := `{"type": "text_delta", "text": "Hello World"}`

	var delta TextDelta
	if err := json.Unmarshal([]byte(input), &delta); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if delta.Type != "text_delta" {
		t.Errorf("Type: got %q, want 'text_delta'", delta.Type)
	}
	if delta.Text != "Hello World" {
		t.Errorf("Text: got %q, want 'Hello World'", delta.Text)
	}
}

func TestInputJSONDelta_JSONUnmarshal(t *testing.T) {
	input := `{"type": "input_json_delta", "partial_json": "{\"key\": \"value\"}"}`

	var delta InputJSONDelta
	if err := json.Unmarshal([]byte(input), &delta); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if delta.Type != "input_json_delta" {
		t.Errorf("Type: got %q, want 'input_json_delta'", delta.Type)
	}
	if delta.PartialJSON != `{"key": "value"}` {
		t.Errorf("PartialJSON: got %q, want '{\"key\": \"value\"}'", delta.PartialJSON)
	}
}

func TestMessageDelta_JSONUnmarshal(t *testing.T) {
	input := `{"stop_reason": "end_turn", "stop_sequence": ""}`

	var delta MessageDelta
	if err := json.Unmarshal([]byte(input), &delta); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if delta.StopReason != "end_turn" {
		t.Errorf("StopReason: got %q, want 'end_turn'", delta.StopReason)
	}
}
