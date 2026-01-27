package user

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/content"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/terminal"
	"github.com/johnnyfreeman/viewscreen/textutil"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
)

// ToolResultContent represents tool result content
type ToolResultContent struct {
	Type       string          `json:"type"`
	ToolUseID  string          `json:"tool_use_id"`
	Text       string          `json:"text"`    // For synthetic text messages
	RawContent json.RawMessage `json:"content"` // For tool results
	IsError    bool            `json:"is_error"`
}

// Content returns the content as a string, handling both string and array formats.
func (t *ToolResultContent) Content() string {
	// For synthetic text messages, return the Text field directly
	if t.Text != "" {
		return t.Text
	}
	return content.ExtractText(t.RawContent)
}

// Message represents the message object in user events
type Message struct {
	Role    string              `json:"role"`
	Content []ToolResultContent `json:"content"`
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

// MarkdownRenderer is an alias for types.MarkdownRenderer for backward compatibility.
type MarkdownRenderer = types.MarkdownRenderer

// Renderer handles rendering user events
type Renderer struct {
	output           io.Writer
	configChecker    ConfigChecker
	styleApplier     render.StyleApplier
	highlighter      CodeHighlighter
	markdownRenderer MarkdownRenderer
	toolContext      *tools.ToolContext
	contentCleaner   *textutil.ContentCleaner
	// Registry for result-specific renderers
	resultRegistry *ResultRegistry
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
func WithStyleApplier(sa render.StyleApplier) RendererOption {
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
func WithToolContext(tc *tools.ToolContext) RendererOption {
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

// WithContentCleaner sets a custom content cleaner
func WithContentCleaner(cc *textutil.ContentCleaner) RendererOption {
	return func(r *Renderer) {
		r.contentCleaner = cc
	}
}

// NewRenderer creates a new user Renderer with default dependencies
func NewRenderer() *Renderer {
	cc := DefaultConfigChecker{}
	sa := render.DefaultStyleApplier{}
	ch := NewDefaultCodeHighlighter(cc.NoColor())

	// Build the result registry with renderers in priority order
	registry := NewResultRegistry()
	registry.Register(NewEditRenderer(sa, ch))
	registry.Register(NewWriteRenderer(sa))
	registry.Register(NewTodoRenderer(sa))

	return &Renderer{
		output:           os.Stdout,
		configChecker:    cc,
		styleApplier:     sa,
		highlighter:      ch,
		markdownRenderer: render.NewMarkdownRenderer(cc.NoColor(), terminal.Width()),
		toolContext:      &tools.ToolContext{},
		contentCleaner:   textutil.DefaultContentCleaner(),
		resultRegistry:   registry,
	}
}

// NewRendererWithOptions creates a new user Renderer with custom options
func NewRendererWithOptions(opts ...RendererOption) *Renderer {
	r := NewRenderer()
	for _, opt := range opts {
		opt(r)
	}
	// Rebuild result registry with potentially updated dependencies
	r.resultRegistry = NewResultRegistry()
	r.resultRegistry.Register(NewEditRenderer(r.styleApplier, r.highlighter))
	r.resultRegistry.Register(NewWriteRenderer(r.styleApplier))
	r.resultRegistry.Register(NewTodoRenderer(r.styleApplier))
	return r
}

// SetToolContext sets the tool context for syntax highlighting
func (r *Renderer) SetToolContext(ctx tools.ToolContext) {
	*r.toolContext = ctx
}

// Render outputs the user event to the terminal
func (r *Renderer) Render(event Event) {
	out := render.WriterOutput(r.output)
	r.renderTo(out, event, r.styleApplier.OutputPrefix(), r.styleApplier.OutputContinue())
}

// RenderNested outputs the user event with nested indentation for sub-agent tools
func (r *Renderer) RenderNested(event Event) {
	out := render.WriterOutput(r.output)
	r.renderTo(out, event, style.NestedOutputPrefix, style.NestedOutputContinue)
}

// renderTo is the unified rendering method that writes to any output.
// This eliminates duplication between Render and RenderToString.
func (r *Renderer) renderTo(out *render.Output, event Event, outputPrefix, outputContinue string) {
	// Handle synthetic messages (e.g., skill content) in verbose mode
	if event.IsSynthetic {
		if r.configChecker.IsVerbose() {
			r.renderSyntheticMessageTo(out, event)
		}
		return
	}

	// Try specialized result renderers via registry
	ctx := &RenderContext{
		Output:         out,
		OutputPrefix:   outputPrefix,
		OutputContinue: outputContinue,
	}
	if r.resultRegistry.TryRender(ctx, event.ToolUseResult) {
		return
	}

	for _, content := range event.Message.Content {
		contentStr := content.Content()
		if content.IsError {
			// Show error with output prefix
			errMsg := r.contentCleaner.Clean(contentStr)
			errMsg = textutil.Truncate(errMsg, 200)
			fmt.Fprintf(out, "%s%s\n", outputPrefix, r.styleApplier.ErrorRender(errMsg))
		} else if contentStr != "" {
			// Clean up the content using the content cleaner pipeline
			cleaned := r.contentCleaner.Clean(contentStr)

			lines := strings.Split(cleaned, "\n")
			lineCount := len(lines)

			if r.configChecker.IsVerbose() {
				// Apply syntax highlighting first
				highlighted := r.highlightContent(cleaned)

				// Truncate to max lines
				truncated, remaining := textutil.TruncateLines(highlighted, textutil.DefaultMaxLines)
				resultLines := strings.Split(truncated, "\n")

				pw := textutil.NewPrefixedWriter(out, outputPrefix, outputContinue)
				for _, line := range resultLines {
					pw.WriteLine(line)
				}

				// Show truncation indicator if content was truncated
				if remaining > 0 {
					indicator := fmt.Sprintf("… (%d more lines)", remaining)
					pw.WriteLinef("%s", r.styleApplier.MutedRender(indicator))
				}
			} else {
				// Show summary in non-verbose mode
				summary := fmt.Sprintf("Read %d lines", lineCount)
				fmt.Fprintf(out, "%s%s\n", outputPrefix, r.styleApplier.MutedRender(summary))
			}
		}
	}
}

// renderSyntheticMessageTo renders a synthetic user message to any output.
func (r *Renderer) renderSyntheticMessageTo(out *render.Output, event Event) {
	for _, content := range event.Message.Content {
		// Synthetic messages have type "text" with Text field populated
		if content.Type == "text" && content.Text != "" {
			cleaned := r.contentCleaner.Clean(content.Text)
			lines := strings.Split(cleaned, "\n")

			// Render as markdown if renderer is available
			if r.markdownRenderer != nil {
				rendered := r.markdownRenderer.Render(cleaned)
				fmt.Fprint(out, rendered)
				if !strings.HasSuffix(rendered, "\n") {
					fmt.Fprintln(out)
				}
			} else {
				// Fallback to plain text with truncation
				truncated, remaining := textutil.TruncateLines(cleaned, textutil.DefaultMaxLines)
				resultLines := strings.Split(truncated, "\n")

				pw := textutil.NewPrefixedWriter(out, r.styleApplier.OutputPrefix(), r.styleApplier.OutputContinue())
				for _, line := range resultLines {
					pw.WriteLine(line)
				}

				if remaining > 0 {
					indicator := fmt.Sprintf("… (%d more lines)", remaining)
					pw.WriteLinef("%s", r.styleApplier.MutedRender(indicator))
					return
				}
			}

			// Show line count summary
			if len(lines) > 0 {
				summary := fmt.Sprintf("(%d lines)", len(lines))
				fmt.Fprintf(out, "%s%s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender(summary))
			}
		}
	}
}

// highlightContent applies syntax highlighting based on context
func (r *Renderer) highlightContent(content string) string {
	// Try to detect language from the last tool's file path
	if r.toolContext != nil && r.toolContext.FilePath != "" {
		lang := render.DetectLanguageFromPath(r.toolContext.FilePath)
		if lang != "" {
			return r.highlighter.Highlight(content, lang)
		}
	}

	// Try to auto-detect from content
	path := ""
	if r.toolContext != nil {
		path = r.toolContext.FilePath
	}
	return r.highlighter.HighlightFile(content, path)
}

// RenderNestedToString renders the user event with nested indentation
func (r *Renderer) RenderNestedToString(event Event) string {
	out := render.StringOutput()
	r.renderTo(out, event, style.NestedOutputPrefix, style.NestedOutputContinue)
	return out.String()
}

// RenderToString renders the user event to a string
func (r *Renderer) RenderToString(event Event) string {
	out := render.StringOutput()
	r.renderTo(out, event, r.styleApplier.OutputPrefix(), r.styleApplier.OutputContinue())
	return out.String()
}
