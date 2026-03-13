package user

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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


// Renderer handles rendering user events
type Renderer struct {
	output           io.Writer
	config           config.Provider
	styleApplier     render.StyleApplier
	highlighter      render.CodeHighlighter
	markdownRenderer types.MarkdownRenderer
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

// WithConfigProvider sets a custom config provider
func WithConfigProvider(cp config.Provider) RendererOption {
	return func(r *Renderer) {
		r.config = cp
	}
}

// WithStyleApplier sets a custom style applier
func WithStyleApplier(sa render.StyleApplier) RendererOption {
	return func(r *Renderer) {
		r.styleApplier = sa
	}
}

// WithCodeHighlighter sets a custom code highlighter
func WithCodeHighlighter(ch render.CodeHighlighter) RendererOption {
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
func WithMarkdownRenderer(mr types.MarkdownRenderer) RendererOption {
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

// NewRenderer creates a new user Renderer with the given options
func NewRenderer(opts ...RendererOption) *Renderer {
	cfg := config.Get()
	sa := render.DefaultStyleApplier{}
	ch := render.NewCodeRenderer(cfg.NoColor())

	r := &Renderer{
		output:           os.Stdout,
		config:           cfg,
		styleApplier:     sa,
		highlighter:      ch,
		markdownRenderer: render.NewMarkdownRenderer(cfg.NoColor(), terminal.Width()),
		toolContext:      &tools.ToolContext{},
		contentCleaner:   textutil.DefaultContentCleaner(),
	}
	for _, opt := range opts {
		opt(r)
	}
	// Build result registry with final dependencies (after options applied)
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
		if r.config.IsVerbose() {
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
			fmt.Fprintf(out, "%s%s\n", outputPrefix, r.styleApplier.ErrorText(errMsg))
		} else if contentStr != "" {
			// Clean up the content using the content cleaner pipeline
			cleaned := r.contentCleaner.Clean(contentStr)

			lines := strings.Split(cleaned, "\n")
			lineCount := len(lines)

			// Verbosity levels for tool output:
			// Writes (Edit, Write, NotebookEdit):
			//   -v:   truncated to 10 lines
			//   -vv:  not truncated
			// Reads (everything else):
			//   -v:   no content shown (summary only)
			//   -vv:  truncated to 5 lines
			//   -vvv: truncated to 10 lines
			isWriteTool := r.toolContext != nil && (r.toolContext.ToolName == "Edit" || r.toolContext.ToolName == "Write" || r.toolContext.ToolName == "NotebookEdit")
			level := r.config.GetVerboseLevel()

			var maxLines int // 0 = don't expand, -1 = no limit
			if isWriteTool {
				switch {
				case level >= 2:
					maxLines = -1
				case level >= 1:
					maxLines = 10
				}
			} else {
				switch {
				case level >= 3:
					maxLines = 10
				case level >= 2:
					maxLines = 5
				}
			}

			if maxLines != 0 {
				highlighted := r.highlightContent(cleaned)

				if maxLines < 0 {
					// No truncation
					pw := textutil.NewPrefixedWriter(out, outputPrefix, outputContinue)
					for _, line := range strings.Split(highlighted, "\n") {
						pw.WriteLine(line)
					}
				} else {
					truncated, remaining := textutil.TruncateLines(highlighted, maxLines)
					resultLines := strings.Split(truncated, "\n")

					pw := textutil.NewPrefixedWriter(out, outputPrefix, outputContinue)
					for _, line := range resultLines {
						pw.WriteLine(line)
					}

					if remaining > 0 {
						pw.WriteLinef("%s", r.styleApplier.MutedText(textutil.TruncationIndicator(remaining)))
					}
				}
			} else {
				summary := fmt.Sprintf("%d lines", lineCount)
				fmt.Fprintf(out, "%s%s\n", outputPrefix, r.styleApplier.MutedText(summary))
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
					pw.WriteLinef("%s", r.styleApplier.MutedText(textutil.TruncationIndicator(remaining)))
					return
				}
			}

			// Show line count summary
			if len(lines) > 0 {
				summary := fmt.Sprintf("(%d lines)", len(lines))
				fmt.Fprintf(out, "%s%s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText(summary))
			}
		}
	}
}

// highlightContent applies syntax highlighting based on context
func (r *Renderer) highlightContent(content string) string {
	// Use file path for language detection (chroma handles this internally)
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

// SubAgentPromptMaxLines is the maximum number of lines to show for a sub-agent prompt.
const SubAgentPromptMaxLines = 3

// RenderSubAgentPromptToString renders a sub-agent prompt with truncation.
// Shows the first few lines of the prompt followed by a line count indicator.
func (r *Renderer) RenderSubAgentPromptToString(event Event) string {
	out := render.StringOutput()
	r.renderSubAgentPromptTo(out, event, r.styleApplier.OutputPrefix(), r.styleApplier.OutputContinue())
	return out.String()
}

// RenderNestedSubAgentPromptToString renders a nested sub-agent prompt with truncation.
func (r *Renderer) RenderNestedSubAgentPromptToString(event Event) string {
	out := render.StringOutput()
	r.renderSubAgentPromptTo(out, event, style.NestedOutputPrefix, style.NestedOutputContinue)
	return out.String()
}

// renderSubAgentPromptTo renders a sub-agent prompt with truncation to any output.
func (r *Renderer) renderSubAgentPromptTo(out *render.Output, event Event, outputPrefix, outputContinue string) {
	for _, content := range event.Message.Content {
		if content.Type != "text" {
			continue
		}
		text := content.Text
		if text == "" {
			text = content.Content()
		}
		if text == "" {
			continue
		}

		// Clean and split into lines
		cleaned := r.contentCleaner.Clean(text)
		lines := strings.Split(cleaned, "\n")

		// Remove trailing empty lines
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}

		totalLines := len(lines)
		if totalLines == 0 {
			continue
		}

		pw := textutil.NewPrefixedWriter(out, outputPrefix, outputContinue)

		if totalLines <= SubAgentPromptMaxLines {
			// Show all lines
			for _, line := range lines {
				pw.WriteLine(r.styleApplier.MutedText(line))
			}
		} else {
			// Show first lines, then truncation indicator
			for i := 0; i < SubAgentPromptMaxLines; i++ {
				pw.WriteLine(r.styleApplier.MutedText(lines[i]))
			}
			pw.WriteLine(r.styleApplier.MutedText(fmt.Sprintf("(%d lines)", totalLines)))
		}
	}
}
