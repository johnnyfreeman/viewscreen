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
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/tools"
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
	pendingTools   *tools.ToolUseTracker
	dispatcher     *EventDispatcher
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
		pendingTools:   tools.NewToolUseTracker(),
	}
	for _, opt := range opts {
		opt(p)
	}
	p.dispatcher = p.buildDispatcher()
	return p
}


// buildDispatcher creates and configures the event dispatcher with all handlers.
func (p *Parser) buildDispatcher() *EventDispatcher {
	d := NewEventDispatcher(p.errOutput)
	d.Register("system", p.handleSystem)
	d.Register("assistant", p.handleAssistant)
	d.Register("user", p.handleUser)
	d.Register("stream_event", p.handleStream)
	d.Register("result", p.handleResult)
	return d
}

func (p *Parser) handleSystem(line []byte) error {
	var event system.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return err
	}
	system.Render(event)
	return nil
}

func (p *Parser) handleAssistant(line []byte) error {
	var event assistant.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return err
	}
	// Buffer tool_use blocks instead of rendering immediately
	// They will be rendered together with their results
	for _, block := range event.Message.Content {
		if block.Type == "tool_use" && block.ID != "" {
			if !p.streamRenderer.InToolUseBlock {
				p.pendingTools.Add(block.ID, block, event.ParentToolUseID)
			}
		}
	}
	// Render text blocks (tool_use rendering is always suppressed)
	assistant.Render(event, p.streamRenderer.InTextBlock, true)
	p.streamRenderer.ResetBlockState()
	return nil
}

func (p *Parser) handleUser(line []byte) error {
	var event user.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return err
	}
	// Match tool results with pending tool_use blocks
	var isNested bool
	for _, content := range event.Message.Content {
		if content.Type == "tool_result" && content.ToolUseID != "" {
			if pending, ok := p.pendingTools.Get(content.ToolUseID); ok {
				// Check if this is a nested tool (parent is still pending)
				isNested = p.pendingTools.IsNested(pending)
				if isNested {
					tools.RenderNestedToolUse(pending.Block)
				} else {
					tools.RenderToolUse(pending.Block)
				}
				p.pendingTools.Remove(content.ToolUseID)
			}
		}
	}
	// Render the tool result (with nested prefix if applicable)
	if isNested {
		user.RenderNested(event)
	} else {
		user.Render(event)
	}
	return nil
}

func (p *Parser) handleStream(line []byte) error {
	var event stream.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return err
	}
	p.streamRenderer.Render(event)
	return nil
}

func (p *Parser) handleResult(line []byte) error {
	// Flush any orphaned pending tools before rendering result
	p.pendingTools.ForEach(func(id string, pending tools.PendingTool) {
		tools.RenderToolUse(pending.Block)
		fmt.Println(style.OutputPrefix + style.Muted.Render("(no result)"))
	})
	p.pendingTools.Clear()
	var event result.Event
	if err := json.Unmarshal(line, &event); err != nil {
		return err
	}
	result.Render(event)
	return nil
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

		// Dispatch to the appropriate handler
		if !p.dispatcher.Dispatch(base.Type, []byte(line)) {
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
