package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/johnnyfreeman/viewscreen/render"
)

// WriteResult represents the tool_use_result for Write operations
type WriteResult struct {
	Type     string `json:"type"` // "create" for new files
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
}

// WriteRenderer handles rendering of write/create results.
type WriteRenderer struct {
	styleApplier render.StyleApplier
}

// NewWriteRenderer creates a new WriteRenderer with the given dependencies.
func NewWriteRenderer(styleApplier render.StyleApplier) *WriteRenderer {
	return &WriteRenderer{
		styleApplier: styleApplier,
	}
}

// TryRender implements ResultRenderer interface.
// Attempts to render a write/create result with a concise summary.
// Returns true if it was a write/create result and was rendered, false otherwise.
func (wr *WriteRenderer) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	if len(toolUseResult) == 0 {
		return false
	}

	var writeResult WriteResult
	if err := json.Unmarshal(toolUseResult, &writeResult); err != nil {
		return false
	}

	// Check if this is a write/create result
	if writeResult.Type != "create" || writeResult.FilePath == "" {
		return false
	}

	// Count lines in the created file
	lineCount := 1
	if writeResult.Content != "" {
		lineCount = len(strings.Split(writeResult.Content, "\n"))
	}

	// Show a summary of the created file
	summary := fmt.Sprintf("Created (%d lines)", lineCount)
	fmt.Fprintf(ctx.Output, "%s%s\n", ctx.OutputPrefix, wr.styleApplier.UVMutedText(summary))

	return true
}
