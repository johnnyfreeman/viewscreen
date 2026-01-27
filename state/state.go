package state

import (
	"encoding/json"

	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/user"
)

// Todo represents a tracked task item
type Todo struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"` // "pending", "in_progress", "completed"
	ActiveForm  string `json:"activeForm"`
}

// State holds the centralized session state extracted from events
type State struct {
	// Session info from system event
	Model             string
	Version           string
	CWD               string
	ToolsCount        int
	Agents            []string
	PermissionMode    string

	// Original prompt (if available)
	Prompt            string

	// Runtime tracking
	TurnCount         int
	TotalCost         float64

	// Todos from TodoWrite results
	Todos             []Todo

	// Current tool being executed (for spinner display)
	CurrentTool       string
	CurrentToolInput  string
	ToolInProgress    bool

	// Usage tracking
	InputTokens       int
	OutputTokens      int
	CacheCreated      int
	CacheRead         int

	// Session status
	IsError           bool
	DurationMS        int
	DurationAPIMS     int
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		Todos: make([]Todo, 0),
	}
}

// UpdateFromSystemEvent extracts state from a system event
func (s *State) UpdateFromSystemEvent(event system.Event) {
	s.Model = event.Model
	s.Version = event.ClaudeCodeVersion
	s.CWD = event.CWD
	s.ToolsCount = len(event.Tools)
	s.Agents = event.Agents
	s.PermissionMode = event.PermissionMode
}

// IncrementTurnCount increments the turn counter
func (s *State) IncrementTurnCount() {
	s.TurnCount++
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
	// Clear tool in progress since we got a result
	s.ClearCurrentTool()

	// Try to parse as todo result
	var todoResult user.TodoResult
	if err := json.Unmarshal(toolUseResult, &todoResult); err == nil && len(todoResult.NewTodos) > 0 {
		s.Todos = make([]Todo, len(todoResult.NewTodos))
		for i, t := range todoResult.NewTodos {
			s.Todos[i] = Todo{
				Subject:    t.Content,
				Status:     t.Status,
				ActiveForm: t.ActiveForm,
			}
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
