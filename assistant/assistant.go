package assistant

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/terminal"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
)

// MarkdownRenderer is an alias for types.MarkdownRenderer for backward compatibility.
type MarkdownRenderer = types.MarkdownRenderer

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


// ToolUseRenderer is a function type for rendering tool use blocks to a writer
type ToolUseRenderer func(out *render.Output, block types.ContentBlock)

// Renderer handles rendering assistant events
type Renderer struct {
	output           io.Writer
	markdownRenderer types.MarkdownRenderer
	toolUseRenderer  ToolUseRenderer
	config           config.Provider
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
func WithMarkdownRenderer(mr types.MarkdownRenderer) RendererOption {
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

// WithConfigProvider sets a custom config provider
func WithConfigProvider(cp config.Provider) RendererOption {
	return func(r *Renderer) {
		r.config = cp
	}
}

// defaultToolUseRenderer renders a tool_use block to an Output using HeaderRenderer.
// It uses HeaderRenderer.RenderBlockToOutput which writes directly to the output,
// avoiding string allocation.
func defaultToolUseRenderer(out *render.Output, block types.ContentBlock) {
	tools.NewHeaderRenderer().RenderBlockToOutput(out, block)
}

// NewRenderer creates a new assistant Renderer with default dependencies
func NewRenderer() *Renderer {
	cfg := config.DefaultProvider{}
	return &Renderer{
		output:           os.Stdout,
		markdownRenderer: render.NewMarkdownRenderer(cfg.NoColor(), terminal.Width()),
		toolUseRenderer:  defaultToolUseRenderer,
		config:           cfg,
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

// renderTo is the unified rendering method that writes to any output.
// This eliminates duplication between Render and RenderToString.
func (r *Renderer) renderTo(out *render.Output, event Event, inTextBlock, inToolUseBlock bool) {
	if event.Error != "" {
		fmt.Fprintln(out, style.ApplyErrorGradient(style.Bullet+"Error"))
		fmt.Fprintf(out, "%s%s\n", style.OutputPrefix, style.ErrorText(event.Error))
	}

	for _, block := range event.Message.Content {
		switch block.Type {
		case "text":
			// Only render if we weren't streaming (text would already be shown)
			if !inTextBlock {
				rendered := r.markdownRenderer.Render(block.Text)
				fmt.Fprint(out, rendered)
				if !strings.HasSuffix(rendered, "\n") {
					fmt.Fprintln(out)
				}
			}
		case "tool_use":
			// Only render if we weren't streaming
			if !inToolUseBlock {
				r.toolUseRenderer(out, block)
			}
		}
	}
}

// Render outputs the assistant event to the terminal
// inTextBlock and inToolUseBlock indicate whether we were streaming these block types
func (r *Renderer) Render(event Event, inTextBlock, inToolUseBlock bool) {
	r.renderTo(render.WriterOutput(r.output), event, inTextBlock, inToolUseBlock)
}

// RenderToString renders the assistant event to a string
// inTextBlock and inToolUseBlock indicate whether we were streaming these block types
func (r *Renderer) RenderToString(event Event, inTextBlock, inToolUseBlock bool) string {
	out := render.StringOutput()
	r.renderTo(out, event, inTextBlock, inToolUseBlock)
	return out.String()
}
