package events

import (
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

func TestNewEventProcessor(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	if p == nil {
		t.Fatal("NewEventProcessor should return non-nil processor")
	}
	if p.Renderers() == nil {
		t.Error("Processor should have non-nil renderers")
	}
}

func TestEventProcessor_ProcessSystemEvent(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	event := SystemEvent{
		Data: system.Event{
			Subtype:          "init",
			CWD:              "/test/path",
			Model:            "test-model",
			ClaudeCodeVersion: "1.0.0",
		},
	}

	result := p.Process(event)

	// State should be updated
	if s.Model != "test-model" {
		t.Errorf("State model should be 'test-model', got %q", s.Model)
	}
	if s.CWD != "/test/path" {
		t.Errorf("State CWD should be '/test/path', got %q", s.CWD)
	}

	// Should have rendered output
	if result.Rendered == "" {
		t.Error("ProcessResult should have rendered output for system event")
	}
	if result.HasPendingTools {
		t.Error("System event should not have pending tools")
	}
}

func TestEventProcessor_ProcessAssistantEvent_TextOnly(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	event := AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{Type: "text", Text: "Hello world"},
				},
			},
		},
	}

	result := p.Process(event)

	// Turn count should be incremented
	if s.TurnCount != 1 {
		t.Errorf("TurnCount should be 1, got %d", s.TurnCount)
	}

	// Should have rendered output
	if result.Rendered == "" {
		t.Error("ProcessResult should have rendered output for assistant text")
	}
	if result.HasPendingTools {
		t.Error("Text-only assistant event should not have pending tools")
	}
}

func TestEventProcessor_ProcessAssistantEvent_WithToolUse(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	event := AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ID:   "tool-123",
						Name: "Bash",
					},
				},
			},
		},
	}

	result := p.Process(event)

	// Should have pending tools
	if !result.HasPendingTools {
		t.Error("ProcessResult should have pending tools after tool_use")
	}
	if !p.HasPendingTools() {
		t.Error("Processor should report pending tools")
	}

	// State should show current tool
	if s.CurrentTool != "Bash" {
		t.Errorf("State CurrentTool should be 'Bash', got %q", s.CurrentTool)
	}
}

func TestEventProcessor_ProcessUserEvent_MatchesTool(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// First, process an assistant event with tool_use
	assistantEvent := AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ID:   "tool-123",
						Name: "Bash",
					},
				},
			},
		},
	}
	p.Process(assistantEvent)

	// Now process the tool result
	userEvent := UserEvent{
		Data: user.Event{
			Message: user.Message{
				Content: []user.ToolResultContent{
					{
						Type:      "tool_result",
						ToolUseID: "tool-123",
					},
				},
			},
		},
	}

	result := p.Process(userEvent)

	// Pending tools should be cleared
	if result.HasPendingTools {
		t.Error("ProcessResult should not have pending tools after matching")
	}
	if p.HasPendingTools() {
		t.Error("Processor should not have pending tools after matching")
	}

	// Should have rendered output (tool header + result)
	if result.Rendered == "" {
		t.Error("ProcessResult should have rendered output for user event")
	}
}

func TestEventProcessor_ProcessResultEvent(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	event := ResultEvent{
		Data: result.Event{
			Subtype:    "success",
			NumTurns:   5,
			TotalCostUSD: 0.05,
			DurationMS: 1000,
		},
	}

	processResult := p.Process(event)

	// State should be updated with result info
	if s.TurnCount != 5 {
		t.Errorf("State TurnCount should be 5, got %d", s.TurnCount)
	}
	if s.TotalCost != 0.05 {
		t.Errorf("State TotalCost should be 0.05, got %f", s.TotalCost)
	}
	if s.DurationMS != 1000 {
		t.Errorf("State DurationMS should be 1000, got %d", s.DurationMS)
	}

	// Should have rendered output
	if processResult.Rendered == "" {
		t.Error("ProcessResult should have rendered output for result event")
	}
}

func TestEventProcessor_ProcessResultEvent_FlushesOrphanedTools(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// Add a tool that will be orphaned
	assistantEvent := AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ID:   "orphan-tool",
						Name: "Read",
					},
				},
			},
		},
	}
	p.Process(assistantEvent)

	// Now process result without matching the tool
	resultEvent := ResultEvent{
		Data: result.Event{
			Subtype: "success",
		},
	}

	processResult := p.Process(resultEvent)

	// Orphaned tool should be rendered
	if !strings.Contains(processResult.Rendered, "Read") {
		t.Error("Result should contain orphaned tool name 'Read'")
	}
	if !strings.Contains(processResult.Rendered, "(no result)") {
		t.Error("Result should contain '(no result)' for orphaned tools")
	}

	// No more pending tools
	if p.HasPendingTools() {
		t.Error("Processor should not have pending tools after result event")
	}
}

func TestEventProcessor_ProcessUnknownEvent(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// Process nil or unknown event type
	result := p.Process(nil)

	if result.Rendered != "" {
		t.Error("ProcessResult should have empty rendered for nil event")
	}
	if result.HasPendingTools {
		t.Error("ProcessResult should not have pending tools for nil event")
	}
}

func TestEventProcessor_ForEachPendingTool(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// Add a pending tool
	event := AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ID:   "tool-abc",
						Name: "Bash",
					},
				},
			},
		},
	}
	p.Process(event)

	// Iterate over pending tools
	var foundTool bool
	p.ForEachPendingTool(func(id string, pending tools.PendingTool) {
		if id == "tool-abc" {
			foundTool = true
		}
	})

	if !foundTool {
		t.Error("ForEachPendingTool should iterate over pending tool")
	}
}

func TestEventProcessor_RenderPendingTool(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// Add a pending tool
	event := AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ID:   "tool-xyz",
						Name: "Read",
					},
				},
			},
		},
	}
	p.Process(event)

	// Render the pending tool with a custom icon
	var rendered string
	p.ForEachPendingTool(func(id string, pending tools.PendingTool) {
		rendered = p.RenderPendingTool(pending, "* ")
	})

	// Should have rendered output containing the tool name
	if rendered == "" {
		t.Error("RenderPendingTool should return non-empty string")
	}
	if !strings.Contains(rendered, "Read") {
		t.Error("RenderPendingTool output should contain tool name 'Read'")
	}
}

func TestEventProcessor_MultipleTurns(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// Simulate multiple turns
	for i := 0; i < 3; i++ {
		event := AssistantEvent{
			Data: assistant.Event{
				Message: assistant.Message{
					Content: []types.ContentBlock{
						{Type: "text", Text: "Turn content"},
					},
				},
			},
		}
		p.Process(event)
	}

	if s.TurnCount != 3 {
		t.Errorf("TurnCount should be 3 after 3 assistant events, got %d", s.TurnCount)
	}
}
