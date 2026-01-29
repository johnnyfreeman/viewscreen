package tools

import (
	"github.com/johnnyfreeman/viewscreen/types"
)

// AssistantMessage represents the minimal interface needed for buffering tool_use blocks.
// This allows the tracker to accept assistant events without importing the assistant package.
type AssistantMessage struct {
	Content         []types.ContentBlock
	ParentToolUseID *string
}

// UserMessage represents the minimal interface needed for matching tool results.
// This allows the tracker to accept user events without importing the user package.
type UserMessage struct {
	Content []UserToolResult
}

// UserToolResult represents a tool result content block.
type UserToolResult struct {
	Type      string
	ToolUseID string
}

// PendingTool holds a tool_use block waiting for its result.
type PendingTool struct {
	Block           types.ContentBlock
	ParentToolUseID *string
}

// ToolUseTracker tracks pending tool_use blocks waiting for their results.
// It buffers tool invocations until their results arrive, enabling proper
// pairing of tool headers with tool outputs.
type ToolUseTracker struct {
	pending        map[string]PendingTool
	headerRendered map[string]bool // tracks tools whose headers were rendered early
}

// NewToolUseTracker creates a new tracker.
func NewToolUseTracker() *ToolUseTracker {
	return &ToolUseTracker{
		pending:        make(map[string]PendingTool),
		headerRendered: make(map[string]bool),
	}
}

// Add registers a pending tool_use block.
func (t *ToolUseTracker) Add(id string, block types.ContentBlock, parentToolUseID *string) {
	t.pending[id] = PendingTool{
		Block:           block,
		ParentToolUseID: parentToolUseID,
	}
}

// Get retrieves a pending tool by ID, returning the tool and whether it exists.
func (t *ToolUseTracker) Get(id string) (PendingTool, bool) {
	p, ok := t.pending[id]
	return p, ok
}

// Remove deletes a pending tool by ID.
func (t *ToolUseTracker) Remove(id string) {
	delete(t.pending, id)
	delete(t.headerRendered, id)
}

// MarkHeaderRendered marks a tool's header as already rendered.
// This is used when a sub-agent prompt triggers early header rendering.
func (t *ToolUseTracker) MarkHeaderRendered(id string) {
	t.headerRendered[id] = true
}

// IsHeaderRendered checks if a tool's header was already rendered.
func (t *ToolUseTracker) IsHeaderRendered(id string) bool {
	return t.headerRendered[id]
}

// IsParentPending checks if a parent tool_use is still pending (waiting for result).
func (t *ToolUseTracker) IsParentPending(parentID string) bool {
	_, ok := t.pending[parentID]
	return ok
}

// IsNested checks if the given pending tool is nested (its parent is also pending).
func (t *ToolUseTracker) IsNested(pending PendingTool) bool {
	return pending.ParentToolUseID != nil && t.IsParentPending(*pending.ParentToolUseID)
}

// Len returns the number of pending tools.
func (t *ToolUseTracker) Len() int {
	return len(t.pending)
}

// ForEach iterates over all pending tools. The iteration order is not guaranteed.
func (t *ToolUseTracker) ForEach(fn func(id string, pending PendingTool)) {
	for id, p := range t.pending {
		fn(id, p)
	}
}

// Clear removes all pending tools.
func (t *ToolUseTracker) Clear() {
	t.pending = make(map[string]PendingTool)
	t.headerRendered = make(map[string]bool)
}

// ResolvedTool contains the data for a tool that has been matched or orphaned.
// It is a pure data structure - rendering is handled by HeaderRenderer.
type ResolvedTool struct {
	Block    types.ContentBlock
	IsNested bool
}

// MatchedTool represents a tool_use block matched with its result.
type MatchedTool struct {
	ResolvedTool
	HeaderRendered bool // true if header was already rendered (e.g., by sub-agent prompt)
}

// OrphanedTool represents a pending tool that has no matching result.
type OrphanedTool struct {
	ResolvedTool
	ID string
}

// MatchAndRemove finds pending tools by their IDs, removes them from the tracker,
// and returns information about each matched tool.
// This is the core matching logic used when processing tool results.
func (t *ToolUseTracker) MatchAndRemove(toolUseIDs []string) []MatchedTool {
	var matched []MatchedTool

	for _, id := range toolUseIDs {
		if pending, ok := t.Get(id); ok {
			isNested := t.IsNested(pending)
			headerRendered := t.IsHeaderRendered(id)
			matched = append(matched, MatchedTool{
				ResolvedTool: ResolvedTool{
					Block:    pending.Block,
					IsNested: isNested,
				},
				HeaderRendered: headerRendered,
			})
			t.Remove(id)
		}
	}

	return matched
}

// FlushAll removes all pending tools and returns them as orphaned.
// Call this when processing a result event to handle any tools that didn't get results.
func (t *ToolUseTracker) FlushAll() []OrphanedTool {
	var orphaned []OrphanedTool
	t.ForEach(func(id string, pending PendingTool) {
		orphaned = append(orphaned, OrphanedTool{
			ResolvedTool: ResolvedTool{
				Block:    pending.Block,
				IsNested: t.IsNested(pending),
			},
			ID: id,
		})
	})
	t.Clear()
	return orphaned
}

// BufferFromAssistantMessage buffers tool_use blocks from an assistant message.
// The inToolUseBlock parameter indicates if we're currently streaming a tool_use block,
// in which case we skip buffering (the tool will be rendered by the stream handler).
// Returns true if any tools were buffered.
func (t *ToolUseTracker) BufferFromAssistantMessage(msg AssistantMessage, inToolUseBlock bool) bool {
	buffered := false
	for _, block := range msg.Content {
		if block.Type == "tool_use" && block.ID != "" {
			if !inToolUseBlock {
				t.Add(block.ID, block, msg.ParentToolUseID)
				buffered = true
			}
		}
	}
	return buffered
}

// MatchFromUserMessage matches tool_result content blocks with pending tool_use blocks.
// It extracts tool_use IDs from the message and removes matching tools from the tracker.
func (t *ToolUseTracker) MatchFromUserMessage(msg UserMessage) []MatchedTool {
	var toolUseIDs []string
	for _, content := range msg.Content {
		if content.Type == "tool_result" && content.ToolUseID != "" {
			toolUseIDs = append(toolUseIDs, content.ToolUseID)
		}
	}
	return t.MatchAndRemove(toolUseIDs)
}

// ResolveParentEarly finds a pending tool by ID, marks its header as rendered,
// and returns the resolved tool data for early header rendering.
// This is used when a sub-agent prompt triggers early rendering of the parent tool.
// Returns nil if the tool is not found or its header was already rendered.
func (t *ToolUseTracker) ResolveParentEarly(parentID string) *ResolvedTool {
	pending, ok := t.Get(parentID)
	if !ok || t.IsHeaderRendered(parentID) {
		return nil
	}
	t.MarkHeaderRendered(parentID)
	return &ResolvedTool{
		Block:    pending.Block,
		IsNested: t.IsNested(pending),
	}
}
