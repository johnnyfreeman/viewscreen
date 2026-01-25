package assistant

import (
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

var markdownRenderer *render.MarkdownRenderer

func getMarkdownRenderer() *render.MarkdownRenderer {
	if markdownRenderer == nil {
		width := 80
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			width = w
		}
		markdownRenderer = render.NewMarkdownRenderer(config.NoColor, width)
	}
	return markdownRenderer
}

// Render outputs the assistant event to the terminal
// inTextBlock and inToolUseBlock indicate whether we were streaming these block types
func Render(event Event, inTextBlock, inToolUseBlock bool) {
	if event.Error != "" {
		fmt.Println(style.Error.Bold(true).Render(fmt.Sprintf("%sError", style.Bullet)))
		fmt.Printf("%s%s\n", style.OutputPrefix, style.Error.Render(event.Error))
	}

	for _, block := range event.Message.Content {
		switch block.Type {
		case "text":
			// Only render if we weren't streaming (text would already be shown)
			if !inTextBlock {
				// Use markdown renderer for non-streamed text
				rendered := getMarkdownRenderer().Render(block.Text)
				fmt.Print(rendered)
				if !strings.HasSuffix(rendered, "\n") {
					fmt.Println()
				}
			}
		case "tool_use":
			// Only render if we weren't streaming
			if !inToolUseBlock {
				tools.RenderToolUse(block)
			}
		}
	}
}
