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
// Renders a write/create result as a diff with all lines shown as additions.
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

	lines := strings.Split(writeResult.Content, "\n")
	lineCount := len(lines)

	level := wr.config.GetVerboseLevel()

	// Write tools: -v = 10 lines, -vv = no limit
	maxLines := 10
	switch {
	case level >= 2:
		maxLines = -1
	case level >= 1:
		maxLines = 10
	}

	// Always show the summary header
	summary := fmt.Sprintf("Created (%d lines)", lineCount)
	fmt.Fprintf(ctx.Output, "%s%s\n", ctx.OutputPrefix, wr.styleApplier.MutedText(summary))

	if writeResult.Content == "" {
		return true
	}

	// Calculate line number column width
	numWidth := len(fmt.Sprintf("%d", lineCount))
	sep := wr.styleApplier.LineNumberSepRender("│")
	op := wr.styleApplier.SuccessText("+")

	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)

	for i, line := range lines {
		if maxLines >= 0 && i >= maxLines {
			remaining := lineCount - i
			if remaining > 0 {
				pw.WriteLinef("%s", wr.styleApplier.MutedText(textutil.TruncationIndicator(remaining)))
			}
			break
		}

		lineNum := fmt.Sprintf("%*d", numWidth, i+1)
		lineNums := wr.styleApplier.LineNumberRender(lineNum)
		styled := wr.highlighter.HighlightFileWithBg(line, writeResult.FilePath, wr.styleApplier.DiffAddBg())

		pw.WriteLinef("%s %s %s %s", lineNums, sep, op, styled)
	}

	return true
}
