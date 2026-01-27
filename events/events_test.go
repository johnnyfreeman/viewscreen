package events

import (
	"encoding/json"
	"testing"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
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

func TestMatchToolResults_NoMatch(t *testing.T) {
	tracker := tools.NewToolUseTracker()

	event := user.Event{
		Message: user.Message{
			Content: []user.ToolResultContent{
				{Type: "tool_result", ToolUseID: "unknown-tool"},
			},
		},
	}

	matched := MatchToolResults(event, tracker)
	if len(matched) != 0 {
		t.Errorf("MatchToolResults should return empty slice for no matches, got %d", len(matched))
	}
}

func TestMatchToolResults_SingleMatch(t *testing.T) {
	tracker := tools.NewToolUseTracker()

	// Add a pending tool
	block := types.ContentBlock{
		Type: "tool_use",
		ID:   "tool-123",
		Name: "Bash",
	}
	tracker.Add("tool-123", block, nil)

	event := user.Event{
		Message: user.Message{
			Content: []user.ToolResultContent{
				{Type: "tool_result", ToolUseID: "tool-123"},
			},
		},
	}

	matched := MatchToolResults(event, tracker)
	if len(matched) != 1 {
		t.Fatalf("MatchToolResults should return 1 match, got %d", len(matched))
	}
	if matched[0].Block.ID != "tool-123" {
		t.Errorf("Matched block ID should be 'tool-123', got %q", matched[0].Block.ID)
	}
	if matched[0].IsNested {
		t.Error("Matched tool should not be nested")
	}

	// Tool should be removed from tracker
	if tracker.Len() != 0 {
		t.Error("Tracker should be empty after matching")
	}
}

func TestMatchToolResults_NestedTool(t *testing.T) {
	tracker := tools.NewToolUseTracker()

	// Add parent tool
	parentBlock := types.ContentBlock{
		Type: "tool_use",
		ID:   "parent-123",
		Name: "Task",
	}
	tracker.Add("parent-123", parentBlock, nil)

	// Add nested child tool
	parentID := "parent-123"
	childBlock := types.ContentBlock{
		Type: "tool_use",
		ID:   "child-456",
		Name: "Read",
	}
	tracker.Add("child-456", childBlock, &parentID)

	// Match the child tool result
	event := user.Event{
		Message: user.Message{
			Content: []user.ToolResultContent{
				{Type: "tool_result", ToolUseID: "child-456"},
			},
		},
	}

	matched := MatchToolResults(event, tracker)
	if len(matched) != 1 {
		t.Fatalf("MatchToolResults should return 1 match, got %d", len(matched))
	}
	if !matched[0].IsNested {
		t.Error("Matched tool should be nested when parent is still pending")
	}

	// Only child should be removed, parent still pending
	if tracker.Len() != 1 {
		t.Errorf("Tracker should have 1 remaining tool, got %d", tracker.Len())
	}
}

func TestBufferToolUse_BuffersNewTool(t *testing.T) {
	tracker := tools.NewToolUseTracker()
	streamRenderer := stream.NewRenderer()

	event := createAssistantEventWithToolUse("tool-123", "Bash")

	buffered := BufferToolUse(event, tracker, streamRenderer)
	if !buffered {
		t.Error("BufferToolUse should return true when tool is buffered")
	}
	if tracker.Len() != 1 {
		t.Errorf("Tracker should have 1 tool, got %d", tracker.Len())
	}

	pending, ok := tracker.Get("tool-123")
	if !ok {
		t.Fatal("Tracker should contain the buffered tool")
	}
	if pending.Block.Name != "Bash" {
		t.Errorf("Buffered tool Name should be 'Bash', got %q", pending.Block.Name)
	}
}

func TestBufferToolUse_IgnoresEmptyID(t *testing.T) {
	tracker := tools.NewToolUseTracker()
	streamRenderer := stream.NewRenderer()

	event := createAssistantEventWithToolUse("", "Bash")

	buffered := BufferToolUse(event, tracker, streamRenderer)
	if buffered {
		t.Error("BufferToolUse should return false when tool has empty ID")
	}
	if tracker.Len() != 0 {
		t.Errorf("Tracker should be empty, got %d tools", tracker.Len())
	}
}

func TestBufferToolUse_IgnoresNonToolUseBlocks(t *testing.T) {
	tracker := tools.NewToolUseTracker()
	streamRenderer := stream.NewRenderer()

	event := createAssistantEventWithTextBlock("hello world")

	buffered := BufferToolUse(event, tracker, streamRenderer)
	if buffered {
		t.Error("BufferToolUse should return false for non-tool_use blocks")
	}
	if tracker.Len() != 0 {
		t.Errorf("Tracker should be empty, got %d tools", tracker.Len())
	}
}

func TestFlushOrphanedTools_Empty(t *testing.T) {
	tracker := tools.NewToolUseTracker()

	orphaned := FlushOrphanedTools(tracker)
	if len(orphaned) != 0 {
		t.Errorf("FlushOrphanedTools should return empty slice for empty tracker, got %d", len(orphaned))
	}
}

func TestFlushOrphanedTools_ReturnsAllPending(t *testing.T) {
	tracker := tools.NewToolUseTracker()

	tracker.Add("tool-1", types.ContentBlock{ID: "tool-1", Name: "Bash"}, nil)
	tracker.Add("tool-2", types.ContentBlock{ID: "tool-2", Name: "Read"}, nil)

	orphaned := FlushOrphanedTools(tracker)
	if len(orphaned) != 2 {
		t.Fatalf("FlushOrphanedTools should return 2 orphaned tools, got %d", len(orphaned))
	}

	// Check that tracker is cleared
	if tracker.Len() != 0 {
		t.Errorf("Tracker should be empty after flush, got %d", tracker.Len())
	}

	// Check that all tools are returned (order not guaranteed)
	ids := make(map[string]bool)
	for _, o := range orphaned {
		ids[o.ID] = true
	}
	if !ids["tool-1"] {
		t.Error("Orphaned tools should include tool-1")
	}
	if !ids["tool-2"] {
		t.Error("Orphaned tools should include tool-2")
	}
}

func TestFlushOrphanedTools_NestedDetection(t *testing.T) {
	tracker := tools.NewToolUseTracker()

	// Add parent
	tracker.Add("parent", types.ContentBlock{ID: "parent", Name: "Task"}, nil)

	// Add nested child
	parentID := "parent"
	tracker.Add("child", types.ContentBlock{ID: "child", Name: "Read"}, &parentID)

	orphaned := FlushOrphanedTools(tracker)

	// Find parent and child in results
	var parentOrphan, childOrphan *OrphanedTool
	for i := range orphaned {
		if orphaned[i].ID == "parent" {
			parentOrphan = &orphaned[i]
		} else if orphaned[i].ID == "child" {
			childOrphan = &orphaned[i]
		}
	}

	if parentOrphan == nil {
		t.Fatal("Orphaned tools should include parent")
	}
	if childOrphan == nil {
		t.Fatal("Orphaned tools should include child")
	}

	if parentOrphan.IsNested {
		t.Error("Parent tool should not be nested")
	}
	if !childOrphan.IsNested {
		t.Error("Child tool should be nested")
	}
}

// Helper functions

func createAssistantEventWithToolUse(id, name string) assistant.Event {
	return assistant.Event{
		Message: assistant.Message{
			Content: []types.ContentBlock{
				{
					Type: "tool_use",
					ID:   id,
					Name: name,
				},
			},
		},
	}
}

func createAssistantEventWithTextBlock(text string) assistant.Event {
	return assistant.Event{
		Message: assistant.Message{
			Content: []types.ContentBlock{
				{
					Type: "text",
					Text: text,
				},
			},
		},
	}
}
