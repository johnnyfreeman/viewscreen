package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/state"
)

// EventHandler is called for each parsed event
type EventHandler func(eventType string, line []byte) error

// Parser handles reading and dispatching events from an input source.
// It uses EventProcessor internally for event processing, writing output
// directly to stdout for streaming display.
type Parser struct {
	input        io.Reader
	output       io.Writer
	errOutput    io.Writer
	eventHandler EventHandler
	processor    *events.EventProcessor
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

// WithEventHandler sets a custom event handler for testing
func WithEventHandler(h EventHandler) Option {
	return func(p *Parser) {
		p.eventHandler = h
	}
}

// WithRendererSet sets a custom renderer set for the internal EventProcessor.
func WithRendererSet(rs *events.RendererSet) Option {
	return func(p *Parser) {
		// Create a new processor with the custom renderers
		p.processor = events.NewEventProcessorWithRenderers(state.NewState(), rs)
	}
}

// WithOutput sets a custom output writer (default: os.Stdout)
func WithOutput(w io.Writer) Option {
	return func(p *Parser) {
		p.output = w
	}
}

// NewParser creates a new Parser with default options
func NewParser() *Parser {
	return NewParserWithOptions()
}

// NewParserWithOptions creates a new Parser with custom options
func NewParserWithOptions(opts ...Option) *Parser {
	p := &Parser{
		input:     os.Stdin,
		output:    os.Stdout,
		errOutput: os.Stderr,
		processor: events.NewEventProcessor(state.NewState()),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Renderers returns the underlying RendererSet for tests that need to inspect state.
func (p *Parser) Renderers() *events.RendererSet {
	return p.processor.Renderers()
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

		// Parse the event using the events package
		parsed := events.Parse(line)
		if parsed == nil {
			continue
		}

		// Handle parse errors
		if parseErr, ok := parsed.(events.ParseError); ok {
			if parseErr.Err != nil {
				fmt.Fprintf(p.errOutput, "Error parsing JSON: %v\n", parseErr.Err)
			} else {
				fmt.Fprintf(p.errOutput, "%s\n", parseErr.Line)
			}
			continue
		}

		// Call event handler if set (for testing)
		if p.eventHandler != nil {
			eventType := eventTypeName(parsed)
			if err := p.eventHandler(eventType, []byte(line)); err != nil {
				return err
			}
		}

		// Process the event through EventProcessor and write output
		result := p.processor.Process(parsed)
		if result.Rendered != "" {
			fmt.Fprint(p.output, result.Rendered)
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	return nil
}

// eventTypeName returns the event type name for the handler callback
func eventTypeName(event events.Event) string {
	switch event.(type) {
	case events.SystemEvent:
		return "system"
	case events.AssistantEvent:
		return "assistant"
	case events.UserEvent:
		return "user"
	case events.StreamEvent:
		return "stream_event"
	case events.ResultEvent:
		return "result"
	default:
		return "unknown"
	}
}
