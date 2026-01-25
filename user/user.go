package user

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jfreeman/viewscreen/config"
	"github.com/jfreeman/viewscreen/render"
	"github.com/jfreeman/viewscreen/style"
	"github.com/jfreeman/viewscreen/terminal"
	"github.com/jfreeman/viewscreen/types"
)

// ToolResultContent represents tool result content
type ToolResultContent struct {
	Type       string          `json:"type"`
	ToolUseID  string          `json:"tool_use_id"`
	RawContent json.RawMessage `json:"content"`
	IsError    bool            `json:"is_error"`
}

// ContentBlock represents a single content block when content is an array
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Content returns the content as a string, handling both string and array formats
func (t *ToolResultContent) Content() string {
	if len(t.RawContent) == 0 {
		return ""
	}

	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(t.RawContent, &str); err == nil {
		return str
	}

	// Try to unmarshal as array of content blocks
	var blocks []ContentBlock
	if err := json.Unmarshal(t.RawContent, &blocks); err == nil {
		var parts []string
		for _, block := range blocks {
			if block.Type == "text" && block.Text != "" {
				parts = append(parts, block.Text)
			}
		}
		return strings.Join(parts, "\n")
	}

	// Fallback: return raw string representation
	return string(t.RawContent)
}

// Message represents the message object in user events
type Message struct {
	Role    string              `json:"role"`
	Content []ToolResultContent `json:"content"`
}

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

// Event represents a user (tool result) event
type Event struct {
	types.BaseEvent
	Message       Message         `json:"message"`
	ToolUseResult json.RawMessage `json:"tool_use_result"`
}

// toolContext tracks the last tool used to help with syntax highlighting
var lastToolName string
var lastToolPath string

// SetToolContext sets context about the last tool used (called from stream renderer)
func SetToolContext(toolName, path string) {
	lastToolName = toolName
	lastToolPath = path
}

// codeRenderer is lazily initialized
var codeRenderer *render.CodeRenderer

func getCodeRenderer() *render.CodeRenderer {
	if codeRenderer == nil {
		codeRenderer = render.NewCodeRenderer(config.NoColor)
	}
	return codeRenderer
}

// Render outputs the user event to the terminal
func Render(event Event) {
	// Try to render as edit result with diff first
	if config.Verbose && tryRenderEditResult(event.ToolUseResult) {
		return
	}

	for _, content := range event.Message.Content {
		contentStr := content.Content()
		if content.IsError {
			// Show error with output prefix
			errMsg := terminal.StripSystemReminders(contentStr)
			errMsg = terminal.Truncate(errMsg, 200)
			fmt.Printf("%s%s\n", style.OutputPrefix, style.Error.Render(errMsg))
		} else if contentStr != "" {
			// Clean up the content
			cleaned := terminal.StripSystemReminders(contentStr)
			cleaned = terminal.StripLineNumbers(cleaned)

			lines := strings.Split(cleaned, "\n")
			lineCount := len(lines)

			if config.Verbose {
				// Apply syntax highlighting first
				highlighted := highlightContent(cleaned)

				// Truncate to max lines
				truncated, remaining := terminal.TruncateLines(highlighted, terminal.DefaultMaxLines)
				resultLines := strings.Split(truncated, "\n")

				for i, line := range resultLines {
					if i == 0 {
						fmt.Printf("%s%s\n", style.OutputPrefix, line)
					} else {
						fmt.Printf("%s%s\n", style.OutputContinue, line)
					}
				}

				// Show truncation indicator if content was truncated
				if remaining > 0 {
					indicator := fmt.Sprintf("… (%d more lines)", remaining)
					fmt.Printf("%s%s\n", style.OutputContinue, style.Muted.Render(indicator))
				}
			} else {
				// Show summary in non-verbose mode
				summary := fmt.Sprintf("Read %d lines", lineCount)
				fmt.Printf("%s%s\n", style.OutputPrefix, style.Muted.Render(summary))
			}
		}
	}
}

// highlightContent applies syntax highlighting based on context
func highlightContent(content string) string {
	cr := getCodeRenderer()

	// Try to detect language from the last tool's file path
	if lastToolPath != "" {
		lang := render.DetectLanguageFromPath(lastToolPath)
		if lang != "" {
			return cr.Highlight(content, lang)
		}
	}

	// Try to auto-detect from content
	return cr.HighlightFile(content, lastToolPath)
}

// tryRenderEditResult attempts to render an edit result with delta-style diff
// Returns true if it rendered, false if not an edit result
func tryRenderEditResult(toolUseResult json.RawMessage) bool {
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

	// Get syntax highlighter for the file type
	cr := getCodeRenderer()
	lang := render.DetectLanguageFromPath(editResult.FilePath)

	// Separator character for line numbers
	sep := style.LineNumberSep.Render("│")

	first := true
	lineCount := 0
	for _, hunk := range editResult.StructuredPatch {
		oldLine := hunk.OldStart
		newLine := hunk.NewStart

		for _, line := range hunk.Lines {
			if len(line) == 0 {
				continue
			}

			// Check truncation limit
			if lineCount >= terminal.DefaultMaxLines {
				// Count remaining lines
				remaining := 0
				for _, h := range editResult.StructuredPatch {
					remaining += len(h.Lines)
				}
				remaining -= lineCount
				if remaining > 0 {
					indicator := fmt.Sprintf("… (%d more lines)", remaining)
					fmt.Printf("%s%s\n", style.OutputContinue, style.Muted.Render(indicator))
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
				op = style.Success.Render("+")
				newLine++
			case '-':
				// Removed line: show old line number with - indicator
				lineNum = fmt.Sprintf("%*d", numWidth, oldLine)
				op = style.Error.Render("-")
				oldLine++
			default:
				// Context line: show new line number with space
				lineNum = fmt.Sprintf("%*d", numWidth, newLine)
				op = " "
				oldLine++
				newLine++
			}
			lineNums := style.LineNumber.Render(lineNum)

			// Syntax highlight with appropriate background for diff lines
			var styled string
			switch prefix {
			case '+':
				if lang != "" {
					styled = cr.HighlightWithBg(content, lang, style.DiffAddBg)
				} else {
					styled = style.DiffAdd.Render(content)
				}
			case '-':
				if lang != "" {
					styled = cr.HighlightWithBg(content, lang, style.DiffRemoveBg)
				} else {
					styled = style.DiffRemove.Render(content)
				}
			default:
				if lang != "" {
					styled = cr.Highlight(content, lang)
				} else {
					styled = content
				}
			}

			// Output with separators: ⎿ 123 │ + code
			if first {
				fmt.Printf("%s%s %s %s %s\n", style.OutputPrefix, lineNums, sep, op, styled)
				first = false
			} else {
				fmt.Printf("%s%s %s %s %s\n", style.OutputContinue, lineNums, sep, op, styled)
			}
			lineCount++
		}
	}
	return true
}
