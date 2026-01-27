package tools

import "github.com/johnnyfreeman/viewscreen/types"

// PendingTool holds a tool_use block waiting for its result.
type PendingTool struct {
	Block           types.ContentBlock
	ParentToolUseID *string
}

// ToolUseTracker tracks pending tool_use blocks waiting for their results.
// It buffers tool invocations until their results arrive, enabling proper
// pairing of tool headers with tool outputs.
type ToolUseTracker struct {
	pending map[string]PendingTool
}

// NewToolUseTracker creates a new tracker.
func NewToolUseTracker() *ToolUseTracker {
	return &ToolUseTracker{
		pending: make(map[string]PendingTool),
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
}

// MatchedTool represents a tool_use block matched with its result.
type MatchedTool struct {
	Block    types.ContentBlock
	IsNested bool
}

// MatchAndRemove finds pending tools by their IDs, removes them from the tracker,
// and returns information about each matched tool.
// This is the core matching logic used when processing tool results.
func (t *ToolUseTracker) MatchAndRemove(toolUseIDs []string) []MatchedTool {
	var matched []MatchedTool

	for _, id := range toolUseIDs {
		if pending, ok := t.Get(id); ok {
			isNested := t.IsNested(pending)
			matched = append(matched, MatchedTool{
				Block:    pending.Block,
				IsNested: isNested,
			})
			t.Remove(id)
		}
	}

	return matched
}

// OrphanedTool represents a pending tool that has no matching result.
type OrphanedTool struct {
	ID       string
	Block    types.ContentBlock
	IsNested bool
}

// FlushAll removes all pending tools and returns them as orphaned.
// Call this when processing a result event to handle any tools that didn't get results.
func (t *ToolUseTracker) FlushAll() []OrphanedTool {
	var orphaned []OrphanedTool
	t.ForEach(func(id string, pending PendingTool) {
		orphaned = append(orphaned, OrphanedTool{
			ID:       id,
			Block:    pending.Block,
			IsNested: t.IsNested(pending),
		})
	})
	t.Clear()
	return orphaned
}
