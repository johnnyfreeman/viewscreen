package assistant

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
	"golang.org/x/term"
)

// Message represents the message object in assistant events
type Message struct {
	Model      string               `json:"model"`
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	Role       string               `json:"role"`
	Content    []types.ContentBlock `json:"content"`
	StopReason *string              `json:"stop_reason"`
	Usage      *types.Usage         `json:"usage"`
}

// Event represents an assistant message event
type Event struct {
	types.BaseEvent
	Message Message `json:"message"`
	Error   string  `json:"error,omitempty"`
}

// MarkdownRendererInterface abstracts markdown rendering for testability
type MarkdownRendererInterface interface {
	Render(content string) string
}

// ToolUseRenderer is a function type for rendering tool use blocks
type ToolUseRenderer func(block types.ContentBlock)

// Renderer handles rendering assistant events
type Renderer struct {
	output           io.Writer
	markdownRenderer MarkdownRendererInterface
	toolUseRenderer  ToolUseRenderer
}

// RendererOption is a functional option for configuring a Renderer
type RendererOption func(*Renderer)

// WithOutput sets a custom output writer
func WithOutput(w io.Writer) RendererOption {
	return func(r *Renderer) {
		r.output = w
	}
}

// WithMarkdownRenderer sets a custom markdown renderer
func WithMarkdownRenderer(mr MarkdownRendererInterface) RendererOption {
	return func(r *Renderer) {
		r.markdownRenderer = mr
	}
}

// WithToolUseRenderer sets a custom tool use renderer
func WithToolUseRenderer(tr ToolUseRenderer) RendererOption {
	return func(r *Renderer) {
		r.toolUseRenderer = tr
	}
}

// NewRenderer creates a new assistant Renderer with default dependencies
func NewRenderer() *Renderer {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	return &Renderer{
		output:           os.Stdout,
		markdownRenderer: render.NewMarkdownRenderer(config.NoColor, width),
		toolUseRenderer:  tools.RenderToolUseDefault,
	}
}

// NewRendererWithOptions creates a new assistant Renderer with custom options
func NewRendererWithOptions(opts ...RendererOption) *Renderer {
	r := NewRenderer()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Render outputs the assistant event to the terminal
// inTextBlock and inToolUseBlock indicate whether we were streaming these block types
func (r *Renderer) Render(event Event, inTextBlock, inToolUseBlock bool) {
	if event.Error != "" {
		fmt.Fprintln(r.output, style.ApplyErrorGradient(style.Bullet+"Error"))
		fmt.Fprintf(r.output, "%s%s\n", style.OutputPrefix, style.Error.Render(event.Error))
	}

	for _, block := range event.Message.Content {
		switch block.Type {
		case "text":
			// Only render if we weren't streaming (text would already be shown)
			if !inTextBlock {
				// Use markdown renderer for non-streamed text
				rendered := r.markdownRenderer.Render(block.Text)
				fmt.Fprint(r.output, rendered)
				if !strings.HasSuffix(rendered, "\n") {
					fmt.Fprintln(r.output)
				}
			}
		case "tool_use":
			// Only render if we weren't streaming
			if !inToolUseBlock {
				r.toolUseRenderer(block)
			}
		}
	}
}

// Package-level renderer for backward compatibility
var defaultRenderer *Renderer

func getDefaultRenderer() *Renderer {
	if defaultRenderer == nil {
		defaultRenderer = NewRenderer()
	}
	return defaultRenderer
}

// Render is a package-level convenience function for backward compatibility
// inTextBlock and inToolUseBlock indicate whether we were streaming these block types
func Render(event Event, inTextBlock, inToolUseBlock bool) {
	getDefaultRenderer().Render(event, inTextBlock, inToolUseBlock)
}
