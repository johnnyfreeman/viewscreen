package state

import (
	"testing"
	"time"

	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/system"
)

func TestNewState(t *testing.T) {
	s := NewState()
	if s == nil {
		t.Fatal("NewState should return non-nil")
	}
	if s.InputTokens != 0 {
		t.Errorf("InputTokens should be 0, got %d", s.InputTokens)
	}
	if s.OutputTokens != 0 {
		t.Errorf("OutputTokens should be 0, got %d", s.OutputTokens)
	}
}

func TestAccumulateUsage(t *testing.T) {
	s := NewState()

	s.AccumulateUsage(100, 50, 200, 300)
	if s.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", s.InputTokens)
	}
	if s.OutputTokens != 50 {
		t.Errorf("OutputTokens = %d, want 50", s.OutputTokens)
	}
	if s.CacheCreated != 200 {
		t.Errorf("CacheCreated = %d, want 200", s.CacheCreated)
	}
	if s.CacheRead != 300 {
		t.Errorf("CacheRead = %d, want 300", s.CacheRead)
	}

	// Accumulate more
	s.AccumulateUsage(50, 25, 0, 100)
	if s.InputTokens != 150 {
		t.Errorf("InputTokens = %d, want 150", s.InputTokens)
	}
	if s.OutputTokens != 75 {
		t.Errorf("OutputTokens = %d, want 75", s.OutputTokens)
	}
	if s.CacheCreated != 200 {
		t.Errorf("CacheCreated = %d, want 200", s.CacheCreated)
	}
	if s.CacheRead != 400 {
		t.Errorf("CacheRead = %d, want 400", s.CacheRead)
	}
}

func TestAccumulateUsage_ThenResultOverrides(t *testing.T) {
	s := NewState()

	// Accumulate from per-turn usage
	s.AccumulateUsage(100, 50, 200, 300)

	// Result event should override with authoritative totals
	s.UpdateFromResultEvent(result.Event{
		NumTurns:     5,
		TotalCostUSD: 0.05,
		Usage: result.Usage{
			InputTokens:              500,
			OutputTokens:             200,
			CacheCreationInputTokens: 1000,
			CacheReadInputTokens:     2000,
		},
	})

	if s.InputTokens != 500 {
		t.Errorf("InputTokens = %d, want 500 (should be overridden by result)", s.InputTokens)
	}
	if s.OutputTokens != 200 {
		t.Errorf("OutputTokens = %d, want 200 (should be overridden by result)", s.OutputTokens)
	}
}

func TestUpdateFromSystemEvent_EmptyFieldsDoNotOverwrite(t *testing.T) {
	s := NewState()
	// Set initial state
	s.Model = "parent-model"
	s.Version = "2.0.0"
	s.CWD = "/parent/cwd"
	s.ToolsCount = 10
	s.Agents = []string{"agent1"}
	s.PermissionMode = "auto"

	// Update with empty event (like a subagent system event)
	s.UpdateFromSystemEvent(system.Event{})

	if s.Model != "parent-model" {
		t.Errorf("Model should not be overwritten by empty value, got %q", s.Model)
	}
	if s.Version != "2.0.0" {
		t.Errorf("Version should not be overwritten by empty value, got %q", s.Version)
	}
	if s.CWD != "/parent/cwd" {
		t.Errorf("CWD should not be overwritten by empty value, got %q", s.CWD)
	}
	if s.ToolsCount != 10 {
		t.Errorf("ToolsCount should not be overwritten by empty tools, got %d", s.ToolsCount)
	}
	if len(s.Agents) != 1 || s.Agents[0] != "agent1" {
		t.Errorf("Agents should not be overwritten by empty value, got %v", s.Agents)
	}
	if s.PermissionMode != "auto" {
		t.Errorf("PermissionMode should not be overwritten by empty value, got %q", s.PermissionMode)
	}
}

func TestUpdateFromSystemEvent_NonEmptyFieldsDoOverwrite(t *testing.T) {
	s := NewState()
	s.Model = "old-model"
	s.Version = "1.0.0"
	s.CWD = "/old/path"
	s.ToolsCount = 5

	s.UpdateFromSystemEvent(system.Event{
		Model:             "new-model",
		ClaudeCodeVersion: "3.0.0",
		CWD:               "/new/path",
		Tools:             []string{"tool1", "tool2"},
		Agents:            []string{"agent-new"},
		PermissionMode:    "manual",
	})

	if s.Model != "new-model" {
		t.Errorf("Model should be updated to 'new-model', got %q", s.Model)
	}
	if s.Version != "3.0.0" {
		t.Errorf("Version should be updated to '3.0.0', got %q", s.Version)
	}
	if s.CWD != "/new/path" {
		t.Errorf("CWD should be updated to '/new/path', got %q", s.CWD)
	}
	if s.ToolsCount != 2 {
		t.Errorf("ToolsCount should be 2, got %d", s.ToolsCount)
	}
	if s.PermissionMode != "manual" {
		t.Errorf("PermissionMode should be 'manual', got %q", s.PermissionMode)
	}
}

func TestNewState_HasStartTime(t *testing.T) {
	before := time.Now()
	s := NewState()
	after := time.Now()

	if s.StartTime.Before(before) || s.StartTime.After(after) {
		t.Errorf("StartTime = %v, expected between %v and %v", s.StartTime, before, after)
	}
}

func TestState_Elapsed(t *testing.T) {
	s := NewState()
	// Set start time to a known value in the past
	s.StartTime = time.Now().Add(-5 * time.Second)

	elapsed := s.Elapsed()
	if elapsed < 5*time.Second {
		t.Errorf("Elapsed() = %v, expected >= 5s", elapsed)
	}
	if elapsed > 6*time.Second {
		t.Errorf("Elapsed() = %v, expected < 6s", elapsed)
	}
}

func TestState_CostRate(t *testing.T) {
	t.Run("returns zero when elapsed less than 1 second", func(t *testing.T) {
		s := NewState()
		s.TotalCost = 0.10
		// StartTime is now, so elapsed < 1s
		rate := s.CostRate()
		if rate != 0 {
			t.Errorf("CostRate() = %f, want 0 for < 1s elapsed", rate)
		}
	})

	t.Run("returns zero when no cost", func(t *testing.T) {
		s := NewState()
		s.StartTime = time.Now().Add(-5 * time.Minute)
		s.TotalCost = 0
		rate := s.CostRate()
		if rate != 0 {
			t.Errorf("CostRate() = %f, want 0", rate)
		}
	})

	t.Run("calculates correct rate", func(t *testing.T) {
		s := NewState()
		s.StartTime = time.Now().Add(-2 * time.Minute)
		s.TotalCost = 0.10
		rate := s.CostRate()
		// Should be approximately $0.05/min
		if rate < 0.04 || rate > 0.06 {
			t.Errorf("CostRate() = %f, want ~0.05", rate)
		}
	})
}

func TestState_TodoProgress(t *testing.T) {
	t.Run("empty todos", func(t *testing.T) {
		s := NewState()
		completed, total := s.TodoProgress()
		if completed != 0 || total != 0 {
			t.Errorf("TodoProgress() = (%d, %d), want (0, 0)", completed, total)
		}
	})

	t.Run("mixed statuses", func(t *testing.T) {
		s := NewState()
		s.Todos = []Todo{
			{Content: "A", Status: "completed"},
			{Content: "B", Status: "in_progress"},
			{Content: "C", Status: "completed"},
			{Content: "D", Status: "pending"},
		}
		completed, total := s.TodoProgress()
		if completed != 2 || total != 4 {
			t.Errorf("TodoProgress() = (%d, %d), want (2, 4)", completed, total)
		}
	})

	t.Run("all completed", func(t *testing.T) {
		s := NewState()
		s.Todos = []Todo{
			{Content: "A", Status: "completed"},
			{Content: "B", Status: "completed"},
		}
		completed, total := s.TodoProgress()
		if completed != 2 || total != 2 {
			t.Errorf("TodoProgress() = (%d, %d), want (2, 2)", completed, total)
		}
	})

	t.Run("none completed", func(t *testing.T) {
		s := NewState()
		s.Todos = []Todo{
			{Content: "A", Status: "pending"},
			{Content: "B", Status: "in_progress"},
		}
		completed, total := s.TodoProgress()
		if completed != 0 || total != 2 {
			t.Errorf("TodoProgress() = (%d, %d), want (0, 2)", completed, total)
		}
	})
}
