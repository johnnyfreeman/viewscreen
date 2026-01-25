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

// EventHandler is called for each parsed event
type EventHandler func(eventType string, line []byte) error

// Parser handles reading and dispatching events from an input source
type Parser struct {
	input          io.Reader
	errOutput      io.Writer
	streamRenderer *stream.Renderer
	eventHandler   EventHandler
}

// Option configures a Parser
type Option func(*Parser)

// WithInput sets a custom input reader (default: os.Stdin)
func WithInput(r io.Reader) Option {
	return func(p *Parser) {
		p.input = r
	}
}

// WithErrOutput sets a custom error output writer (default: os.Stderr)
func WithErrOutput(w io.Writer) Option {
	return func(p *Parser) {
		p.errOutput = w
	}
}

// WithStreamRenderer sets a custom stream renderer
func WithStreamRenderer(r *stream.Renderer) Option {
	return func(p *Parser) {
		p.streamRenderer = r
	}
}

// WithEventHandler sets a custom event handler for testing
func WithEventHandler(h EventHandler) Option {
	return func(p *Parser) {
		p.eventHandler = h
	}
}

// NewParser creates a new Parser with default options
func NewParser() *Parser {
	return NewParserWithOptions()
}

// NewParserWithOptions creates a new Parser with custom options
func NewParserWithOptions(opts ...Option) *Parser {
	p := &Parser{
		input:          os.Stdin,
		errOutput:      os.Stderr,
		streamRenderer: stream.NewRenderer(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Run reads events from input and renders them
func (p *Parser) Run() error {
	scanner := bufio.NewScanner(p.input)

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
			fmt.Fprintf(p.errOutput, "Error parsing JSON: %v\n", err)
			continue
		}

		// Call event handler if set (for testing)
		if p.eventHandler != nil {
			if err := p.eventHandler(base.Type, []byte(line)); err != nil {
				return err
			}
		}

		switch base.Type {
		case "system":
			var event system.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(p.errOutput, "Error parsing system event: %v\n", err)
				continue
			}
			system.Render(event)

		case "assistant":
			var event assistant.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(p.errOutput, "Error parsing assistant event: %v\n", err)
				continue
			}
			assistant.Render(event, p.streamRenderer.InTextBlock, p.streamRenderer.InToolUseBlock)
			p.streamRenderer.ResetBlockState()

		case "user":
			var event user.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(p.errOutput, "Error parsing user event: %v\n", err)
				continue
			}
			user.Render(event)

		case "stream_event":
			var event stream.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(p.errOutput, "Error parsing stream event: %v\n", err)
				continue
			}
			p.streamRenderer.Render(event)

		case "result":
			var event result.Event
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				fmt.Fprintf(p.errOutput, "Error parsing result event: %v\n", err)
				continue
			}
			result.Render(event)

		default:
			fmt.Fprintf(p.errOutput, "Unknown event type: %s\n", base.Type)
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	return nil
}
