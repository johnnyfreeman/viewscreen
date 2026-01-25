package parser

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/types"
)

func TestNewParser(t *testing.T) {
	p := NewParser()

	if p == nil {
		t.Fatal("NewParser returned nil")
	}

	if p.streamRenderer == nil {
		t.Error("expected streamRenderer to be non-nil")
	}
}

func TestNewParserWithOptions(t *testing.T) {
	t.Run("with custom input", func(t *testing.T) {
		input := strings.NewReader("")
		p := NewParserWithOptions(WithInput(input))

		if p.input != input {
			t.Error("expected custom input reader")
		}
	})

	t.Run("with custom error output", func(t *testing.T) {
		errOut := &bytes.Buffer{}
		p := NewParserWithOptions(WithErrOutput(errOut))

		if p.errOutput != errOut {
			t.Error("expected custom error output writer")
		}
	})

	t.Run("with custom stream renderer", func(t *testing.T) {
		sr := stream.NewRenderer()
		p := NewParserWithOptions(WithStreamRenderer(sr))

		if p.streamRenderer != sr {
			t.Error("expected custom stream renderer")
		}
	})

	t.Run("with event handler", func(t *testing.T) {
		handler := func(eventType string, line []byte) error {
			return nil
		}
		p := NewParserWithOptions(WithEventHandler(handler))

		if p.eventHandler == nil {
			t.Error("expected event handler to be set")
		}
	})
}

func TestParser_Run_EmptyInput(t *testing.T) {
	p := NewParserWithOptions(
		WithInput(strings.NewReader("")),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("expected no error for empty input, got: %v", err)
	}
}

func TestParser_Run_SkipsEmptyLines(t *testing.T) {
	input := "\n\n\n"
	var eventsCalled int
	p := NewParserWithOptions(
		WithInput(strings.NewReader(input)),
		WithEventHandler(func(eventType string, line []byte) error {
			eventsCalled++
			return nil
		}),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if eventsCalled != 0 {
		t.Errorf("expected no events for empty lines, got %d", eventsCalled)
	}
}

func TestParser_Run_InvalidJSON(t *testing.T) {
	input := "this is not json"
	errOut := &bytes.Buffer{}

	p := NewParserWithOptions(
		WithInput(strings.NewReader(input)),
		WithErrOutput(errOut),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(errOut.String(), "Error parsing JSON") {
		t.Errorf("expected error message in stderr, got: %s", errOut.String())
	}
}

func TestParser_Run_UnknownEventType(t *testing.T) {
	event := map[string]any{
		"type": "unknown_event_type",
	}
	eventJSON, _ := json.Marshal(event)
	errOut := &bytes.Buffer{}

	p := NewParserWithOptions(
		WithInput(strings.NewReader(string(eventJSON))),
		WithErrOutput(errOut),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(errOut.String(), "Unknown event type: unknown_event_type") {
		t.Errorf("expected unknown event type message in stderr, got: %s", errOut.String())
	}
}

func TestParser_Run_EventHandlerCalled(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		eventPayload map[string]any
	}{
		{
			name:      "system event",
			eventType: "system",
			eventPayload: map[string]any{
				"type":               "system",
				"subtype":            "init",
				"cwd":                "/test",
				"model":              "test-model",
				"claude_code_version": "1.0.0",
				"tools":              []string{},
			},
		},
		{
			name:      "assistant event",
			eventType: "assistant",
			eventPayload: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"id":      "msg_123",
					"type":    "message",
					"role":    "assistant",
					"model":   "test-model",
					"content": []any{},
				},
			},
		},
		{
			name:      "user event",
			eventType: "user",
			eventPayload: map[string]any{
				"type": "user",
				"message": map[string]any{
					"role":    "user",
					"content": []any{},
				},
			},
		},
		{
			name:      "stream_event",
			eventType: "stream_event",
			eventPayload: map[string]any{
				"type": "stream_event",
				"event": map[string]any{
					"type": "message_start",
				},
			},
		},
		{
			name:      "result event",
			eventType: "result",
			eventPayload: map[string]any{
				"type":        "result",
				"subtype":     "success",
				"is_error":    false,
				"duration_ms": 100,
				"result":      "test result",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventJSON, _ := json.Marshal(tt.eventPayload)

			var capturedType string
			var capturedLine []byte

			// Use a custom output buffer for stream renderer to suppress output
			outputBuf := &bytes.Buffer{}
			sr := stream.NewRendererWithOptions(stream.WithOutput(outputBuf))

			p := NewParserWithOptions(
				WithInput(strings.NewReader(string(eventJSON))),
				WithErrOutput(io.Discard),
				WithStreamRenderer(sr),
				WithEventHandler(func(eventType string, line []byte) error {
					capturedType = eventType
					capturedLine = line
					return nil
				}),
			)

			err := p.Run()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if capturedType != tt.eventType {
				t.Errorf("expected event type %q, got %q", tt.eventType, capturedType)
			}

			if len(capturedLine) == 0 {
				t.Error("expected captured line to be non-empty")
			}
		})
	}
}

func TestParser_Run_MultipleEvents(t *testing.T) {
	events := []map[string]any{
		{
			"type":               "system",
			"subtype":            "init",
			"cwd":                "/test",
			"model":              "test-model",
			"claude_code_version": "1.0.0",
			"tools":              []string{},
		},
		{
			"type": "result",
			"subtype": "success",
			"is_error": false,
			"duration_ms": 100,
			"result": "done",
		},
	}

	var lines []string
	for _, e := range events {
		j, _ := json.Marshal(e)
		lines = append(lines, string(j))
	}
	input := strings.Join(lines, "\n")

	var capturedTypes []string
	p := NewParserWithOptions(
		WithInput(strings.NewReader(input)),
		WithErrOutput(io.Discard),
		WithEventHandler(func(eventType string, line []byte) error {
			capturedTypes = append(capturedTypes, eventType)
			return nil
		}),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(capturedTypes) != 2 {
		t.Errorf("expected 2 events, got %d", len(capturedTypes))
	}

	if capturedTypes[0] != "system" {
		t.Errorf("expected first event type 'system', got %q", capturedTypes[0])
	}

	if capturedTypes[1] != "result" {
		t.Errorf("expected second event type 'result', got %q", capturedTypes[1])
	}
}

func TestParser_Run_EventHandlerError(t *testing.T) {
	event := map[string]any{
		"type":               "system",
		"subtype":            "init",
		"cwd":                "/test",
		"model":              "test-model",
		"claude_code_version": "1.0.0",
		"tools":              []string{},
	}
	eventJSON, _ := json.Marshal(event)

	handlerError := io.EOF // Use a recognizable error
	p := NewParserWithOptions(
		WithInput(strings.NewReader(string(eventJSON))),
		WithErrOutput(io.Discard),
		WithEventHandler(func(eventType string, line []byte) error {
			return handlerError
		}),
	)

	err := p.Run()
	if err != handlerError {
		t.Errorf("expected handler error to propagate, got: %v", err)
	}
}

func TestParser_Run_ParseErrorsForEachEventType(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		invalidJSON  string
		expectedErr  string
	}{
		{
			name:        "invalid system event",
			eventType:   "system",
			invalidJSON: `{"type":"system","subtype":123}`,
			expectedErr: "Error parsing system event",
		},
		{
			name:        "invalid assistant event",
			eventType:   "assistant",
			invalidJSON: `{"type":"assistant","message":"not_an_object"}`,
			expectedErr: "Error parsing assistant event",
		},
		{
			name:        "invalid user event",
			eventType:   "user",
			invalidJSON: `{"type":"user","message":"not_an_object"}`,
			expectedErr: "Error parsing user event",
		},
		{
			name:        "invalid stream_event",
			eventType:   "stream_event",
			invalidJSON: `{"type":"stream_event","event":"not_an_object"}`,
			expectedErr: "Error parsing stream event",
		},
		{
			name:        "invalid result event",
			eventType:   "result",
			invalidJSON: `{"type":"result","duration_ms":"not_a_number"}`,
			expectedErr: "Error parsing result event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errOut := &bytes.Buffer{}
			outputBuf := &bytes.Buffer{}
			sr := stream.NewRendererWithOptions(stream.WithOutput(outputBuf))

			p := NewParserWithOptions(
				WithInput(strings.NewReader(tt.invalidJSON)),
				WithErrOutput(errOut),
				WithStreamRenderer(sr),
			)

			err := p.Run()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !strings.Contains(errOut.String(), tt.expectedErr) {
				t.Errorf("expected error message %q in stderr, got: %s", tt.expectedErr, errOut.String())
			}
		})
	}
}

func TestParser_Run_StreamRendererIntegration(t *testing.T) {
	// Test that stream events properly update stream renderer state
	events := []map[string]any{
		{
			"type": "stream_event",
			"event": map[string]any{
				"type":  "content_block_start",
				"index": 0,
				"content_block": map[string]any{
					"type": "text",
				},
			},
		},
	}

	var lines []string
	for _, e := range events {
		j, _ := json.Marshal(e)
		lines = append(lines, string(j))
	}
	input := strings.Join(lines, "\n")

	outputBuf := &bytes.Buffer{}
	sr := stream.NewRendererWithOptions(stream.WithOutput(outputBuf))

	p := NewParserWithOptions(
		WithInput(strings.NewReader(input)),
		WithErrOutput(io.Discard),
		WithStreamRenderer(sr),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// After processing content_block_start for text, InTextBlock should be true
	if !sr.InTextBlock {
		t.Error("expected InTextBlock to be true after text block start")
	}
}

func TestParser_Run_AssistantResetsBlockState(t *testing.T) {
	// Set up stream renderer with active blocks
	outputBuf := &bytes.Buffer{}
	sr := stream.NewRendererWithOptions(stream.WithOutput(outputBuf))
	sr.InTextBlock = true
	sr.InToolUseBlock = true

	// Send an assistant event
	event := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"id":      "msg_123",
			"type":    "message",
			"role":    "assistant",
			"model":   "test-model",
			"content": []any{},
		},
	}
	eventJSON, _ := json.Marshal(event)

	p := NewParserWithOptions(
		WithInput(strings.NewReader(string(eventJSON))),
		WithErrOutput(io.Discard),
		WithStreamRenderer(sr),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// After assistant event, block state should be reset
	if sr.InTextBlock {
		t.Error("expected InTextBlock to be false after assistant event")
	}
	if sr.InToolUseBlock {
		t.Error("expected InToolUseBlock to be false after assistant event")
	}
}

func TestParser_Run_LargeInput(t *testing.T) {
	// Test that parser can handle large JSON lines (up to 10MB buffer)
	largeContent := strings.Repeat("x", 100000) // 100KB of content
	event := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"id":    "msg_123",
			"type":  "message",
			"role":  "assistant",
			"model": "test-model",
			"content": []map[string]any{
				{
					"type": "text",
					"text": largeContent,
				},
			},
		},
	}
	eventJSON, _ := json.Marshal(event)

	var capturedType string
	p := NewParserWithOptions(
		WithInput(strings.NewReader(string(eventJSON))),
		WithErrOutput(io.Discard),
		WithEventHandler(func(eventType string, line []byte) error {
			capturedType = eventType
			return nil
		}),
	)

	err := p.Run()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if capturedType != "assistant" {
		t.Errorf("expected event type 'assistant', got %q", capturedType)
	}
}

func TestBaseEvent_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected types.BaseEvent
	}{
		{
			name:     "system event",
			json:     `{"type": "system"}`,
			expected: types.BaseEvent{Type: "system"},
		},
		{
			name:     "assistant event",
			json:     `{"type": "assistant"}`,
			expected: types.BaseEvent{Type: "assistant"},
		},
		{
			name:     "user event",
			json:     `{"type": "user"}`,
			expected: types.BaseEvent{Type: "user"},
		},
		{
			name:     "stream_event",
			json:     `{"type": "stream_event"}`,
			expected: types.BaseEvent{Type: "stream_event"},
		},
		{
			name:     "result event",
			json:     `{"type": "result"}`,
			expected: types.BaseEvent{Type: "result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got types.BaseEvent
			if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if got.Type != tt.expected.Type {
				t.Errorf("Type: got %q, want %q", got.Type, tt.expected.Type)
			}
		})
	}
}
