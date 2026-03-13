package user

import (
	"encoding/json"
	"fmt"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/textutil"
)

// PatchHunk represents a single hunk in a structured patch
type PatchHunk struct {
	OldStart int      `json:"oldStart"`
	OldLines int      `json:"oldLines"`
	NewStart int      `json:"newStart"`
	NewLines int      `json:"newLines"`
	Lines    []string `json:"lines"`
}

// EditResult represents the tool_use_result for Edit operations
type EditResult struct {
	FilePath        string      `json:"filePath"`
	OldString       string      `json:"oldString"`
	NewString       string      `json:"newString"`
	StructuredPatch []PatchHunk `json:"structuredPatch"`
}

// EditRenderer handles rendering of edit results with syntax-highlighted diffs.
type EditRenderer struct {
	styleApplier render.StyleApplier
	highlighter  render.CodeHighlighter
	config       config.Provider
}

// NewEditRenderer creates a new EditRenderer with the given dependencies.
func NewEditRenderer(styleApplier render.StyleApplier, highlighter render.CodeHighlighter, cfg config.Provider) *EditRenderer {
	return &EditRenderer{
		styleApplier: styleApplier,
		highlighter:  highlighter,
		config:       cfg,
	}
}

// TryRender implements ResultRenderer interface.
// Attempts to render an edit result with syntax-highlighted diff.
// Returns true if it was an edit result and was rendered, false otherwise.
func (er *EditRenderer) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	if len(toolUseResult) == 0 {
		return false
	}

	var editResult EditResult
	if err := json.Unmarshal(toolUseResult, &editResult); err != nil {
		return false
	}

	// Check if this is an edit result with a structured patch
	if editResult.FilePath == "" || len(editResult.StructuredPatch) == 0 {
		return false
	}

	// Calculate max line number for column width
	maxLine := 0
	for _, hunk := range editResult.StructuredPatch {
		if endOld := hunk.OldStart + hunk.OldLines; endOld > maxLine {
			maxLine = endOld
		}
		if endNew := hunk.NewStart + hunk.NewLines; endNew > maxLine {
			maxLine = endNew
		}
	}
	numWidth := len(fmt.Sprintf("%d", maxLine))

	// Separator character for line numbers
	sep := er.styleApplier.LineNumberSepRender("│")

	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)
	lineCount := 0

	for _, hunk := range editResult.StructuredPatch {
		oldLine := hunk.OldStart
		newLine := hunk.NewStart

		for _, line := range hunk.Lines {
			if len(line) == 0 {
				continue
			}

			// Write tools: -v = 10 lines, -vv = no limit
			maxLines := 10
			if er.config.GetVerboseLevel() >= 2 {
				maxLines = -1
			}
			if maxLines >= 0 && lineCount >= maxLines {
				// Count remaining lines
				remaining := 0
				for _, h := range editResult.StructuredPatch {
					remaining += len(h.Lines)
				}
				remaining -= lineCount
				if remaining > 0 {
					pw.WriteLinef("%s", er.styleApplier.MutedText(textutil.TruncationIndicator(remaining)))
				}
				return true
			}

			prefix := line[0]
			content := line[1:] // Strip the +/- prefix

			// Format line number and operation indicator
			var lineNum string
			var op string
			switch prefix {
			case '+':
				// Added line: show new line number with + indicator
				lineNum = fmt.Sprintf("%*d", numWidth, newLine)
				op = er.styleApplier.SuccessText("+")
				newLine++
			case '-':
				// Removed line: show old line number with - indicator
				lineNum = fmt.Sprintf("%*d", numWidth, oldLine)
				op = er.styleApplier.ErrorText("-")
				oldLine++
			default:
				// Context line: show new line number with space
				lineNum = fmt.Sprintf("%*d", numWidth, newLine)
				op = " "
				oldLine++
				newLine++
			}
			lineNums := er.styleApplier.LineNumberRender(lineNum)

			// Syntax highlight with appropriate background for diff lines
			// HighlightFileWithBg uses the filename for language detection (via chroma)
			var styled string
			switch prefix {
			case '+':
				styled = er.highlighter.HighlightFileWithBg(content, editResult.FilePath, er.styleApplier.DiffAddBg())
			case '-':
				styled = er.highlighter.HighlightFileWithBg(content, editResult.FilePath, er.styleApplier.DiffRemoveBg())
			default:
				styled = er.highlighter.HighlightFile(content, editResult.FilePath)
			}

			// Output with separators: ⎿ 123 │ + code
			pw.WriteLinef("%s %s %s %s", lineNums, sep, op, styled)
			lineCount++
		}
	}
	return true
}
