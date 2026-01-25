package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

// Parser handles reading and dispatching events from stdin
type Parser struct {
	streamRenderer *stream.Renderer
}

// NewParser creates a new Parser
func NewParser() *Parser {
	return &Parser{
		streamRenderer: stream.NewRenderer(),
	}
}

// Run reads events from stdin and renders them
func (p *Parser) Run() error {
	scanner := bufio.NewScanner(os.Stdin)

	// Increase buffer size for large JSON lines
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse base event to determine type
		var base types.BaseEvent
		if err := json.Unmarshal([]byte(line), &base); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			continue
		}

		switch base.Type {
		case "system":
			var event system.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing system event: %v\n", err)
				continue
			}
			system.Render(event)

		case "assistant":
			var event assistant.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing assistant event: %v\n", err)
				continue
			}
			assistant.Render(event, p.streamRenderer.InTextBlock, p.streamRenderer.InToolUseBlock)
			p.streamRenderer.ResetBlockState()

		case "user":
			var event user.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing user event: %v\n", err)
				continue
			}
			user.Render(event)

		case "stream_event":
			var event stream.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing stream event: %v\n", err)
				continue
			}
			p.streamRenderer.Render(event)

		case "result":
			var event result.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing result event: %v\n", err)
				continue
			}
			result.Render(event)

		default:
			fmt.Fprintf(os.Stderr, "Unknown event type: %s\n", base.Type)
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	return nil
}
