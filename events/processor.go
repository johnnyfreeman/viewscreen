package events

import (
	"strings"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/user"
)

// ProcessResult contains the output from processing an event.
type ProcessResult struct {
	// Rendered is the rendered content to append to output
	Rendered string
	// HasPendingTools indicates whether there are pending tools waiting for results
	HasPendingTools bool
}

// EventProcessor processes parsed events, updates state, and produces rendered output.
// It separates the event processing logic from the TUI concerns.
type EventProcessor struct {
	renderers *RendererSet
	state     *state.State
}

// NewEventProcessor creates a new EventProcessor with the given state.
func NewEventProcessor(s *state.State) *EventProcessor {
	return &EventProcessor{
		renderers: NewRendererSet(),
		state:     s,
	}
}

// NewEventProcessorWithRenderers creates a new EventProcessor with custom renderers.
// This allows reusing an existing RendererSet, useful for testing or when
// renderers need specific configuration.
func NewEventProcessorWithRenderers(s *state.State, rs *RendererSet) *EventProcessor {
	return &EventProcessor{
		renderers: rs,
		state:     s,
	}
}

// Renderers returns the underlying RendererSet for direct access when needed.
func (p *EventProcessor) Renderers() *RendererSet {
	return p.renderers
}

// SetWidth updates the word-wrap width for all markdown renderers.
// This is called when the viewport resizes.
func (p *EventProcessor) SetWidth(width int) {
	p.renderers.SetWidth(width)
}

// Process handles a parsed event and returns the rendered result.
func (p *EventProcessor) Process(event Event) ProcessResult {
	switch e := event.(type) {
	case SystemEvent:
		return p.processSystem(e.Data)
	case AssistantEvent:
		return p.processAssistant(e.Data)
	case UserEvent:
		return p.processUser(e.Data)
	case StreamEvent:
		return p.processStream(e.Data)
	case ResultEvent:
		return p.processResult(e.Data)
	default:
		return ProcessResult{}
	}
}

// HasPendingTools returns whether there are pending tools awaiting results.
func (p *EventProcessor) HasPendingTools() bool {
	return p.renderers.PendingTools.Len() > 0
}

// RenderPendingTool renders a pending tool with the given icon (for spinner animation).
func (p *EventProcessor) RenderPendingTool(pending tools.PendingTool, icon string) string {
	opts := []tools.HeaderRendererOption{tools.WithIcon(icon)}
	if p.renderers.PendingTools.IsNested(pending) {
		opts = append(opts, tools.WithNested())
	}
	r := tools.NewHeaderRenderer(opts...)
	str, _ := r.RenderBlockToString(pending.Block)
	return str
}

// ForEachPendingTool iterates over all pending tools.
func (p *EventProcessor) ForEachPendingTool(fn func(id string, pending tools.PendingTool)) {
	p.renderers.PendingTools.ForEach(fn)
}

func (p *EventProcessor) processSystem(event system.Event) ProcessResult {
	p.state.UpdateFromSystemEvent(event)
	rendered := p.renderers.System.RenderToString(event)
	return ProcessResult{Rendered: rendered}
}

func (p *EventProcessor) processAssistant(event assistant.Event) ProcessResult {
	p.state.IncrementTurnCount()
	r := p.renderers

	// Buffer tool_use blocks using the tracker's method
	msg := tools.AssistantMessage{
		Content:         event.Message.Content,
		ParentToolUseID: event.ParentToolUseID,
	}
	if r.PendingTools.BufferFromAssistantMessage(msg, r.Stream.InToolUseBlock()) {
		// Update state to show the first pending tool
		for _, block := range event.Message.Content {
			if block.Type == "tool_use" && block.ID != "" {
				p.state.SetCurrentTool(block.Name, tools.GetToolArgFromBlock(block))
				break
			}
		}
	}

	// Render text blocks only (tools are buffered)
	rendered := r.Assistant.RenderToString(
		event,
		r.Stream.InTextBlock(),
		true, // Suppress tool rendering - we handle it separately
	)
	r.Stream.ResetBlockState()

	return ProcessResult{
		Rendered:        rendered,
		HasPendingTools: r.PendingTools.Len() > 0,
	}
}

func (p *EventProcessor) processUser(event user.Event) ProcessResult {
	p.state.UpdateFromToolUseResult(event.ToolUseResult)
	r := p.renderers

	var content strings.Builder
	var isNested bool

	// Check if this is a sub-agent prompt (text content with parent_tool_use_id)
	if event.ParentToolUseID != nil && p.isSubAgentPrompt(event) {
		// Resolve the parent tool early and render its header
		if resolved := r.PendingTools.ResolveParentEarly(*event.ParentToolUseID); resolved != nil {
			isNested = resolved.IsNested
			str, _ := tools.RenderResolved(*resolved)
			content.WriteString(str)
		}
		// Render the prompt text with truncation
		if isNested {
			content.WriteString(r.User.RenderNestedSubAgentPromptToString(event))
		} else {
			content.WriteString(r.User.RenderSubAgentPromptToString(event))
		}
		return ProcessResult{
			Rendered:        content.String(),
			HasPendingTools: r.PendingTools.Len() > 0,
		}
	}

	// Match tool results with pending tools using the tracker's method
	msg := tools.UserMessage{
		Content: make([]tools.UserToolResult, len(event.Message.Content)),
	}
	for i, c := range event.Message.Content {
		msg.Content[i] = tools.UserToolResult{
			Type:      c.Type,
			ToolUseID: c.ToolUseID,
		}
	}
	matched := r.PendingTools.MatchFromUserMessage(msg)

	// Render matched tool headers (unless already rendered) and set context
	for _, match := range matched {
		isNested = match.IsNested
		if !match.HeaderRendered {
			str, ctx := tools.RenderResolved(match.ResolvedTool)
			content.WriteString(str)
			r.User.SetToolContext(ctx)
		} else {
			// Header was rendered early, just set context
			_, ctx := tools.RenderResolved(match.ResolvedTool)
			r.User.SetToolContext(ctx)
		}
	}

	// Clear tool state if no more pending tools
	if r.PendingTools.Len() == 0 {
		p.state.ClearCurrentTool()
	}

	// Render the tool result (with nested prefix if applicable)
	if isNested {
		content.WriteString(r.User.RenderNestedToString(event))
	} else {
		content.WriteString(r.User.RenderToString(event))
	}

	return ProcessResult{
		Rendered:        content.String(),
		HasPendingTools: r.PendingTools.Len() > 0,
	}
}

// isSubAgentPrompt checks if a user event is a sub-agent prompt.
// Sub-agent prompts have text content (not tool_result) and are used to pass
// the prompt to a Task sub-agent.
func (p *EventProcessor) isSubAgentPrompt(event user.Event) bool {
	for _, c := range event.Message.Content {
		if c.Type == "text" {
			return true
		}
	}
	return false
}

func (p *EventProcessor) processStream(event stream.Event) ProcessResult {
	r := p.renderers
	rendered := r.Stream.RenderToString(event)

	// Update state for tool progress tracking
	if event.Event.Type == "content_block_start" && r.Stream.InToolUseBlock() {
		p.state.SetCurrentTool(r.Stream.CurrentBlockType(), "")
	}

	return ProcessResult{Rendered: rendered}
}

func (p *EventProcessor) processResult(event result.Event) ProcessResult {
	r := p.renderers

	var content strings.Builder

	// Flush any orphaned pending tools using the tracker's method
	orphaned := r.PendingTools.FlushAll()
	for _, o := range orphaned {
		str, _ := tools.RenderResolved(o.ResolvedTool)
		content.WriteString(str)
		content.WriteString(style.OutputPrefix + style.MutedText("(no result)") + "\n")
	}

	p.state.ClearCurrentTool()
	p.state.UpdateFromResultEvent(event)
	content.WriteString(r.Result.RenderToString(event))

	return ProcessResult{Rendered: content.String()}
}

