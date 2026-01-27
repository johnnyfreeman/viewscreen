package stream

import (
	"encoding/json"
	"testing"

	"github.com/johnnyfreeman/viewscreen/types"
)

func TestBlockType_String(t *testing.T) {
	tests := []struct {
		blockType BlockType
		expected  string
	}{
		{BlockNone, "none"},
		{BlockText, "text"},
		{BlockToolUse, "tool_use"},
	}

	for _, tt := range tests {
		if got := tt.blockType.String(); got != tt.expected {
			t.Errorf("BlockType(%d).String() = %q, want %q", tt.blockType, got, tt.expected)
		}
	}
}

func TestNewBlockState(t *testing.T) {
	s := NewBlockState()

	if s == nil {
		t.Fatal("NewBlockState returned nil")
	}

	if s.Type() != BlockNone {
		t.Errorf("expected initial type to be BlockNone, got %v", s.Type())
	}

	if s.Index() != -1 {
		t.Errorf("expected initial index to be -1, got %d", s.Index())
	}

	if s.ToolName() != "" {
		t.Errorf("expected initial tool name to be empty, got %q", s.ToolName())
	}

	if s.InTextBlock() {
		t.Error("expected InTextBlock to be false initially")
	}

	if s.InToolUseBlock() {
		t.Error("expected InToolUseBlock to be false initially")
	}

	if s.TextContent() != "" {
		t.Errorf("expected initial text content to be empty, got %q", s.TextContent())
	}

	if s.ToolInput() != "" {
		t.Errorf("expected initial tool input to be empty, got %q", s.ToolInput())
	}
}

func makeTestContentBlock(blockType, name string) json.RawMessage {
	block := types.ContentBlock{
		Type: blockType,
		Name: name,
	}
	data, _ := json.Marshal(block)
	return data
}

func TestBlockState_StartBlock_Text(t *testing.T) {
	s := NewBlockState()

	ok := s.StartBlock(0, makeTestContentBlock("text", ""))

	if !ok {
		t.Error("expected StartBlock to return true for text block")
	}
	if s.Type() != BlockText {
		t.Errorf("expected type to be BlockText, got %v", s.Type())
	}
	if s.Index() != 0 {
		t.Errorf("expected index to be 0, got %d", s.Index())
	}
	if !s.InTextBlock() {
		t.Error("expected InTextBlock to be true")
	}
	if s.InToolUseBlock() {
		t.Error("expected InToolUseBlock to be false")
	}
}

func TestBlockState_StartBlock_ToolUse(t *testing.T) {
	s := NewBlockState()

	ok := s.StartBlock(1, makeTestContentBlock("tool_use", "Read"))

	if !ok {
		t.Error("expected StartBlock to return true for tool_use block")
	}
	if s.Type() != BlockToolUse {
		t.Errorf("expected type to be BlockToolUse, got %v", s.Type())
	}
	if s.Index() != 1 {
		t.Errorf("expected index to be 1, got %d", s.Index())
	}
	if s.ToolName() != "Read" {
		t.Errorf("expected tool name to be 'Read', got %q", s.ToolName())
	}
	if s.InTextBlock() {
		t.Error("expected InTextBlock to be false")
	}
	if !s.InToolUseBlock() {
		t.Error("expected InToolUseBlock to be true")
	}
}

func TestBlockState_StartBlock_EmptyContentBlock(t *testing.T) {
	s := NewBlockState()

	ok := s.StartBlock(0, nil)

	if ok {
		t.Error("expected StartBlock to return false for empty content block")
	}
	if s.Type() != BlockNone {
		t.Errorf("expected type to remain BlockNone, got %v", s.Type())
	}
}

func TestBlockState_StartBlock_InvalidJSON(t *testing.T) {
	s := NewBlockState()

	ok := s.StartBlock(0, json.RawMessage(`invalid json`))

	if ok {
		t.Error("expected StartBlock to return false for invalid JSON")
	}
	// Index is set even on parse failure
	if s.Index() != 0 {
		t.Errorf("expected index to be 0 even on failure, got %d", s.Index())
	}
}

func TestBlockState_StartBlock_UnknownType(t *testing.T) {
	s := NewBlockState()

	ok := s.StartBlock(0, makeTestContentBlock("unknown_type", ""))

	if ok {
		t.Error("expected StartBlock to return false for unknown block type")
	}
	if s.Type() != BlockNone {
		t.Errorf("expected type to be BlockNone, got %v", s.Type())
	}
}

func TestBlockState_StartBlock_ResetsBuffers(t *testing.T) {
	s := NewBlockState()

	// Start text block and accumulate content
	s.StartBlock(0, makeTestContentBlock("text", ""))
	s.AccumulateText("old content")

	// Start new text block - should reset buffer
	s.StartBlock(1, makeTestContentBlock("text", ""))
	s.AccumulateText("new content")

	if s.TextContent() != "new content" {
		t.Errorf("expected text buffer to be reset, got %q", s.TextContent())
	}
}

func TestBlockState_AccumulateText(t *testing.T) {
	s := NewBlockState()

	// Can't accumulate without starting a text block
	if s.AccumulateText("test") {
		t.Error("expected AccumulateText to return false when not in text block")
	}

	s.StartBlock(0, makeTestContentBlock("text", ""))

	if !s.AccumulateText("Hello ") {
		t.Error("expected AccumulateText to return true in text block")
	}
	if !s.AccumulateText("World!") {
		t.Error("expected AccumulateText to return true in text block")
	}

	if s.TextContent() != "Hello World!" {
		t.Errorf("expected text content to be 'Hello World!', got %q", s.TextContent())
	}
}

func TestBlockState_AccumulateText_IgnoredInToolUseBlock(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("tool_use", "Read"))

	if s.AccumulateText("test") {
		t.Error("expected AccumulateText to return false in tool_use block")
	}
}

func TestBlockState_AccumulateToolInput(t *testing.T) {
	s := NewBlockState()

	// Can't accumulate without starting a tool_use block
	if s.AccumulateToolInput(`{"test": true}`) {
		t.Error("expected AccumulateToolInput to return false when not in tool_use block")
	}

	s.StartBlock(0, makeTestContentBlock("tool_use", "Bash"))

	if !s.AccumulateToolInput(`{"command": `) {
		t.Error("expected AccumulateToolInput to return true in tool_use block")
	}
	if !s.AccumulateToolInput(`"ls -la"}`) {
		t.Error("expected AccumulateToolInput to return true in tool_use block")
	}

	expected := `{"command": "ls -la"}`
	if s.ToolInput() != expected {
		t.Errorf("expected tool input to be %q, got %q", expected, s.ToolInput())
	}
}

func TestBlockState_AccumulateToolInput_IgnoredInTextBlock(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("text", ""))

	if s.AccumulateToolInput(`{"test": true}`) {
		t.Error("expected AccumulateToolInput to return false in text block")
	}
}

func TestBlockState_StopBlock(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("text", ""))
	s.AccumulateText("content")

	stoppedType := s.StopBlock(0)

	if stoppedType != BlockText {
		t.Errorf("expected StopBlock to return BlockText, got %v", stoppedType)
	}
	// Type is preserved until Reset() is called
	if s.Type() != BlockText {
		t.Errorf("expected type to be preserved after StopBlock, got %v", s.Type())
	}
}

func TestBlockState_StopBlock_WrongIndex(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("text", ""))

	stoppedType := s.StopBlock(1) // Wrong index

	if stoppedType != BlockNone {
		t.Errorf("expected StopBlock to return BlockNone for wrong index, got %v", stoppedType)
	}
}

func TestBlockState_Reset(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("tool_use", "Read"))
	s.AccumulateToolInput(`{"test": true}`)

	s.Reset()

	if s.Type() != BlockNone {
		t.Errorf("expected type to be BlockNone after reset, got %v", s.Type())
	}
	if s.Index() != -1 {
		t.Errorf("expected index to be -1 after reset, got %d", s.Index())
	}
	if s.ToolName() != "" {
		t.Errorf("expected tool name to be empty after reset, got %q", s.ToolName())
	}
	if s.InTextBlock() {
		t.Error("expected InTextBlock to be false after reset")
	}
	if s.InToolUseBlock() {
		t.Error("expected InToolUseBlock to be false after reset")
	}
}

func TestBlockState_ResetMessage(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(5, makeTestContentBlock("text", ""))

	s.ResetMessage()

	// Only index is reset, type is preserved
	if s.Index() != -1 {
		t.Errorf("expected index to be -1 after ResetMessage, got %d", s.Index())
	}
	if s.Type() != BlockText {
		t.Errorf("expected type to be preserved after ResetMessage, got %v", s.Type())
	}
}

func TestBlockState_ParseToolInput_Valid(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("tool_use", "Read"))
	s.AccumulateToolInput(`{"file_path": "/test.go", "count": 42}`)

	input, ok := s.ParseToolInput()

	if !ok {
		t.Error("expected ParseToolInput to return true for valid JSON")
	}
	if input["file_path"] != "/test.go" {
		t.Errorf("expected file_path to be '/test.go', got %v", input["file_path"])
	}
	// JSON numbers are parsed as float64
	if input["count"] != float64(42) {
		t.Errorf("expected count to be 42, got %v", input["count"])
	}
}

func TestBlockState_ParseToolInput_Invalid(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("tool_use", "Read"))
	s.AccumulateToolInput(`invalid json`)

	input, ok := s.ParseToolInput()

	if ok {
		t.Error("expected ParseToolInput to return false for invalid JSON")
	}
	if input != nil {
		t.Errorf("expected nil input for invalid JSON, got %v", input)
	}
}

func TestBlockState_ParseToolInput_Empty(t *testing.T) {
	s := NewBlockState()
	s.StartBlock(0, makeTestContentBlock("tool_use", "Read"))
	// Don't accumulate anything

	input, ok := s.ParseToolInput()

	if ok {
		t.Error("expected ParseToolInput to return false for empty buffer")
	}
	if input != nil {
		t.Errorf("expected nil input for empty buffer, got %v", input)
	}
}

func TestBlockState_FullTextFlow(t *testing.T) {
	s := NewBlockState()

	// Start block
	if !s.StartBlock(0, makeTestContentBlock("text", "")) {
		t.Fatal("failed to start text block")
	}

	// Accumulate deltas
	s.AccumulateText("Hello ")
	s.AccumulateText("World!")

	// Verify state during accumulation
	if !s.InTextBlock() {
		t.Error("expected to be in text block")
	}
	if s.TextContent() != "Hello World!" {
		t.Errorf("expected 'Hello World!', got %q", s.TextContent())
	}

	// Stop block
	stoppedType := s.StopBlock(0)
	if stoppedType != BlockText {
		t.Errorf("expected stopped type to be BlockText, got %v", stoppedType)
	}

	// Content still available after stop
	if s.TextContent() != "Hello World!" {
		t.Errorf("expected content to persist after stop, got %q", s.TextContent())
	}

	// Reset for next message
	s.Reset()
	if s.InTextBlock() {
		t.Error("expected not to be in text block after reset")
	}
}

func TestBlockState_FullToolUseFlow(t *testing.T) {
	s := NewBlockState()

	// Start block
	if !s.StartBlock(0, makeTestContentBlock("tool_use", "Bash")) {
		t.Fatal("failed to start tool_use block")
	}

	// Accumulate JSON deltas
	s.AccumulateToolInput(`{"command": `)
	s.AccumulateToolInput(`"ls -la"}`)

	// Verify state during accumulation
	if !s.InToolUseBlock() {
		t.Error("expected to be in tool_use block")
	}
	if s.ToolName() != "Bash" {
		t.Errorf("expected tool name 'Bash', got %q", s.ToolName())
	}

	// Stop block
	stoppedType := s.StopBlock(0)
	if stoppedType != BlockToolUse {
		t.Errorf("expected stopped type to be BlockToolUse, got %v", stoppedType)
	}

	// Parse accumulated input
	input, ok := s.ParseToolInput()
	if !ok {
		t.Fatal("failed to parse tool input")
	}
	if input["command"] != "ls -la" {
		t.Errorf("expected command 'ls -la', got %v", input["command"])
	}

	// Reset for next message
	s.Reset()
	if s.InToolUseBlock() {
		t.Error("expected not to be in tool_use block after reset")
	}
}

func TestBlockState_SwitchingBlockTypes(t *testing.T) {
	s := NewBlockState()

	// Start text block
	s.StartBlock(0, makeTestContentBlock("text", ""))
	s.AccumulateText("text content")

	if !s.InTextBlock() {
		t.Error("expected to be in text block")
	}

	// Switch to tool_use block (simulates new content_block_start)
	s.StartBlock(1, makeTestContentBlock("tool_use", "Read"))
	s.AccumulateToolInput(`{"test": true}`)

	if s.InTextBlock() {
		t.Error("expected not to be in text block after switching")
	}
	if !s.InToolUseBlock() {
		t.Error("expected to be in tool_use block after switching")
	}
	if s.Index() != 1 {
		t.Errorf("expected index to be 1 after switching, got %d", s.Index())
	}
}
