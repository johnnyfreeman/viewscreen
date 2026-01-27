package stream

import (
	"encoding/json"
	"strings"

	"github.com/johnnyfreeman/viewscreen/types"
)

// BlockType represents the type of content block being processed
type BlockType int

const (
	BlockNone BlockType = iota
	BlockText
	BlockToolUse
)

// String returns the string representation of a BlockType
func (b BlockType) String() string {
	switch b {
	case BlockText:
		return "text"
	case BlockToolUse:
		return "tool_use"
	default:
		return "none"
	}
}

// BlockState tracks the state of streaming content blocks.
// It manages block transitions, content accumulation, and provides
// a clean interface for querying block state.
type BlockState struct {
	blockType  BlockType
	blockIndex int
	toolName   string
	textBuf    strings.Builder
	toolBuf    strings.Builder
}

// NewBlockState creates a new BlockState with initial values
func NewBlockState() *BlockState {
	return &BlockState{
		blockType:  BlockNone,
		blockIndex: -1,
	}
}

// Type returns the current block type
func (s *BlockState) Type() BlockType {
	return s.blockType
}

// Index returns the current block index
func (s *BlockState) Index() int {
	return s.blockIndex
}

// ToolName returns the current tool name (only valid when Type is BlockToolUse)
func (s *BlockState) ToolName() string {
	return s.toolName
}

// InTextBlock returns true if currently processing a text block
func (s *BlockState) InTextBlock() bool {
	return s.blockType == BlockText
}

// InToolUseBlock returns true if currently processing a tool_use block
func (s *BlockState) InToolUseBlock() bool {
	return s.blockType == BlockToolUse
}

// TextContent returns the accumulated text content
func (s *BlockState) TextContent() string {
	return s.textBuf.String()
}

// ToolInput returns the accumulated tool input JSON
func (s *BlockState) ToolInput() string {
	return s.toolBuf.String()
}

// StartBlock begins tracking a new content block.
// Returns true if the block was successfully started.
func (s *BlockState) StartBlock(index int, contentBlock json.RawMessage) bool {
	s.blockIndex = index
	s.blockType = BlockNone
	s.toolName = ""

	if len(contentBlock) == 0 {
		return false
	}

	var block types.ContentBlock
	if err := json.Unmarshal(contentBlock, &block); err != nil {
		return false
	}

	switch block.Type {
	case "text":
		s.blockType = BlockText
		s.textBuf.Reset()
		return true
	case "tool_use":
		s.blockType = BlockToolUse
		s.toolName = block.Name
		s.toolBuf.Reset()
		return true
	}

	return false
}

// AccumulateText adds text to the text buffer.
// Returns true if text was accumulated (i.e., we're in a text block).
func (s *BlockState) AccumulateText(text string) bool {
	if s.blockType != BlockText {
		return false
	}
	s.textBuf.WriteString(text)
	return true
}

// AccumulateToolInput adds JSON to the tool input buffer.
// Returns true if input was accumulated (i.e., we're in a tool_use block).
func (s *BlockState) AccumulateToolInput(partialJSON string) bool {
	if s.blockType != BlockToolUse {
		return false
	}
	s.toolBuf.WriteString(partialJSON)
	return true
}

// StopBlock finalizes the current block if the index matches.
// Returns the block type that was stopped, or BlockNone if indices don't match.
func (s *BlockState) StopBlock(index int) BlockType {
	if index != s.blockIndex {
		return BlockNone
	}
	stoppedType := s.blockType
	// Don't reset blockType here - it's needed by Renderer to check
	// after assistant events. Use Reset() explicitly when done.
	return stoppedType
}

// Reset clears the block state after processing is complete.
// Call this after an assistant event to prepare for the next message.
func (s *BlockState) Reset() {
	s.blockType = BlockNone
	s.blockIndex = -1
	s.toolName = ""
}

// ResetMessage resets state for a new message (on message_stop).
// This only resets the block index, preserving type info for post-processing.
func (s *BlockState) ResetMessage() {
	s.blockIndex = -1
}

// ParseToolInput parses the accumulated tool input as JSON.
// Returns the parsed map and true on success, nil and false on failure.
func (s *BlockState) ParseToolInput() (map[string]any, bool) {
	var input map[string]any
	if err := json.Unmarshal([]byte(s.toolBuf.String()), &input); err != nil {
		return nil, false
	}
	return input, true
}
