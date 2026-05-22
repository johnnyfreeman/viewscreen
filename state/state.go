package state

import (
	"encoding/json"
	"time"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/system"
)

// Todo represents a tracked task item from TodoWrite tool results.
type Todo struct {
	Content    string `json:"content"`
	Status     string `json:"status"` // "pending", "in_progress", "completed"
	ActiveForm string `json:"activeForm"`
}

// TodoResult represents the tool_use_result for TodoWrite operations.
type TodoResult struct {
	OldTodos []Todo `json:"oldTodos"`
	NewTodos []Todo `json:"newTodos"`
}

// TaskListItem represents one task in a TaskList tool result.
// The Task* tools replaced TodoWrite in Claude Code 2.1.142.
type TaskListItem struct {
	ID        string   `json:"id"`
	Subject   string   `json:"subject"`
	Status    string   `json:"status"`
	BlockedBy []string `json:"blockedBy"`
}

// State holds the centralized session state extracted from events
type State struct {
	// Agent identifies the CLI that produced the stream ("claude" or "codex").
	// It drives agent-specific branding in the TUI. It is empty until detected
	// from the stream (see events.EventProcessor) or seeded from the spawn
	// configuration in prompt mode.
	Agent string

	// Session info from system event
	Model          string
	Version        string
	CWD            string
	ToolsCount     int
	Agents         []string
	PermissionMode string

	// Original prompt (if available)
	Prompt string

	// Runtime tracking
	TurnCount int
	TotalCost float64

	// Todos from TodoWrite results
	Todos []Todo

	// Current tool being executed (for spinner display)
	CurrentTool      string
	CurrentToolInput string
	ToolInProgress   bool

	// Usage tracking
	InputTokens  int
	OutputTokens int
	CacheCreated int
	CacheRead    int

	// ReasoningTokens counts model "thinking" tokens. Codex reports these on
	// turn.completed (reasoning_output_tokens); Claude's stream does not break
	// them out separately, so this stays zero for Claude streams.
	ReasoningTokens int

	// Session timing
	StartTime time.Time

	// Session status
	IsError       bool
	DurationMS    int
	DurationAPIMS int
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Todos:     make([]Todo, 0),
		StartTime: time.Now(),
	}
}

// Elapsed returns the duration since the session started.
func (s *State) Elapsed() time.Duration {
	return time.Since(s.StartTime)
}

// ReportsCost reports whether the active agent emits a dollar cost for the
// session. Codex's stream carries only token usage — never a cost — so its
// sidebar omits the Cost and Rate fields rather than showing a misleading
// $0.0000. Any other agent (Claude, or an as-yet-undetected stream) is assumed
// to report cost.
func (s *State) ReportsCost() bool {
	return s.Agent != config.AgentCodex
}

// CostRate returns the cost per minute based on total cost and elapsed time.
// Returns 0 if elapsed time is less than 1 second (to avoid division by zero
// and noisy early values).
func (s *State) CostRate() float64 {
	elapsed := s.Elapsed()
	if elapsed < time.Second {
		return 0
	}
	return s.TotalCost / elapsed.Minutes()
}

// UpdateFromSystemEvent extracts state from a system event.
// Only overwrites fields that have non-empty values, so subagent system events
// (which carry empty fields) cannot clobber the parent session state.
func (s *State) UpdateFromSystemEvent(event system.Event) {
	if event.Model != "" {
		s.Model = event.Model
	}
	if event.ClaudeCodeVersion != "" {
		s.Version = event.ClaudeCodeVersion
	}
	if event.CWD != "" {
		s.CWD = event.CWD
	}
	if len(event.Tools) > 0 {
		s.ToolsCount = len(event.Tools)
	}
	if len(event.Agents) > 0 {
		s.Agents = event.Agents
	}
	if event.PermissionMode != "" {
		s.PermissionMode = event.PermissionMode
	}
}

// IncrementTurnCount increments the turn counter
func (s *State) IncrementTurnCount() {
	s.TurnCount++
}

// AccumulateUsage adds per-turn token usage to the running totals.
// This is called for each assistant message to provide real-time tracking.
func (s *State) AccumulateUsage(input, output, cacheCreated, cacheRead int) {
	s.InputTokens += input
	s.OutputTokens += output
	s.CacheCreated += cacheCreated
	s.CacheRead += cacheRead
}

// SetCurrentTool sets the current tool being executed
func (s *State) SetCurrentTool(name, input string) {
	s.CurrentTool = name
	s.CurrentToolInput = input
	s.ToolInProgress = true
}

// ClearCurrentTool clears the current tool state
func (s *State) ClearCurrentTool() {
	s.CurrentTool = ""
	s.CurrentToolInput = ""
	s.ToolInProgress = false
}

// UpdateFromToolUseResult extracts state from a tool_use_result
func (s *State) UpdateFromToolUseResult(toolUseResult json.RawMessage) {
	if len(toolUseResult) == 0 {
		return
	}

	// Clear tool in progress since we got a result
	s.ClearCurrentTool()

	// TodoWrite (newTodos) and TaskList (tasks) results each own the full task
	// list. An empty array intentionally clears the sidebar list.
	var raw struct {
		NewTodos json.RawMessage `json:"newTodos"`
		Tasks    json.RawMessage `json:"tasks"`
	}
	if err := json.Unmarshal(toolUseResult, &raw); err != nil {
		return
	}

	if raw.NewTodos != nil {
		var todos []Todo
		if err := json.Unmarshal(raw.NewTodos, &todos); err == nil {
			s.Todos = todos
		}
		return
	}

	if raw.Tasks != nil {
		var tasks []TaskListItem
		if err := json.Unmarshal(raw.Tasks, &tasks); err == nil {
			todos := make([]Todo, len(tasks))
			for i, t := range tasks {
				todos[i] = Todo{Content: t.Subject, Status: t.Status}
			}
			s.Todos = todos
		}
	}
}

// UpdateFromResultEvent extracts final state from a result event
func (s *State) UpdateFromResultEvent(event result.Event) {
	s.TurnCount = event.NumTurns
	s.TotalCost = event.TotalCostUSD
	s.IsError = event.IsError
	s.DurationMS = event.DurationMS
	s.DurationAPIMS = event.DurationAPIMS
	s.InputTokens = event.Usage.InputTokens
	s.OutputTokens = event.Usage.OutputTokens
	s.CacheCreated = event.Usage.CacheCreationInputTokens
	s.CacheRead = event.Usage.CacheReadInputTokens
}

// HasActiveTodos returns true if there are any in-progress todos
func (s *State) HasActiveTodos() bool {
	for _, todo := range s.Todos {
		if todo.Status == "in_progress" {
			return true
		}
	}
	return false
}

// GetActiveTodo returns the first in-progress todo, if any
func (s *State) GetActiveTodo() *Todo {
	for i := range s.Todos {
		if s.Todos[i].Status == "in_progress" {
			return &s.Todos[i]
		}
	}
	return nil
}

// TodoProgress returns the number of completed todos and total todos.
func (s *State) TodoProgress() (completed, total int) {
	total = len(s.Todos)
	for _, todo := range s.Todos {
		if todo.Status == "completed" {
			completed++
		}
	}
	return completed, total
}
