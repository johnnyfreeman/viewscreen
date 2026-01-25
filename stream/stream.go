package stream

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jfreeman/viewscreen/config"
	"github.com/jfreeman/viewscreen/render"
	"github.com/jfreeman/viewscreen/style"
	"github.com/jfreeman/viewscreen/tools"
	"github.com/jfreeman/viewscreen/types"
	"golang.org/x/term"
)

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
	markdownRenderer  *render.MarkdownRenderer
	indicator         *render.StreamingIndicator
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
		width:             width,
	}
}

// Render outputs the stream event to the terminal
func (r *Renderer) Render(event Event) {
	switch event.Event.Type {
	case "message_start":
		// Message starting, nothing to render
	case "content_block_start":
		r.CurrentBlockIndex = event.Event.Index
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
				// Show streaming indicator on first delta
				r.indicator.Show()
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
		// Clear streaming indicator before rendering final content
		r.indicator.Clear()

		if r.InTextBlock && event.Event.Index == r.CurrentBlockIndex {
			// Render buffered text with markdown
			text := r.textBuffer.String()
			if text != "" {
				rendered := r.markdownRenderer.Render(text)
				fmt.Print(rendered)
			}
		} else if r.InToolUseBlock && event.Event.Index == r.CurrentBlockIndex {
			// Parse and display the accumulated tool input with full header
			var input map[string]interface{}
			toolInputStr := r.toolInput.String()
			if err := json.Unmarshal([]byte(toolInputStr), &input); err == nil {
				tools.RenderToolHeader(r.toolName, input)
			} else {
				// Fallback if JSON parse fails
				fmt.Println(style.ToolHeader.Render(fmt.Sprintf("%s%s()", style.Bullet, r.toolName)))
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

// GetBufferedText returns the accumulated text buffer content
func (r *Renderer) GetBufferedText() string {
	return r.textBuffer.String()
}

// ResetBlockState resets the block tracking state after an assistant event
func (r *Renderer) ResetBlockState() {
	r.InTextBlock = false
	r.InToolUseBlock = false
}
