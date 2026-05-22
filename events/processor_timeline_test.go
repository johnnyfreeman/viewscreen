package events

import (
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/codex"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/types"
)

func TestEventProcessorReturnsTimelineEntries(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	result := p.Process(AssistantEvent{Data: assistant.Event{
		Message: assistant.Message{Content: []types.ContentBlock{{Type: "text", Text: "timeline text"}}},
	}})

	if len(result.Batch.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(result.Batch.Entries))
	}
	if !strings.Contains(result.Batch.Entries[0].Text(), "timeline text") {
		t.Fatalf("entry text = %q, want assistant text", result.Batch.Entries[0].Text())
	}
	if s.TurnCount != 1 {
		t.Fatalf("TurnCount = %d, want 1", s.TurnCount)
	}
	if result.Batch.Patch.IncrementTurns != 1 {
		t.Fatalf("batch patch IncrementTurns = %d, want 1", result.Batch.Patch.IncrementTurns)
	}
}

func TestEventProcessorCodexOverlappingActivities(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &codex.Item{
		ID: "one", Type: codex.ItemCommandExecution, Command: "sh -lc 'first'",
	}}})
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &codex.Item{
		ID: "two", Type: codex.ItemWebSearch, Query: "second",
	}}})

	if got := len(p.PendingActivities()); got != 2 {
		t.Fatalf("PendingActivities = %d, want 2", got)
	}
	if s.CurrentTool != "Web Search" {
		t.Fatalf("CurrentTool = %q, want Web Search", s.CurrentTool)
	}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &codex.Item{
		ID: "two", Type: codex.ItemWebSearch, Query: "second",
	}}})

	if got := len(p.PendingActivities()); got != 1 {
		t.Fatalf("PendingActivities after completion = %d, want 1", got)
	}
	if s.CurrentTool != "Shell" {
		t.Fatalf("CurrentTool after completion = %q, want Shell", s.CurrentTool)
	}
}
