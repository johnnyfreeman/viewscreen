package stream

import (
	"encoding/json"
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

// MarkdownRendererInterface is an alias for types.MarkdownRenderer for backward compatibility.
type MarkdownRendererInterface = types.MarkdownRenderer

// IndicatorInterface abstracts the streaming indicator for testability
type IndicatorInterface interface {
	Show()
	Clear()
}

// ToolHeaderRenderer abstracts tool header rendering for testability.
// It returns the rendered string instead of printing directly.
type ToolHeaderRenderer func(toolName string, input map[string]interface{}) string

// EventData represents the nested event in stream_event
type EventData struct {
	Type         string          `json:"type"`
	Index        int             `json:"index"`
	Message      json.RawMessage `json:"message,omitempty"`
	ContentBlock json.RawMessage `json:"content_block,omitempty"`
	Delta        json.RawMessage `json:"delta,omitempty"`
	Usage        *types.Usage    `json:"usage,omitempty"`
}

// Event represents a streaming event
type Event struct {
	types.BaseEvent
	Event EventData `json:"event"`
}

// TextDelta represents a text delta in streaming
type TextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// InputJSONDelta represents a JSON input delta in streaming
type InputJSONDelta struct {
	Type        string `json:"type"`
	PartialJSON string `json:"partial_json"`
}

// MessageDelta represents a message delta
type MessageDelta struct {
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
}

// Renderer handles rendering stream events and tracks streaming state
type Renderer struct {
	CurrentBlockIndex int
	CurrentBlockType  string
	InTextBlock       bool
	InToolUseBlock    bool
	toolName          string
	toolInput         strings.Builder
	textBuffer        strings.Builder
	markdownRenderer  MarkdownRendererInterface
	indicator         IndicatorInterface
	toolHeaderRender  ToolHeaderRenderer
	output            io.Writer
	width             int
}

// NewRenderer creates a new stream Renderer
func NewRenderer() *Renderer {
	// Get terminal width for markdown wrapping
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}

	return &Renderer{
		CurrentBlockIndex: -1,
		markdownRenderer:  render.NewMarkdownRenderer(config.NoColor, width),
		indicator:         render.NewStreamingIndicator(config.NoColor),
		toolHeaderRender:  tools.RenderToolHeaderToString,
		output:            os.Stdout,
		width:             width,
	}
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

// WithIndicator sets a custom streaming indicator
func WithIndicator(i IndicatorInterface) RendererOption {
	return func(r *Renderer) {
		r.indicator = i
	}
}

// WithToolHeaderRenderer sets a custom tool header renderer
func WithToolHeaderRenderer(f ToolHeaderRenderer) RendererOption {
	return func(r *Renderer) {
		r.toolHeaderRender = f
	}
}

// NewRendererWithOptions creates a new stream Renderer with custom options
func NewRendererWithOptions(opts ...RendererOption) *Renderer {
	r := NewRenderer()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// renderTo implements the core rendering logic, writing to the provided output.
// This eliminates duplication between Render and RenderToString.
func (r *Renderer) renderTo(out *render.Output, event Event, showIndicator bool) {
	switch event.Event.Type {
	case "message_start":
		// Message starting, nothing to render
	case "content_block_start":
		r.CurrentBlockIndex = event.Event.Index
		// Reset block state flags - only one block type active at a time
		r.InTextBlock = false
		r.InToolUseBlock = false
		if len(event.Event.ContentBlock) > 0 {
			var block types.ContentBlock
			if err := json.Unmarshal(event.Event.ContentBlock, &block); err == nil {
				r.CurrentBlockType = block.Type
				if block.Type == "text" {
					r.InTextBlock = true
					r.textBuffer.Reset()
				} else if block.Type == "tool_use" {
					r.InToolUseBlock = true
					r.toolName = block.Name
					r.toolInput.Reset()
				}
			}
		}
	case "content_block_delta":
		if len(event.Event.Delta) > 0 {
			// Try text delta first
			var textDelta TextDelta
			if err := json.Unmarshal(event.Event.Delta, &textDelta); err == nil && textDelta.Type == "text_delta" {
				if showIndicator {
					r.indicator.Show()
				}
				// Buffer text for markdown rendering when block completes
				r.textBuffer.WriteString(textDelta.Text)
				return
			}
			// Try input JSON delta
			var jsonDelta InputJSONDelta
			if err := json.Unmarshal(event.Event.Delta, &jsonDelta); err == nil && jsonDelta.Type == "input_json_delta" {
				r.toolInput.WriteString(jsonDelta.PartialJSON)
			}
		}
	case "content_block_stop":
		if showIndicator {
			r.indicator.Clear()
		}

		if r.InTextBlock && event.Event.Index == r.CurrentBlockIndex {
			// Render buffered text with markdown
			text := r.textBuffer.String()
			if text != "" {
				rendered := r.markdownRenderer.Render(text)
				fmt.Fprint(out, rendered)
			}
		} else if r.InToolUseBlock && event.Event.Index == r.CurrentBlockIndex {
			// Parse and display the accumulated tool input with full header
			var input map[string]interface{}
			toolInputStr := r.toolInput.String()
			if err := json.Unmarshal([]byte(toolInputStr), &input); err == nil {
				out.WriteString(r.toolHeaderRender(r.toolName, input))
			} else {
				// Fallback if JSON parse fails
				fmt.Fprintln(out, style.ApplyThemeBoldGradient(style.Bullet+r.toolName))
			}
		}
	case "message_delta":
		// Contains stop_reason, nothing to render
	case "message_stop":
		// Message complete
		// Note: Don't reset InTextBlock/InToolUseBlock here - they persist
		// until after the assistant event is processed to prevent duplicate rendering
		r.CurrentBlockIndex = -1
	}
}

// Render outputs the stream event to the terminal
func (r *Renderer) Render(event Event) {
	r.renderTo(render.WriterOutput(r.output), event, true)
}

// GetBufferedText returns the accumulated text buffer content
func (r *Renderer) GetBufferedText() string {
	return r.textBuffer.String()
}

// ResetBlockState resets the block tracking state after an assistant event
func (r *Renderer) ResetBlockState() {
	r.InTextBlock = false
	r.InToolUseBlock = false
}

// RenderToString renders the stream event to a string and returns it
func (r *Renderer) RenderToString(event Event) string {
	out := render.StringOutput()
	r.renderTo(out, event, false)
	return out.String()
}
