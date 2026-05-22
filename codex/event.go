// Package codex parses and renders the JSONL event stream emitted by the
// Codex CLI ("codex exec --json").
//
// Codex's stream is structured very differently from Claude Code's
// stream-json: instead of message events carrying content blocks, Codex emits
// a flat sequence of envelope events (thread.*, turn.*, item.*) where each
// "item" is a self-contained unit of work (an agent message, a shell command,
// a file change, etc.). This package models that schema and renders it using
// the shared style/render helpers so Codex output looks at home alongside the
// Claude renderers.
package codex

import (
	"encoding/json"
	"strings"
)

// Envelope event types emitted by codex exec --json.
const (
	TypeThreadStarted = "thread.started"
	TypeTurnStarted   = "turn.started"
	TypeTurnCompleted = "turn.completed"
	TypeTurnFailed    = "turn.failed"
	TypeItemStarted   = "item.started"
	TypeItemUpdated   = "item.updated"
	TypeItemCompleted = "item.completed"
	TypeError         = "error"
)

// Item types carried inside item.* envelope events.
const (
	ItemAgentMessage     = "agent_message"
	ItemReasoning        = "reasoning"
	ItemCommandExecution = "command_execution"
	ItemFileChange       = "file_change"
	ItemTodoList         = "todo_list"
	ItemMCPToolCall      = "mcp_tool_call"
	ItemWebSearch        = "web_search"
	ItemError            = "error"
)

// Event is a single parsed line from the codex stream. Only the fields
// relevant to the envelope's Type are populated.
type Event struct {
	Type     string       `json:"type"`
	ThreadID string       `json:"thread_id,omitempty"`
	Usage    *Usage       `json:"usage,omitempty"`
	Error    *ThreadError `json:"error,omitempty"`
	Item     *Item        `json:"item,omitempty"`
	// Message carries the text of a top-level "error" envelope.
	Message string `json:"message,omitempty"`
}

// Usage holds token accounting reported on turn.completed.
type Usage struct {
	InputTokens           int `json:"input_tokens"`
	CachedInputTokens     int `json:"cached_input_tokens"`
	OutputTokens          int `json:"output_tokens"`
	ReasoningOutputTokens int `json:"reasoning_output_tokens"`
}

// ThreadError describes a failed turn.
type ThreadError struct {
	Message string `json:"message"`
}

// Item is the union of every item shape codex emits. Fields are decoded
// leniently so unknown or partially-populated items still render sensibly.
type Item struct {
	Raw json.RawMessage `json:"-"`

	ID   string `json:"id"`
	Type string `json:"type"`

	// agent_message, reasoning, and error items.
	Text    string `json:"text,omitempty"`
	Message string `json:"message,omitempty"`

	// command_execution items.
	Command          string `json:"command,omitempty"`
	AggregatedOutput string `json:"aggregated_output,omitempty"`
	ExitCode         *int   `json:"exit_code,omitempty"`
	Status           string `json:"status,omitempty"`

	// file_change items.
	Changes []FileChange `json:"changes,omitempty"`

	// todo_list items.
	Items []TodoItem `json:"items,omitempty"`

	// mcp_tool_call items.
	Server string `json:"server,omitempty"`
	Tool   string `json:"tool,omitempty"`

	// web_search items.
	Query string `json:"query,omitempty"`
}

// FileChange is a single path touched by a file_change item.
type FileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"` // "add", "update", or "delete"
}

// TodoItem is a single entry in a todo_list item.
type TodoItem struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// IsEventType reports whether t is a codex envelope event type. It is used to
// distinguish codex lines from Claude Code lines, which never use these known
// types or Codex's dotted envelope naming convention.
func IsEventType(t string) bool {
	switch t {
	case TypeThreadStarted, TypeTurnStarted, TypeTurnCompleted, TypeTurnFailed,
		TypeItemStarted, TypeItemUpdated, TypeItemCompleted, TypeError:
		return true
	default:
		return strings.Contains(t, ".")
	}
}

// UnmarshalJSON decodes a codex item while preserving the raw payload. Codex
// can add item types before viewscreen knows their typed fields, and keeping
// the original JSON lets the renderer still show useful detail for those
// forward-compatible fallbacks.
func (i *Item) UnmarshalJSON(data []byte) error {
	type item Item
	var decoded item
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*i = Item(decoded)
	i.Raw = append(i.Raw[:0], data...)
	return nil
}

// ParseEvent decodes a single codex JSONL line into an Event.
func ParseEvent(line []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(line, &event); err != nil {
		return Event{}, err
	}
	return event, nil
}
