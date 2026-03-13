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

func TestParse_SubAgentSystemEvent(t *testing.T) {
	parentID := "tool-use-123"
	event := map[string]any{
		"type":               "system",
		"subtype":            "init",
		"parent_tool_use_id": parentID,
		"cwd":                "",
		"model":              "",
		"tools":              []string{},
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	subEvent, ok := result.(SubAgentSystemEvent)
	if !ok {
		t.Fatalf("Parse should return SubAgentSystemEvent when parent_tool_use_id is set, got %T", result)
	}
	if subEvent.Data.ParentToolUseID == nil || *subEvent.Data.ParentToolUseID != parentID {
		t.Errorf("SubAgentSystemEvent ParentToolUseID should be %q", parentID)
	}
}

func TestParse_SystemEvent_WithoutParentToolUseID(t *testing.T) {
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
	_, ok := result.(SystemEvent)
	if !ok {
		t.Fatalf("Parse should return SystemEvent when parent_tool_use_id is absent, got %T", result)
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

func TestParse_RateLimitEvent(t *testing.T) {
	result := Parse(`{"type":"rate_limit_event","rate_limit_info":{"status":"allowed","resetsAt":1773421200,"rateLimitType":"five_hour"},"uuid":"test-uuid","session_id":"test-session"}`)
	ignored, ok := result.(IgnoredEvent)
	if !ok {
		t.Fatalf("Parse should return IgnoredEvent for rate_limit_event, got %T", result)
	}
	if ignored.Type != "rate_limit_event" {
		t.Errorf("IgnoredEvent Type should be 'rate_limit_event', got %q", ignored.Type)
	}
}

func TestParse_SystemEvent_NewFields(t *testing.T) {
	event := map[string]any{
		"type":                "system",
		"subtype":             "init",
		"cwd":                 "/test",
		"model":               "claude-opus-4-6",
		"claude_code_version": "2.1.74",
		"tools":               []string{"Bash", "Read"},
		"agents":              []string{"general-purpose", "Explore"},
		"mcp_servers": []map[string]string{
			{"name": "test-server", "status": "connected"},
		},
		"slash_commands":  []string{"simplify", "loop"},
		"skills":          []string{"simplify"},
		"plugins":         []map[string]string{{"name": "gopls-lsp", "path": "/test/path"}},
		"fast_mode_state": "off",
		"apiKeySource":    "none",
		"output_style":    "default",
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	sysEvent, ok := result.(SystemEvent)
	if !ok {
		t.Fatalf("Parse should return SystemEvent, got %T", result)
	}
	if len(sysEvent.Data.MCPServers) != 1 || sysEvent.Data.MCPServers[0].Name != "test-server" {
		t.Errorf("MCPServers not parsed correctly: %+v", sysEvent.Data.MCPServers)
	}
	if len(sysEvent.Data.Plugins) != 1 || sysEvent.Data.Plugins[0].Name != "gopls-lsp" {
		t.Errorf("Plugins not parsed correctly: %+v", sysEvent.Data.Plugins)
	}
	if sysEvent.Data.FastModeState != "off" {
		t.Errorf("FastModeState should be 'off', got %q", sysEvent.Data.FastModeState)
	}
	if len(sysEvent.Data.Skills) != 1 || sysEvent.Data.Skills[0] != "simplify" {
		t.Errorf("Skills not parsed correctly: %+v", sysEvent.Data.Skills)
	}
	if len(sysEvent.Data.SlashCommands) != 2 {
		t.Errorf("SlashCommands not parsed correctly: %+v", sysEvent.Data.SlashCommands)
	}
}

func TestParse_ResultEvent_NewFields(t *testing.T) {
	event := map[string]any{
		"type":            "result",
		"subtype":         "success",
		"is_error":        false,
		"duration_ms":     5000,
		"duration_api_ms": 4800,
		"num_turns":       2,
		"result":          "done",
		"stop_reason":     "end_turn",
		"total_cost_usd":  0.05,
		"fast_mode_state": "off",
		"usage": map[string]any{
			"input_tokens":                100,
			"output_tokens":               50,
			"cache_creation_input_tokens": 10,
			"cache_read_input_tokens":     20,
			"service_tier":                "standard",
			"speed":                       "standard",
			"server_tool_use": map[string]any{
				"web_search_requests": 0,
				"web_fetch_requests":  0,
			},
		},
	}
	eventJSON, _ := json.Marshal(event)

	result := Parse(string(eventJSON))
	resultEvent, ok := result.(ResultEvent)
	if !ok {
		t.Fatalf("Parse should return ResultEvent, got %T", result)
	}
	if resultEvent.Data.StopReason != "end_turn" {
		t.Errorf("StopReason should be 'end_turn', got %q", resultEvent.Data.StopReason)
	}
	if resultEvent.Data.FastModeState != "off" {
		t.Errorf("FastModeState should be 'off', got %q", resultEvent.Data.FastModeState)
	}
	if resultEvent.Data.Usage.Speed != "standard" {
		t.Errorf("Usage.Speed should be 'standard', got %q", resultEvent.Data.Usage.Speed)
	}
	if resultEvent.Data.Usage.ServiceTier != "standard" {
		t.Errorf("Usage.ServiceTier should be 'standard', got %q", resultEvent.Data.Usage.ServiceTier)
	}
}
