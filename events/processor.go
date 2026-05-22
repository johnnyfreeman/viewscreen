package events

import (
	"strings"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/codex"
	"github.com/johnnyfreeman/viewscreen/config"
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
	p.detectAgent(event)
	switch e := event.(type) {
	case SystemEvent:
		return p.processSystem(e.Data)
	case SubAgentSystemEvent:
		return ProcessResult{}
	case AssistantEvent:
		return p.processAssistant(e.Data)
	case UserEvent:
		return p.processUser(e.Data)
	case StreamEvent:
		return p.processStream(e.Data)
	case ResultEvent:
		return p.processResult(e.Data)
	case CodexEvent:
		return p.processCodex(e.Data)
	case IgnoredEvent:
		return ProcessResult{}
	default:
		return ProcessResult{}
	}
}

// detectAgent records which CLI produced the stream so the TUI can brand
// itself accordingly. Codex envelope events are unambiguous; every other
// concrete event type comes from Claude Code's stream-json. Ignored or empty
// events leave the current value untouched so a seeded agent (prompt mode)
// survives until a definitive event arrives.
func (p *EventProcessor) detectAgent(event Event) {
	switch event.(type) {
	case CodexEvent:
		p.state.Agent = config.AgentCodex
	case SystemEvent, SubAgentSystemEvent, AssistantEvent, UserEvent, StreamEvent, ResultEvent:
		p.state.Agent = config.AgentClaude
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
	// Claude transcript logs include many non-init system metadata events
	// (e.g. turn_duration, local_command). Ignore these to avoid rendering
	// repeated empty "Session Started" blocks.
	if event.Subtype != "" && event.Subtype != "init" {
		return ProcessResult{}
	}

	p.state.UpdateFromSystemEvent(event)

	// Skip rendering for system events with no meaningful data.
	// These are typically subagent events that lack parent_tool_use_id.
	if event.Model == "" && event.CWD == "" && len(event.Tools) == 0 {
		return ProcessResult{}
	}

	rendered := p.renderers.System.RenderToString(event)
	return ProcessResult{Rendered: rendered}
}

func (p *EventProcessor) processAssistant(event assistant.Event) ProcessResult {
	p.state.IncrementTurnCount()

	// Accumulate per-turn token usage for real-time tracking
	if u := event.Message.Usage; u != nil {
		p.state.AccumulateUsage(u.InputTokens, u.OutputTokens, u.CacheCreationInputTokens, u.CacheReadInputTokens)
	}

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
		toolName := r.Stream.CurrentToolName()
		if toolName == "" {
			toolName = r.Stream.CurrentBlockType()
		}
		p.state.SetCurrentTool(toolName, "")
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

// processCodex handles an event from the Codex CLI stream. Codex events are
// rendered by a dedicated codex.Renderer; this method also folds the event's
// effect into the shared state (token usage, the live "Running" spinner, and
// the sidebar task list) so the TUI reflects a codex stream the same way it
// reflects a Claude one.
func (p *EventProcessor) processCodex(event codex.Event) ProcessResult {
	p.applyCodexState(event)
	return ProcessResult{Rendered: p.renderers.Codex.Render(event)}
}

// applyCodexState updates the shared TUI state from a codex event. Codex runs
// one item at a time, so an item.started for a long-running item (a shell
// command, MCP call, file change, or web search) sets the current tool for the
// spinner, and the matching item.completed — or the end of the turn — clears it.
// turn.started bumps the turn counter so the sidebar reflects work in progress
// the way it does for Claude (whose assistant messages drive the count).
func (p *EventProcessor) applyCodexState(event codex.Event) {
	switch event.Type {
	case codex.TypeTurnStarted:
		p.state.IncrementTurnCount()
	case codex.TypeTurnCompleted:
		if event.Usage != nil {
			u := event.Usage
			p.state.AccumulateUsage(u.InputTokens, u.OutputTokens, 0, u.CachedInputTokens)
			p.state.ReasoningTokens += u.ReasoningOutputTokens
		}
		p.state.ClearCurrentTool()
	case codex.TypeTurnFailed:
		p.state.ClearCurrentTool()
	case codex.TypeItemStarted, codex.TypeItemUpdated, codex.TypeItemCompleted:
		p.applyCodexItem(event.Type, event.Item)
	}
}

// applyCodexItem applies a single codex work item to the spinner and task list.
// The spinner label matches the item's inline header so the live view and the
// scrollback agree on what codex is doing.
func (p *EventProcessor) applyCodexItem(phase string, item *codex.Item) {
	if item == nil {
		return
	}
	completed := phase == codex.TypeItemCompleted
	switch item.Type {
	case codex.ItemTodoList:
		// Replace the sidebar task list on every update so it tracks the
		// latest completion state, even though the inline render dedupes by id.
		p.state.Todos = codexTodos(item.Items)
	case codex.ItemCommandExecution:
		p.setOrClearCurrentTool(completed, "Shell", codex.ShellCommand(item.Command))
	case codex.ItemMCPToolCall:
		p.setOrClearCurrentTool(completed, codex.MCPLabel(item), "")
	case codex.ItemFileChange:
		p.setOrClearCurrentTool(completed, "Edit", codex.FileChangeSummary(item.Changes))
	case codex.ItemWebSearch:
		p.setOrClearCurrentTool(completed, "Web Search", item.Query)
	}
}

// setOrClearCurrentTool shows the running tool while an item is in flight and
// clears it once the item completes.
func (p *EventProcessor) setOrClearCurrentTool(completed bool, name, input string) {
	if completed {
		p.state.ClearCurrentTool()
	} else {
		p.state.SetCurrentTool(name, input)
	}
}

// codexTodos maps codex todo_list items onto the shared todo model. Codex only
// reports a boolean completion (no explicit in-progress marker), so each item is
// either completed or pending.
func codexTodos(items []codex.TodoItem) []state.Todo {
	todos := make([]state.Todo, len(items))
	for i, it := range items {
		status := "pending"
		if it.Completed {
			status = "completed"
		}
		todos[i] = state.Todo{Content: it.Text, Status: status}
	}
	return todos
}
