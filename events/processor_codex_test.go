package events

import (
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/codex"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/system"
)

func TestProcessCodex_RendersAndAccumulatesUsage(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	// thread.started should produce visible output.
	res := p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeThreadStarted, ThreadID: "t1"}})
	if !strings.Contains(res.Rendered, "Codex Session") {
		t.Errorf("thread.started rendered %q, want it to mention Codex Session", res.Rendered)
	}

	// turn.completed should fold token usage into shared state for the sidebar.
	usage := &codex.Usage{InputTokens: 100, OutputTokens: 20, CachedInputTokens: 40, ReasoningOutputTokens: 12}
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeTurnCompleted, Usage: usage}})

	if s.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", s.InputTokens)
	}
	if s.OutputTokens != 20 {
		t.Errorf("OutputTokens = %d, want 20", s.OutputTokens)
	}
	if s.CacheRead != 40 {
		t.Errorf("CacheRead = %d, want 40", s.CacheRead)
	}
	if s.ReasoningTokens != 12 {
		t.Errorf("ReasoningTokens = %d, want 12", s.ReasoningTokens)
	}

	// Codex reports no dollar cost, so the sidebar must not show a cost field.
	if s.ReportsCost() {
		t.Error("ReportsCost() = true for codex stream, want false")
	}
}

func TestEventProcessor_DetectsAgent(t *testing.T) {
	t.Run("codex events brand the session as codex", func(t *testing.T) {
		s := state.NewState()
		p := NewEventProcessor(s)

		p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeThreadStarted, ThreadID: "t1"}})

		if s.Agent != config.AgentCodex {
			t.Errorf("Agent = %q, want %q", s.Agent, config.AgentCodex)
		}
	})

	t.Run("claude stream-json events brand the session as claude", func(t *testing.T) {
		s := state.NewState()
		p := NewEventProcessor(s)

		p.Process(SystemEvent{Data: system.Event{Subtype: "init", Model: "test-model"}})

		if s.Agent != config.AgentClaude {
			t.Errorf("Agent = %q, want %q", s.Agent, config.AgentClaude)
		}
	})

	t.Run("seeded agent survives until a definitive event arrives", func(t *testing.T) {
		s := state.NewState()
		s.Agent = config.AgentCodex // seeded in prompt mode
		p := NewEventProcessor(s)

		// An ignored event must not clobber the seeded agent.
		p.Process(IgnoredEvent{})
		if s.Agent != config.AgentCodex {
			t.Errorf("after ignored event Agent = %q, want %q", s.Agent, config.AgentCodex)
		}

		// A definitive codex event keeps it codex.
		p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeTurnStarted}})
		if s.Agent != config.AgentCodex {
			t.Errorf("after codex event Agent = %q, want %q", s.Agent, config.AgentCodex)
		}
	})

	t.Run("assistant events brand the session as claude", func(t *testing.T) {
		s := state.NewState()
		p := NewEventProcessor(s)

		p.Process(AssistantEvent{Data: assistant.Event{}})

		if s.Agent != config.AgentClaude {
			t.Errorf("Agent = %q, want %q", s.Agent, config.AgentClaude)
		}
	})
}

func TestProcessCodex_CommandLifecycle(t *testing.T) {
	p := NewEventProcessor(state.NewState())
	exit := 0
	item := codex.Item{
		ID:               "c1",
		Type:             codex.ItemCommandExecution,
		Command:          "/usr/bin/zsh -lc ls",
		AggregatedOutput: "foo.txt\n",
		ExitCode:         &exit,
		Status:           "completed",
	}

	started := p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if !strings.Contains(started.Rendered, "ls") {
		t.Errorf("started rendered %q, want command", started.Rendered)
	}

	completed := p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &item}})
	if !strings.Contains(completed.Rendered, "foo.txt") {
		t.Errorf("completed rendered %q, want output", completed.Rendered)
	}
	if strings.Contains(completed.Rendered, "Shell") {
		t.Errorf("completed should not repeat the header, got %q", completed.Rendered)
	}
}
