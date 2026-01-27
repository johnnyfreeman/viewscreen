package events

import (
	"encoding/json"
	"testing"
)

func TestParse_EmptyLine(t *testing.T) {
	result := Parse("")
	if result != nil {
		t.Error("Parse should return nil for empty line")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	result := Parse("not valid json")
	parseErr, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for invalid JSON")
	}
	if parseErr.Err == nil {
		t.Error("ParseError should have non-nil Err for invalid JSON")
	}
	if parseErr.Line != "not valid json" {
		t.Errorf("ParseError Line should be original line, got %q", parseErr.Line)
	}
}

func TestParse_UnknownEventType(t *testing.T) {
	result := Parse(`{"type":"unknown"}`)
	parseErr, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for unknown event type")
	}
	if parseErr.Line != "Unknown event type: unknown" {
		t.Errorf("ParseError Line should describe unknown type, got %q", parseErr.Line)
	}
}

func TestParse_SystemEvent(t *testing.T) {
	event := map[string]any{
		"type":                "system",
		"subtype":             "init",
		"cwd":                 "/test",
		"model":               "test-model",
		"claude_code_version": "1.0.0",
		"tools":               []string{},
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	sysEvent, ok := result.(SystemEvent)
	if !ok {
		t.Fatalf("Parse should return SystemEvent, got %T", result)
	}
	if sysEvent.Data.Subtype != "init" {
		t.Errorf("SystemEvent Subtype should be 'init', got %q", sysEvent.Data.Subtype)
	}
}

func TestParse_AssistantEvent(t *testing.T) {
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

	result := Parse(string(eventJSON))
	asstEvent, ok := result.(AssistantEvent)
	if !ok {
		t.Fatalf("Parse should return AssistantEvent, got %T", result)
	}
	if asstEvent.Data.Message.ID != "msg_123" {
		t.Errorf("AssistantEvent Message.ID should be 'msg_123', got %q", asstEvent.Data.Message.ID)
	}
}

func TestParse_UserEvent(t *testing.T) {
	event := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": []any{},
		},
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	userEvent, ok := result.(UserEvent)
	if !ok {
		t.Fatalf("Parse should return UserEvent, got %T", result)
	}
	if userEvent.Data.Message.Role != "user" {
		t.Errorf("UserEvent Message.Role should be 'user', got %q", userEvent.Data.Message.Role)
	}
}

func TestParse_StreamEvent(t *testing.T) {
	event := map[string]any{
		"type": "stream_event",
		"event": map[string]any{
			"type": "message_start",
		},
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	streamEvent, ok := result.(StreamEvent)
	if !ok {
		t.Fatalf("Parse should return StreamEvent, got %T", result)
	}
	if streamEvent.Data.Event.Type != "message_start" {
		t.Errorf("StreamEvent Event.Type should be 'message_start', got %q", streamEvent.Data.Event.Type)
	}
}

func TestParse_ResultEvent(t *testing.T) {
	event := map[string]any{
		"type":        "result",
		"subtype":     "success",
		"is_error":    false,
		"duration_ms": 100,
		"result":      "test result",
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	resultEvent, ok := result.(ResultEvent)
	if !ok {
		t.Fatalf("Parse should return ResultEvent, got %T", result)
	}
	if resultEvent.Data.Subtype != "success" {
		t.Errorf("ResultEvent Subtype should be 'success', got %q", resultEvent.Data.Subtype)
	}
}

func TestParse_InvalidSystemEvent(t *testing.T) {
	// Invalid because subtype should be a string, not a number
	result := Parse(`{"type":"system","subtype":123}`)
	_, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for invalid system event")
	}
}

func TestParse_InvalidAssistantEvent(t *testing.T) {
	result := Parse(`{"type":"assistant","message":"not_an_object"}`)
	_, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for invalid assistant event")
	}
}

func TestParse_InvalidUserEvent(t *testing.T) {
	result := Parse(`{"type":"user","message":"not_an_object"}`)
	_, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for invalid user event")
	}
}

func TestParse_InvalidStreamEvent(t *testing.T) {
	result := Parse(`{"type":"stream_event","event":"not_an_object"}`)
	_, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for invalid stream event")
	}
}

func TestParse_InvalidResultEvent(t *testing.T) {
	result := Parse(`{"type":"result","duration_ms":"not_a_number"}`)
	_, ok := result.(ParseError)
	if !ok {
		t.Fatal("Parse should return ParseError for invalid result event")
	}
}
