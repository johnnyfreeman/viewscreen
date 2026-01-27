package tools

import (
	"testing"

	"github.com/johnnyfreeman/viewscreen/types"
)

func TestNewToolUseTracker(t *testing.T) {
	tracker := NewToolUseTracker()
	if tracker == nil {
		t.Fatal("NewToolUseTracker() returned nil")
	}
	if tracker.Len() != 0 {
		t.Errorf("NewToolUseTracker() should start empty, got Len() = %d", tracker.Len())
	}
}

func TestToolUseTracker_AddAndGet(t *testing.T) {
	tracker := NewToolUseTracker()
	block := types.ContentBlock{
		Type: "tool_use",
		ID:   "tool-123",
		Name: "Bash",
	}

	tracker.Add("tool-123", block, nil)

	pending, ok := tracker.Get("tool-123")
	if !ok {
		t.Fatal("Get() should find added tool")
	}
	if pending.Block.ID != "tool-123" {
		t.Errorf("Get() returned wrong block ID: got %q, want %q", pending.Block.ID, "tool-123")
	}
	if pending.Block.Name != "Bash" {
		t.Errorf("Get() returned wrong block Name: got %q, want %q", pending.Block.Name, "Bash")
	}
	if pending.ParentToolUseID != nil {
		t.Errorf("Get() returned non-nil ParentToolUseID when nil was added")
	}
}

func TestToolUseTracker_AddWithParent(t *testing.T) {
	tracker := NewToolUseTracker()
	parentID := "parent-456"
	block := types.ContentBlock{
		Type: "tool_use",
		ID:   "child-789",
		Name: "Read",
	}

	tracker.Add("child-789", block, &parentID)

	pending, ok := tracker.Get("child-789")
	if !ok {
		t.Fatal("Get() should find added tool")
	}
	if pending.ParentToolUseID == nil {
		t.Fatal("Get() returned nil ParentToolUseID when one was added")
	}
	if *pending.ParentToolUseID != "parent-456" {
		t.Errorf("Get() returned wrong ParentToolUseID: got %q, want %q", *pending.ParentToolUseID, "parent-456")
	}
}

func TestToolUseTracker_GetNotFound(t *testing.T) {
	tracker := NewToolUseTracker()

	_, ok := tracker.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for nonexistent ID")
	}
}

func TestToolUseTracker_Remove(t *testing.T) {
	tracker := NewToolUseTracker()
	block := types.ContentBlock{
		Type: "tool_use",
		ID:   "tool-123",
		Name: "Bash",
	}

	tracker.Add("tool-123", block, nil)
	tracker.Remove("tool-123")

	_, ok := tracker.Get("tool-123")
	if ok {
		t.Error("Get() should return false after Remove()")
	}
	if tracker.Len() != 0 {
		t.Errorf("Len() should be 0 after removing the only item, got %d", tracker.Len())
	}
}

func TestToolUseTracker_RemoveNonexistent(t *testing.T) {
	tracker := NewToolUseTracker()

	// Should not panic
	tracker.Remove("nonexistent")
}

func TestToolUseTracker_IsParentPending(t *testing.T) {
	tracker := NewToolUseTracker()

	// Add a parent tool
	parentBlock := types.ContentBlock{
		Type: "tool_use",
		ID:   "parent-123",
		Name: "Task",
	}
	tracker.Add("parent-123", parentBlock, nil)

	if !tracker.IsParentPending("parent-123") {
		t.Error("IsParentPending() should return true for existing tool")
	}
	if tracker.IsParentPending("nonexistent") {
		t.Error("IsParentPending() should return false for nonexistent tool")
	}
}

func TestToolUseTracker_IsNested(t *testing.T) {
	tracker := NewToolUseTracker()

	// Add a parent tool
	parentBlock := types.ContentBlock{
		Type: "tool_use",
		ID:   "parent-123",
		Name: "Task",
	}
	tracker.Add("parent-123", parentBlock, nil)

	// Add a child tool
	parentID := "parent-123"
	childBlock := types.ContentBlock{
		Type: "tool_use",
		ID:   "child-456",
		Name: "Read",
	}
	tracker.Add("child-456", childBlock, &parentID)

	// Get child and check if nested
	childPending, _ := tracker.Get("child-456")
	if !tracker.IsNested(childPending) {
		t.Error("IsNested() should return true when parent is pending")
	}

	// Remove parent
	tracker.Remove("parent-123")

	// Child should no longer be considered nested
	if tracker.IsNested(childPending) {
		t.Error("IsNested() should return false when parent is no longer pending")
	}
}

func TestToolUseTracker_IsNested_NoParent(t *testing.T) {
	tracker := NewToolUseTracker()

	block := types.ContentBlock{
		Type: "tool_use",
		ID:   "tool-123",
		Name: "Bash",
	}
	tracker.Add("tool-123", block, nil)

	pending, _ := tracker.Get("tool-123")
	if tracker.IsNested(pending) {
		t.Error("IsNested() should return false when ParentToolUseID is nil")
	}
}

func TestToolUseTracker_Len(t *testing.T) {
	tracker := NewToolUseTracker()

	if tracker.Len() != 0 {
		t.Errorf("Len() should be 0 initially, got %d", tracker.Len())
	}

	tracker.Add("tool-1", types.ContentBlock{ID: "tool-1"}, nil)
	if tracker.Len() != 1 {
		t.Errorf("Len() should be 1 after adding one item, got %d", tracker.Len())
	}

	tracker.Add("tool-2", types.ContentBlock{ID: "tool-2"}, nil)
	if tracker.Len() != 2 {
		t.Errorf("Len() should be 2 after adding two items, got %d", tracker.Len())
	}

	tracker.Remove("tool-1")
	if tracker.Len() != 1 {
		t.Errorf("Len() should be 1 after removing one item, got %d", tracker.Len())
	}
}

func TestToolUseTracker_ForEach(t *testing.T) {
	tracker := NewToolUseTracker()

	tracker.Add("tool-1", types.ContentBlock{ID: "tool-1", Name: "Bash"}, nil)
	tracker.Add("tool-2", types.ContentBlock{ID: "tool-2", Name: "Read"}, nil)

	visited := make(map[string]bool)
	tracker.ForEach(func(id string, pending PendingTool) {
		visited[id] = true
	})

	if len(visited) != 2 {
		t.Errorf("ForEach() should visit 2 items, visited %d", len(visited))
	}
	if !visited["tool-1"] {
		t.Error("ForEach() should visit tool-1")
	}
	if !visited["tool-2"] {
		t.Error("ForEach() should visit tool-2")
	}
}

func TestToolUseTracker_Clear(t *testing.T) {
	tracker := NewToolUseTracker()

	tracker.Add("tool-1", types.ContentBlock{ID: "tool-1"}, nil)
	tracker.Add("tool-2", types.ContentBlock{ID: "tool-2"}, nil)

	tracker.Clear()

	if tracker.Len() != 0 {
		t.Errorf("Len() should be 0 after Clear(), got %d", tracker.Len())
	}

	_, ok := tracker.Get("tool-1")
	if ok {
		t.Error("Get() should return false after Clear()")
	}
}

func TestToolUseTracker_MatchAndRemove_NoMatches(t *testing.T) {
	tracker := NewToolUseTracker()

	matched := tracker.MatchAndRemove([]string{"nonexistent-1", "nonexistent-2"})
	if len(matched) != 0 {
		t.Errorf("MatchAndRemove() should return empty slice for no matches, got %d", len(matched))
	}
}

func TestToolUseTracker_MatchAndRemove_SingleMatch(t *testing.T) {
	tracker := NewToolUseTracker()

	tracker.Add("tool-123", types.ContentBlock{ID: "tool-123", Name: "Bash"}, nil)
	tracker.Add("tool-456", types.ContentBlock{ID: "tool-456", Name: "Read"}, nil)

	matched := tracker.MatchAndRemove([]string{"tool-123"})
	if len(matched) != 1 {
		t.Fatalf("MatchAndRemove() should return 1 match, got %d", len(matched))
	}
	if matched[0].Block.ID != "tool-123" {
		t.Errorf("Matched block ID should be 'tool-123', got %q", matched[0].Block.ID)
	}
	if matched[0].IsNested {
		t.Error("Matched tool should not be nested")
	}

	// tool-123 removed, tool-456 still present
	if tracker.Len() != 1 {
		t.Errorf("Tracker should have 1 remaining tool, got %d", tracker.Len())
	}
	_, ok := tracker.Get("tool-123")
	if ok {
		t.Error("tool-123 should be removed after matching")
	}
}

func TestToolUseTracker_MatchAndRemove_MultipleMatches(t *testing.T) {
	tracker := NewToolUseTracker()

	tracker.Add("tool-1", types.ContentBlock{ID: "tool-1", Name: "Bash"}, nil)
	tracker.Add("tool-2", types.ContentBlock{ID: "tool-2", Name: "Read"}, nil)
	tracker.Add("tool-3", types.ContentBlock{ID: "tool-3", Name: "Write"}, nil)

	matched := tracker.MatchAndRemove([]string{"tool-1", "tool-3"})
	if len(matched) != 2 {
		t.Fatalf("MatchAndRemove() should return 2 matches, got %d", len(matched))
	}

	// Only tool-2 should remain
	if tracker.Len() != 1 {
		t.Errorf("Tracker should have 1 remaining tool, got %d", tracker.Len())
	}
	_, ok := tracker.Get("tool-2")
	if !ok {
		t.Error("tool-2 should still be present")
	}
}

func TestToolUseTracker_MatchAndRemove_NestedTool(t *testing.T) {
	tracker := NewToolUseTracker()

	// Add parent tool
	tracker.Add("parent", types.ContentBlock{ID: "parent", Name: "Task"}, nil)

	// Add nested child tool
	parentID := "parent"
	tracker.Add("child", types.ContentBlock{ID: "child", Name: "Read"}, &parentID)

	// Match only child (parent still pending)
	matched := tracker.MatchAndRemove([]string{"child"})
	if len(matched) != 1 {
		t.Fatalf("MatchAndRemove() should return 1 match, got %d", len(matched))
	}
	if !matched[0].IsNested {
		t.Error("Matched tool should be nested when parent is still pending")
	}

	// Parent still present
	if tracker.Len() != 1 {
		t.Errorf("Tracker should have 1 remaining tool, got %d", tracker.Len())
	}
}

func TestToolUseTracker_FlushAll_Empty(t *testing.T) {
	tracker := NewToolUseTracker()

	orphaned := tracker.FlushAll()
	if len(orphaned) != 0 {
		t.Errorf("FlushAll() should return empty slice for empty tracker, got %d", len(orphaned))
	}
}

func TestToolUseTracker_FlushAll_ReturnsAll(t *testing.T) {
	tracker := NewToolUseTracker()

	tracker.Add("tool-1", types.ContentBlock{ID: "tool-1", Name: "Bash"}, nil)
	tracker.Add("tool-2", types.ContentBlock{ID: "tool-2", Name: "Read"}, nil)

	orphaned := tracker.FlushAll()
	if len(orphaned) != 2 {
		t.Fatalf("FlushAll() should return 2 orphaned tools, got %d", len(orphaned))
	}

	// Tracker should be empty
	if tracker.Len() != 0 {
		t.Errorf("Tracker should be empty after FlushAll(), got %d", tracker.Len())
	}

	// Check all tools returned
	ids := make(map[string]bool)
	for _, o := range orphaned {
		ids[o.ID] = true
	}
	if !ids["tool-1"] || !ids["tool-2"] {
		t.Error("FlushAll() should return all pending tools")
	}
}

func TestToolUseTracker_FlushAll_NestedDetection(t *testing.T) {
	tracker := NewToolUseTracker()

	// Add parent
	tracker.Add("parent", types.ContentBlock{ID: "parent", Name: "Task"}, nil)

	// Add nested child
	parentID := "parent"
	tracker.Add("child", types.ContentBlock{ID: "child", Name: "Read"}, &parentID)

	orphaned := tracker.FlushAll()

	// Find parent and child
	var parentOrphan, childOrphan *OrphanedTool
	for i := range orphaned {
		if orphaned[i].ID == "parent" {
			parentOrphan = &orphaned[i]
		} else if orphaned[i].ID == "child" {
			childOrphan = &orphaned[i]
		}
	}

	if parentOrphan == nil {
		t.Fatal("FlushAll() should include parent")
	}
	if childOrphan == nil {
		t.Fatal("FlushAll() should include child")
	}

	if parentOrphan.IsNested {
		t.Error("Parent tool should not be nested")
	}
	if !childOrphan.IsNested {
		t.Error("Child tool should be nested")
	}
}
