package user

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/terminal"
	"github.com/johnnyfreeman/viewscreen/types"
	"golang.org/x/term"
)

// ToolResultContent represents tool result content
type ToolResultContent struct {
	Type       string          `json:"type"`
	ToolUseID  string          `json:"tool_use_id"`
	Text       string          `json:"text"`    // For synthetic text messages
	RawContent json.RawMessage `json:"content"` // For tool results
	IsError    bool            `json:"is_error"`
}

// ContentBlock represents a single content block when content is an array
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Content returns the content as a string, handling both string and array formats
func (t *ToolResultContent) Content() string {
	// For synthetic text messages, return the Text field directly
	if t.Text != "" {
		return t.Text
	}

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
	IsSynthetic   bool            `json:"isSynthetic"`
}

// ConfigChecker abstracts config flag checking for testability
type ConfigChecker interface {
	IsVerbose() bool
	NoColor() bool
}

// DefaultConfigChecker uses the actual config package
type DefaultConfigChecker struct{}

func (d DefaultConfigChecker) IsVerbose() bool { return config.Verbose }
func (d DefaultConfigChecker) NoColor() bool   { return config.NoColor }

// StyleApplier abstracts style application for testability
type StyleApplier interface {
	ErrorRender(text string) string
	MutedRender(text string) string
	SuccessRender(text string) string
	OutputPrefix() string
	OutputContinue() string
	LineNumberRender(text string) string
	LineNumberSepRender(text string) string
	DiffAddRender(text string) string
	DiffRemoveRender(text string) string
	DiffAddBg() lipgloss.Color
	DiffRemoveBg() lipgloss.Color
}

// DefaultStyleApplier uses the actual style package
type DefaultStyleApplier struct{}

func (d DefaultStyleApplier) ErrorRender(text string) string         { return style.Error.Render(text) }
func (d DefaultStyleApplier) MutedRender(text string) string         { return style.Muted.Render(text) }
func (d DefaultStyleApplier) SuccessRender(text string) string       { return style.Success.Render(text) }
func (d DefaultStyleApplier) OutputPrefix() string                   { return style.OutputPrefix }
func (d DefaultStyleApplier) OutputContinue() string                 { return style.OutputContinue }
func (d DefaultStyleApplier) LineNumberRender(text string) string    { return style.LineNumber.Render(text) }
func (d DefaultStyleApplier) LineNumberSepRender(text string) string { return style.LineNumberSep.Render("│") }
func (d DefaultStyleApplier) DiffAddRender(text string) string       { return style.DiffAdd.Render(text) }
func (d DefaultStyleApplier) DiffRemoveRender(text string) string    { return style.DiffRemove.Render(text) }
func (d DefaultStyleApplier) DiffAddBg() lipgloss.Color              { return style.DiffAddBg }
func (d DefaultStyleApplier) DiffRemoveBg() lipgloss.Color           { return style.DiffRemoveBg }

// CodeHighlighter abstracts code highlighting for testability
type CodeHighlighter interface {
	Highlight(code, language string) string
	HighlightFile(code, filename string) string
	HighlightWithBg(code, language string, bgColor lipgloss.Color) string
}

// DefaultCodeHighlighter uses the actual render package
type DefaultCodeHighlighter struct {
	renderer *render.CodeRenderer
}

func NewDefaultCodeHighlighter(noColor bool) *DefaultCodeHighlighter {
	return &DefaultCodeHighlighter{
		renderer: render.NewCodeRenderer(noColor),
	}
}

func (d *DefaultCodeHighlighter) Highlight(code, language string) string {
	return d.renderer.Highlight(code, language)
}

func (d *DefaultCodeHighlighter) HighlightFile(code, filename string) string {
	return d.renderer.HighlightFile(code, filename)
}

func (d *DefaultCodeHighlighter) HighlightWithBg(code, language string, bgColor lipgloss.Color) string {
	return d.renderer.HighlightWithBg(code, language, bgColor)
}

// ToolContext holds information about the last tool used
type ToolContext struct {
	ToolName string
	ToolPath string
}

// MarkdownRenderer abstracts markdown rendering for testability
type MarkdownRenderer interface {
	Render(content string) string
}

// Renderer handles rendering user events
type Renderer struct {
	output           io.Writer
	configChecker    ConfigChecker
	styleApplier     StyleApplier
	highlighter      CodeHighlighter
	markdownRenderer MarkdownRenderer
	toolContext      *ToolContext
}

// RendererOption is a functional option for configuring a Renderer
type RendererOption func(*Renderer)

// WithOutput sets a custom output writer
func WithOutput(w io.Writer) RendererOption {
	return func(r *Renderer) {
		r.output = w
	}
}

// WithConfigChecker sets a custom config checker
func WithConfigChecker(cc ConfigChecker) RendererOption {
	return func(r *Renderer) {
		r.configChecker = cc
	}
}

// WithStyleApplier sets a custom style applier
func WithStyleApplier(sa StyleApplier) RendererOption {
	return func(r *Renderer) {
		r.styleApplier = sa
	}
}

// WithCodeHighlighter sets a custom code highlighter
func WithCodeHighlighter(ch CodeHighlighter) RendererOption {
	return func(r *Renderer) {
		r.highlighter = ch
	}
}

// WithToolContext sets the tool context for syntax highlighting hints
func WithToolContext(tc *ToolContext) RendererOption {
	return func(r *Renderer) {
		r.toolContext = tc
	}
}

// WithMarkdownRenderer sets a custom markdown renderer
func WithMarkdownRenderer(mr MarkdownRenderer) RendererOption {
	return func(r *Renderer) {
		r.markdownRenderer = mr
	}
}

// NewRenderer creates a new user Renderer with default dependencies
func NewRenderer() *Renderer {
	cc := DefaultConfigChecker{}
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	return &Renderer{
		output:           os.Stdout,
		configChecker:    cc,
		styleApplier:     DefaultStyleApplier{},
		highlighter:      NewDefaultCodeHighlighter(cc.NoColor()),
		markdownRenderer: render.NewMarkdownRenderer(cc.NoColor(), width),
		toolContext:      &ToolContext{},
	}
}

// NewRendererWithOptions creates a new user Renderer with custom options
func NewRendererWithOptions(opts ...RendererOption) *Renderer {
	r := NewRenderer()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SetToolContext sets the tool context for syntax highlighting
func (r *Renderer) SetToolContext(toolName, path string) {
	r.toolContext.ToolName = toolName
	r.toolContext.ToolPath = path
}

// Render outputs the user event to the terminal
func (r *Renderer) Render(event Event) {
	// Handle synthetic messages (e.g., skill content) in verbose mode
	if event.IsSynthetic {
		if r.configChecker.IsVerbose() {
			r.renderSyntheticMessage(event)
		}
		return
	}

	// Try to render as edit result with diff first
	// Always show edit diffs by default - developers want to see what changed
	if r.tryRenderEditResult(event.ToolUseResult) {
		return
	}

	for _, content := range event.Message.Content {
		contentStr := content.Content()
		if content.IsError {
			// Show error with output prefix
			errMsg := terminal.StripSystemReminders(contentStr)
			errMsg = terminal.Truncate(errMsg, 200)
			fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputPrefix(), r.styleApplier.ErrorRender(errMsg))
		} else if contentStr != "" {
			// Clean up the content
			cleaned := terminal.StripSystemReminders(contentStr)
			cleaned = terminal.StripLineNumbers(cleaned)

			lines := strings.Split(cleaned, "\n")
			lineCount := len(lines)

			if r.configChecker.IsVerbose() {
				// Apply syntax highlighting first
				highlighted := r.highlightContent(cleaned)

				// Truncate to max lines
				truncated, remaining := terminal.TruncateLines(highlighted, terminal.DefaultMaxLines)
				resultLines := strings.Split(truncated, "\n")

				for i, line := range resultLines {
					if i == 0 {
						fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputPrefix(), line)
					} else {
						fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputContinue(), line)
					}
				}

				// Show truncation indicator if content was truncated
				if remaining > 0 {
					indicator := fmt.Sprintf("… (%d more lines)", remaining)
					fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender(indicator))
				}
			} else {
				// Show summary in non-verbose mode
				summary := fmt.Sprintf("Read %d lines", lineCount)
				fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputPrefix(), r.styleApplier.MutedRender(summary))
			}
		}
	}
}

// renderSyntheticMessage renders a synthetic user message (e.g., skill content)
func (r *Renderer) renderSyntheticMessage(event Event) {
	for _, content := range event.Message.Content {
		// Synthetic messages have type "text" with Text field populated
		if content.Type == "text" && content.Text != "" {
			cleaned := terminal.StripSystemReminders(content.Text)
			lines := strings.Split(cleaned, "\n")

			// Render as markdown if renderer is available
			if r.markdownRenderer != nil {
				rendered := r.markdownRenderer.Render(cleaned)
				fmt.Fprint(r.output, rendered)
				if !strings.HasSuffix(rendered, "\n") {
					fmt.Fprintln(r.output)
				}
			} else {
				// Fallback to plain text with truncation
				truncated, remaining := terminal.TruncateLines(cleaned, terminal.DefaultMaxLines)
				resultLines := strings.Split(truncated, "\n")

				for i, line := range resultLines {
					if i == 0 {
						fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputPrefix(), line)
					} else {
						fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputContinue(), line)
					}
				}

				if remaining > 0 {
					indicator := fmt.Sprintf("… (%d more lines)", remaining)
					fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender(indicator))
					return
				}
			}

			// Show line count summary
			if len(lines) > 0 {
				summary := fmt.Sprintf("(%d lines)", len(lines))
				fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender(summary))
			}
		}
	}
}

// highlightContent applies syntax highlighting based on context
func (r *Renderer) highlightContent(content string) string {
	// Try to detect language from the last tool's file path
	if r.toolContext != nil && r.toolContext.ToolPath != "" {
		lang := render.DetectLanguageFromPath(r.toolContext.ToolPath)
		if lang != "" {
			return r.highlighter.Highlight(content, lang)
		}
	}

	// Try to auto-detect from content
	path := ""
	if r.toolContext != nil {
		path = r.toolContext.ToolPath
	}
	return r.highlighter.HighlightFile(content, path)
}

// tryRenderEditResult attempts to render an edit result with delta-style diff
// Returns true if it rendered, false if not an edit result
func (r *Renderer) tryRenderEditResult(toolUseResult json.RawMessage) bool {
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

	// Get language for syntax highlighting
	lang := render.DetectLanguageFromPath(editResult.FilePath)

	// Separator character for line numbers
	sep := r.styleApplier.LineNumberSepRender("│")

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
					fmt.Fprintf(r.output, "%s%s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender(indicator))
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
				op = r.styleApplier.SuccessRender("+")
				newLine++
			case '-':
				// Removed line: show old line number with - indicator
				lineNum = fmt.Sprintf("%*d", numWidth, oldLine)
				op = r.styleApplier.ErrorRender("-")
				oldLine++
			default:
				// Context line: show new line number with space
				lineNum = fmt.Sprintf("%*d", numWidth, newLine)
				op = " "
				oldLine++
				newLine++
			}
			lineNums := r.styleApplier.LineNumberRender(lineNum)

			// Syntax highlight with appropriate background for diff lines
			var styled string
			switch prefix {
			case '+':
				if lang != "" {
					styled = r.highlighter.HighlightWithBg(content, lang, r.styleApplier.DiffAddBg())
				} else {
					styled = r.styleApplier.DiffAddRender(content)
				}
			case '-':
				if lang != "" {
					styled = r.highlighter.HighlightWithBg(content, lang, r.styleApplier.DiffRemoveBg())
				} else {
					styled = r.styleApplier.DiffRemoveRender(content)
				}
			default:
				if lang != "" {
					styled = r.highlighter.Highlight(content, lang)
				} else {
					styled = content
				}
			}

			// Output with separators: ⎿ 123 │ + code
			if first {
				fmt.Fprintf(r.output, "%s%s %s %s %s\n", r.styleApplier.OutputPrefix(), lineNums, sep, op, styled)
				first = false
			} else {
				fmt.Fprintf(r.output, "%s%s %s %s %s\n", r.styleApplier.OutputContinue(), lineNums, sep, op, styled)
			}
			lineCount++
		}
	}
	return true
}

// Package-level state for backward compatibility
var (
	lastToolName    string
	lastToolPath    string
	defaultRenderer *Renderer
)

// SetToolContext sets context about the last tool used (called from stream renderer)
// This is the package-level function for backward compatibility
func SetToolContext(toolName, path string) {
	lastToolName = toolName
	lastToolPath = path
	// Also update the default renderer if it exists
	if defaultRenderer != nil {
		defaultRenderer.SetToolContext(toolName, path)
	}
}

func getDefaultRenderer() *Renderer {
	if defaultRenderer == nil {
		defaultRenderer = NewRenderer()
		// Sync the tool context
		defaultRenderer.SetToolContext(lastToolName, lastToolPath)
	}
	return defaultRenderer
}

// Render is a package-level convenience function for backward compatibility
func Render(event Event) {
	getDefaultRenderer().Render(event)
}
