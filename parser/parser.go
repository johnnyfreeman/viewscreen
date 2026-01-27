package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/tools"
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
	pendingTools   *tools.ToolUseTracker

	// Renderers for each event type
	systemRenderer    *system.Renderer
	assistantRenderer *assistant.Renderer
	userRenderer      *user.Renderer
	resultRenderer    *result.Renderer
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
		input:             os.Stdin,
		errOutput:         os.Stderr,
		streamRenderer:    stream.NewRenderer(),
		pendingTools:      tools.NewToolUseTracker(),
		systemRenderer:    system.NewRenderer(),
		assistantRenderer: assistant.NewRenderer(),
		userRenderer:      user.NewRenderer(),
		resultRenderer:    result.NewRenderer(),
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
	switch e := event.(type) {
	case events.SystemEvent:
		p.systemRenderer.Render(e.Data)

	case events.AssistantEvent:
		// Buffer tool_use blocks using the events package helper
		events.BufferToolUse(e.Data, p.pendingTools, p.streamRenderer)
		// Render text blocks (tool_use rendering is always suppressed)
		p.assistantRenderer.Render(e.Data, p.streamRenderer.InTextBlock(), true)
		p.streamRenderer.ResetBlockState()

	case events.UserEvent:
		// Match tool results with pending tool_use blocks
		matched := events.MatchToolResults(e.Data, p.pendingTools)

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
			p.userRenderer.SetToolContext(ctx)
		}

		// Render the tool result
		if isNested {
			p.userRenderer.RenderNested(e.Data)
		} else {
			p.userRenderer.Render(e.Data)
		}

	case events.StreamEvent:
		p.streamRenderer.Render(e.Data)

	case events.ResultEvent:
		// Flush any orphaned pending tools
		orphaned := events.FlushOrphanedTools(p.pendingTools)
		for _, o := range orphaned {
			tools.RenderToolUse(o.Block)
			fmt.Println(style.OutputPrefix + style.Muted.Render("(no result)"))
		}
		p.resultRenderer.Render(e.Data)
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
