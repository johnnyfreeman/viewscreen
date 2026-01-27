package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/johnnyfreeman/viewscreen/render"
)

// WriteRenderer handles rendering of write/create results.
type WriteRenderer struct {
	styleApplier StyleApplier
}

// NewWriteRenderer creates a new WriteRenderer with the given dependencies.
func NewWriteRenderer(styleApplier StyleApplier) *WriteRenderer {
	return &WriteRenderer{
		styleApplier: styleApplier,
	}
}

// TryRender attempts to render a write result to the given output.
// Returns true if it was a write/create result and was rendered, false otherwise.
func (wr *WriteRenderer) TryRender(out *render.Output, toolUseResult json.RawMessage, outputPrefix string) bool {
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
	fmt.Fprintf(out, "%s%s\n", outputPrefix, wr.styleApplier.MutedRender(summary))

	return true
}
