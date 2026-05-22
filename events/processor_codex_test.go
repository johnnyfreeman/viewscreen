package events

import (
	"os"
	"path/filepath"
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
	if !strings.Contains(completed.Rendered, "1 lines") {
		t.Errorf("completed rendered %q, want output summary", completed.Rendered)
	}
	if strings.Contains(completed.Rendered, "foo.txt") {
		t.Errorf("completed should not expand output at default verbosity, got %q", completed.Rendered)
	}
	if strings.Contains(completed.Rendered, "Shell") {
		t.Errorf("completed should not repeat the header, got %q", completed.Rendered)
	}
}

func TestProcessCodex_CommandDrivesSpinner(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)
	item := codex.Item{ID: "c1", Type: codex.ItemCommandExecution, Command: "/usr/bin/zsh -lc 'go test ./...'", Status: "in_progress"}

	// A started command becomes the running tool, with the unwrapped command
	// as the spinner input.
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if !s.ToolInProgress {
		t.Fatal("ToolInProgress = false after item.started, want true")
	}
	if s.CurrentTool != "Shell" {
		t.Errorf("CurrentTool = %q, want Shell", s.CurrentTool)
	}
	if s.CurrentToolInput != "go test ./..." {
		t.Errorf("CurrentToolInput = %q, want unwrapped command", s.CurrentToolInput)
	}

	// Completing the command clears the spinner.
	item.Status = "completed"
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &item}})
	if s.ToolInProgress {
		t.Error("ToolInProgress = true after item.completed, want false")
	}
}

func TestProcessCodex_OverlappingCommandsRestoreSpinner(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	slow := codex.Item{ID: "c1", Type: codex.ItemCommandExecution, Command: "/usr/bin/zsh -lc 'rg --files'", Status: "in_progress"}
	fast := codex.Item{ID: "c2", Type: codex.ItemCommandExecution, Command: "/usr/bin/zsh -lc 'git status --short'", Status: "in_progress"}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &slow}})
	if s.CurrentToolInput != "rg --files" {
		t.Fatalf("CurrentToolInput after first start = %q, want rg --files", s.CurrentToolInput)
	}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &fast}})
	if s.CurrentToolInput != "git status --short" {
		t.Fatalf("CurrentToolInput after second start = %q, want git status --short", s.CurrentToolInput)
	}

	fast.Status = "completed"
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &fast}})
	if !s.ToolInProgress {
		t.Fatal("ToolInProgress = false after completing the newest command, want true for the older active command")
	}
	if s.CurrentToolInput != "rg --files" {
		t.Errorf("CurrentToolInput after completing newest command = %q, want rg --files", s.CurrentToolInput)
	}

	slow.Status = "completed"
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &slow}})
	if s.ToolInProgress {
		t.Error("ToolInProgress = true after all commands completed, want false")
	}
}

func TestProcessCodex_FileChangeSnapshotRendersPatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.go")
	if err := os.WriteFile(path, []byte("package main\n\nfunc old() {}\n"), 0o600); err != nil {
		t.Fatalf("write old file: %v", err)
	}

	s := state.NewState()
	p := NewEventProcessor(s)
	p.Process(CodexEvent{Data: codex.Event{
		Type: codex.TypeItemStarted,
		Item: &codex.Item{ID: "f1", Type: codex.ItemFileChange, Changes: []codex.FileChange{{Path: path, Kind: "update"}}, Status: "in_progress"},
	}})

	if err := os.WriteFile(path, []byte("package main\n\nfunc new() {}\n"), 0o600); err != nil {
		t.Fatalf("write new file: %v", err)
	}

	res := p.Process(CodexEvent{Data: codex.Event{
		Type: codex.TypeItemCompleted,
		Item: &codex.Item{ID: "f1", Type: codex.ItemFileChange, Changes: []codex.FileChange{{Path: path, Kind: "update"}}, Status: "completed"},
	}})

	if strings.Contains(res.Rendered, "Edit") {
		t.Fatalf("completion should not repeat header, got %q", res.Rendered)
	}
	for _, want := range []string{"old", "new", "│", "-"} {
		if !strings.Contains(res.Rendered, want) {
			t.Fatalf("rendered completion = %q, want %q", res.Rendered, want)
		}
	}
}

func TestProcessCodex_MCPCallDrivesSpinner(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)
	item := codex.Item{ID: "m1", Type: codex.ItemMCPToolCall, Server: "github", Tool: "create_issue", Status: "in_progress"}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if s.CurrentTool != "github.create_issue" {
		t.Errorf("CurrentTool = %q, want github.create_issue", s.CurrentTool)
	}

	item.Status = "completed"
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &item}})
	if s.ToolInProgress {
		t.Error("ToolInProgress = true after MCP item.completed, want false")
	}
}

func TestProcessCodex_FileChangeDrivesSpinner(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)
	item := codex.Item{
		ID:      "f1",
		Type:    codex.ItemFileChange,
		Changes: []codex.FileChange{{Path: "/tmp/bar.txt", Kind: "add"}},
		Status:  "in_progress",
	}

	// An in-flight file change shows the Edit spinner labeled with the path,
	// matching the inline header.
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if s.CurrentTool != "Edit" {
		t.Errorf("CurrentTool = %q, want Edit", s.CurrentTool)
	}
	if s.CurrentToolInput != "/tmp/bar.txt" {
		t.Errorf("CurrentToolInput = %q, want the file path", s.CurrentToolInput)
	}

	item.Status = "completed"
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &item}})
	if s.ToolInProgress {
		t.Error("ToolInProgress = true after file_change item.completed, want false")
	}
}

func TestProcessCodex_WebSearchDrivesSpinner(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)
	item := codex.Item{ID: "w1", Type: codex.ItemWebSearch, Query: "golang testing", Status: "in_progress"}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if s.CurrentTool != "Web Search" {
		t.Errorf("CurrentTool = %q, want Web Search", s.CurrentTool)
	}
	if s.CurrentToolInput != "golang testing" {
		t.Errorf("CurrentToolInput = %q, want the query", s.CurrentToolInput)
	}

	item.Status = "completed"
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &item}})
	if s.ToolInProgress {
		t.Error("ToolInProgress = true after web_search item.completed, want false")
	}
}

func TestProcessCodex_TurnStartedIncrementsTurnCount(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)

	if s.TurnCount != 0 {
		t.Fatalf("TurnCount = %d before any event, want 0", s.TurnCount)
	}

	// A codex exec is one turn: turn.started bumps the count so the sidebar
	// stops showing a perpetual "Turns: 0".
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeTurnStarted}})
	if s.TurnCount != 1 {
		t.Errorf("TurnCount = %d after turn.started, want 1", s.TurnCount)
	}

	// turn.completed records usage but must not double-count the turn.
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeTurnCompleted, Usage: &codex.Usage{InputTokens: 10}}})
	if s.TurnCount != 1 {
		t.Errorf("TurnCount = %d after turn.completed, want 1", s.TurnCount)
	}
}

func TestProcessCodex_TurnEndClearsSpinner(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)
	item := codex.Item{ID: "c1", Type: codex.ItemCommandExecution, Command: "sleep 100", Status: "in_progress"}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if !s.ToolInProgress {
		t.Fatal("expected a running tool before the turn ends")
	}

	// A turn that ends without a matching item.completed must not leave a
	// stale spinner running.
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeTurnCompleted}})
	if s.ToolInProgress {
		t.Error("ToolInProgress = true after turn.completed, want false")
	}
}

func TestProcessCodex_TodoListPopulatesSidebar(t *testing.T) {
	s := state.NewState()
	p := NewEventProcessor(s)
	item := codex.Item{
		ID:   "t1",
		Type: codex.ItemTodoList,
		Items: []codex.TodoItem{
			{Text: "Write the test", Completed: true},
			{Text: "Make it pass", Completed: false},
		},
	}

	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemStarted, Item: &item}})
	if len(s.Todos) != 2 {
		t.Fatalf("len(Todos) = %d, want 2", len(s.Todos))
	}
	if s.Todos[0].Status != "completed" || s.Todos[1].Status != "pending" {
		t.Errorf("Todos statuses = %q/%q, want completed/pending", s.Todos[0].Status, s.Todos[1].Status)
	}
	completed, total := s.TodoProgress()
	if completed != 1 || total != 2 {
		t.Errorf("TodoProgress = %d/%d, want 1/2", completed, total)
	}

	// A later update with more items checked off replaces the list so the
	// sidebar tracks the latest completion state.
	item.Items[1].Completed = true
	p.Process(CodexEvent{Data: codex.Event{Type: codex.TypeItemCompleted, Item: &item}})
	if completed, total := s.TodoProgress(); completed != 2 || total != 2 {
		t.Errorf("after update TodoProgress = %d/%d, want 2/2", completed, total)
	}
}
