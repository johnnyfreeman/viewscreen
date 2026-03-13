package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/textutil"
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
	highlighter  render.CodeHighlighter
	config       config.Provider
}

// NewWriteRenderer creates a new WriteRenderer with the given dependencies.
func NewWriteRenderer(styleApplier render.StyleApplier, highlighter render.CodeHighlighter, cfg config.Provider) *WriteRenderer {
	return &WriteRenderer{
		styleApplier: styleApplier,
		highlighter:  highlighter,
		config:       cfg,
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

	level := wr.config.GetVerboseLevel()

	// Write tools: -v = 10 lines, -vv = no limit
	var maxLines int
	switch {
	case level >= 2:
		maxLines = -1
	case level >= 1:
		maxLines = 10
	}

	// Always show the summary header
	summary := fmt.Sprintf("Created (%d lines)", lineCount)
	fmt.Fprintf(ctx.Output, "%s%s\n", ctx.OutputPrefix, wr.styleApplier.MutedText(summary))

	// Show content at -v or higher
	if maxLines != 0 && writeResult.Content != "" {
		highlighted := wr.highlighter.HighlightFile(writeResult.Content, writeResult.FilePath)

		if maxLines < 0 {
			pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)
			for _, line := range strings.Split(highlighted, "\n") {
				pw.WriteLine(line)
			}
		} else {
			truncated, remaining := textutil.TruncateLines(highlighted, maxLines)
			pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)
			for _, line := range strings.Split(truncated, "\n") {
				pw.WriteLine(line)
			}
			if remaining > 0 {
				pw.WriteLinef("%s", wr.styleApplier.MutedText(textutil.TruncationIndicator(remaining)))
			}
		}
	}

	return true
}
