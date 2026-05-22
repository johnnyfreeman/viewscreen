package state

import (
	"testing"

	"github.com/johnnyfreeman/viewscreen/timeline"
)

func TestApplyPatchReplacesTodosAndActivity(t *testing.T) {
	s := NewState()
	activity := timeline.Activity{Name: "Read", Input: "README.md"}

	s.ApplyPatch(timeline.StatePatch{
		IncrementTurns:  1,
		CurrentActivity: &activity,
		ReplaceTodos:    true,
		Todos: []timeline.Todo{
			{Content: "write tests", Status: "in_progress"},
		},
		AddUsage: &timeline.Usage{InputTokens: 10, OutputTokens: 2, CacheRead: 3, ReasoningTokens: 4},
	})

	if s.TurnCount != 1 || s.CurrentTool != "Read" || s.CurrentToolInput != "README.md" {
		t.Fatalf("state patch did not update turn/activity: %+v", s)
	}
	if len(s.Todos) != 1 || s.Todos[0].Content != "write tests" {
		t.Fatalf("todos = %+v, want replacement todo", s.Todos)
	}
	if s.InputTokens != 10 || s.OutputTokens != 2 || s.CacheRead != 3 || s.ReasoningTokens != 4 {
		t.Fatalf("usage = in:%d out:%d cache:%d reasoning:%d", s.InputTokens, s.OutputTokens, s.CacheRead, s.ReasoningTokens)
	}

	s.ApplyPatch(timeline.StatePatch{ClearActivity: true})
	if s.ToolInProgress {
		t.Fatalf("ToolInProgress = true, want false")
	}
}
