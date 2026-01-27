package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/tools"
)

// EventHandler is called for each parsed event
type EventHandler func(eventType string, line []byte) error

// Parser handles reading and dispatching events from an input source
type Parser struct {
	input        io.Reader
	errOutput    io.Writer
	eventHandler EventHandler
	renderers    *events.RendererSet
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

// WithRendererSet sets a custom renderer set
func WithRendererSet(rs *events.RendererSet) Option {
	return func(p *Parser) {
		p.renderers = rs
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
		errOutput: os.Stderr,
		renderers: events.NewRendererSet(),
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

		// Process the event
		if err := p.processEvent(parsed); err != nil {
			fmt.Fprintf(p.errOutput, "Error processing event: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	return nil
}

// processEvent handles a parsed event
func (p *Parser) processEvent(event events.Event) error {
	r := p.renderers

	switch e := event.(type) {
	case events.SystemEvent:
		r.System.Render(e.Data)

	case events.AssistantEvent:
		// Buffer tool_use blocks using the events package helper
		events.BufferToolUse(e.Data, r.PendingTools, r.Stream)
		// Render text blocks (tool_use rendering is always suppressed)
		r.Assistant.Render(e.Data, r.Stream.InTextBlock(), true)
		r.Stream.ResetBlockState()

	case events.UserEvent:
		// Match tool results with pending tool_use blocks
		matched := events.MatchToolResults(e.Data, r.PendingTools)

		// Render matched tool headers and set context
		var isNested bool
		for _, m := range matched {
			isNested = m.IsNested
			var ctx tools.ToolContext
			if m.IsNested {
				ctx = tools.RenderNestedToolUse(m.Block)
			} else {
				ctx = tools.RenderToolUse(m.Block)
			}
			r.User.SetToolContext(ctx)
		}

		// Render the tool result
		if isNested {
			r.User.RenderNested(e.Data)
		} else {
			r.User.Render(e.Data)
		}

	case events.StreamEvent:
		r.Stream.Render(e.Data)

	case events.ResultEvent:
		// Flush any orphaned pending tools
		orphaned := events.FlushOrphanedTools(r.PendingTools)
		for _, o := range orphaned {
			tools.RenderToolUse(o.Block)
			fmt.Println(style.OutputPrefix + style.Muted.Render("(no result)"))
		}
		r.Result.Render(e.Data)
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
