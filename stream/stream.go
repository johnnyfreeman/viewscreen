package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/indicator"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/terminal"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
)

// MarkdownRenderer is an alias for types.MarkdownRenderer for backward compatibility.
type MarkdownRenderer = types.MarkdownRenderer

// IndicatorInterface abstracts the streaming indicator for testability
type IndicatorInterface interface {
	Show()
	Clear()
}

// ToolHeaderRenderer abstracts tool header rendering for testability.
// It returns the rendered string (and optional tool context) instead of printing directly.
type ToolHeaderRenderer func(toolName string, input map[string]any) (string, tools.ToolContext)

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
	block            *BlockState
	markdownRenderer MarkdownRenderer
	indicator        IndicatorInterface
	toolHeaderRender ToolHeaderRenderer
	output           io.Writer
	width            int
}

// defaultToolHeaderRenderer adapts HeaderRenderer to the ToolHeaderRenderer interface
func defaultToolHeaderRenderer(toolName string, input map[string]any) (string, tools.ToolContext) {
	return tools.NewHeaderRenderer().RenderToString(toolName, input)
}

// NewRenderer creates a new stream Renderer
func NewRenderer() *Renderer {
	width := terminal.Width()

	return &Renderer{
		block:            NewBlockState(),
		markdownRenderer: render.NewMarkdownRenderer(config.NoColor, width),
		indicator:        indicator.NewStreamingIndicator(config.NoColor),
		toolHeaderRender: defaultToolHeaderRenderer,
		output:           os.Stdout,
		width:            width,
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
func WithMarkdownRenderer(mr MarkdownRenderer) RendererOption {
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
		r.block.StartBlock(event.Event.Index, event.Event.ContentBlock)

	case "content_block_delta":
		if len(event.Event.Delta) > 0 {
			// Try text delta first
			var textDelta TextDelta
			if err := json.Unmarshal(event.Event.Delta, &textDelta); err == nil && textDelta.Type == "text_delta" {
				if showIndicator {
					r.indicator.Show()
				}
				r.block.AccumulateText(textDelta.Text)
				return
			}
			// Try input JSON delta
			var jsonDelta InputJSONDelta
			if err := json.Unmarshal(event.Event.Delta, &jsonDelta); err == nil && jsonDelta.Type == "input_json_delta" {
				r.block.AccumulateToolInput(jsonDelta.PartialJSON)
			}
		}

	case "content_block_stop":
		if showIndicator {
			r.indicator.Clear()
		}

		stoppedType := r.block.StopBlock(event.Event.Index)
		switch stoppedType {
		case BlockText:
			text := r.block.TextContent()
			if text != "" {
				rendered := r.markdownRenderer.Render(text)
				fmt.Fprint(out, rendered)
			}
		case BlockToolUse:
			if input, ok := r.block.ParseToolInput(); ok {
				str, _ := r.toolHeaderRender(r.block.ToolName(), input)
				out.WriteString(str)
			} else {
				// Fallback if JSON parse fails
				fmt.Fprintln(out, style.ApplyThemeBoldGradient(style.Bullet+r.block.ToolName()))
			}
		}

	case "message_delta":
		// Contains stop_reason, nothing to render

	case "message_stop":
		// Message complete
		// Note: Don't reset block type here - it persists until after
		// the assistant event is processed to prevent duplicate rendering
		r.block.ResetMessage()
	}
}

// Render outputs the stream event to the terminal
func (r *Renderer) Render(event Event) {
	r.renderTo(render.WriterOutput(r.output), event, true)
}

// GetBufferedText returns the accumulated text buffer content
func (r *Renderer) GetBufferedText() string {
	return r.block.TextContent()
}

// ResetBlockState resets the block tracking state after an assistant event
func (r *Renderer) ResetBlockState() {
	r.block.Reset()
}

// RenderToString renders the stream event to a string and returns it
func (r *Renderer) RenderToString(event Event) string {
	out := render.StringOutput()
	r.renderTo(out, event, false)
	return out.String()
}

// InTextBlock returns true if currently processing a text block.
// This is used by external code (TUI, parser) to check streaming state.
func (r *Renderer) InTextBlock() bool {
	return r.block.InTextBlock()
}

// InToolUseBlock returns true if currently processing a tool_use block.
// This is used by external code (TUI, parser) to check streaming state.
func (r *Renderer) InToolUseBlock() bool {
	return r.block.InToolUseBlock()
}

// CurrentBlockType returns the current block type as a string.
// This is used by external code to query the block type.
func (r *Renderer) CurrentBlockType() string {
	return r.block.Type().String()
}
