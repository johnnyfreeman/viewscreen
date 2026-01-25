package types

import "encoding/json"

// BaseEvent represents the common fields in all events
type BaseEvent struct {
	Type            string  `json:"type"`
	SessionID       string  `json:"session_id"`
	UUID            string  `json:"uuid"`
	ParentToolUseID *string `json:"parent_tool_use_id"`
}

// ContentBlock represents a content block (text or tool_use)
type ContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens              int    `json:"input_tokens"`
	CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
	OutputTokens             int    `json:"output_tokens"`
	ServiceTier              string `json:"service_tier"`
}
