package events

import (
	"encoding/json"
	"strings"

	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/codex"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/timeline"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/user"
)

// ProcessResult contains the output from processing an event.
type ProcessResult struct {
	// Rendered is the rendered content to append to output
	Rendered string
	// Batch is the provider-neutral timeline update produced by the event.
	Batch timeline.Batch
	// HasPendingTools indicates whether there are pending tools waiting for results
	HasPendingTools bool
}

// EventProcessor processes parsed events, updates state, and produces rendered output.
// It separates the event processing logic from the TUI concerns.
type EventProcessor struct {
	renderers *RendererSet
	state     *state.State

	codexActiveTools map[string]codexActiveTool
	codexToolOrder   []string
	activities       map[string]timeline.Activity
	activityOrder    []string
}

type codexActiveTool struct {
	name  string
	input string
}

// NewEventProcessor creates a new EventProcessor with the given state.
func NewEventProcessor(s *state.State) *EventProcessor {
	return &EventProcessor{
		renderers:        NewRendererSet(),
		state:            s,
		codexActiveTools: make(map[string]codexActiveTool),
		activities:       make(map[string]timeline.Activity),
	}
}

// NewEventProcessorWithRenderers creates a new EventProcessor with custom renderers.
// This allows reusing an existing RendererSet, useful for testing or when
// renderers need specific configuration.
func NewEventProcessorWithRenderers(s *state.State, rs *RendererSet) *EventProcessor {
	return &EventProcessor{
		renderers:        rs,
		state:            s,
		codexActiveTools: make(map[string]codexActiveTool),
		activities:       make(map[string]timeline.Activity),
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

func processResultFromRendered(rendered, kind string) ProcessResult {
	return processResultFromBatch(rendered, kind, timeline.StatePatch{})
}

func processResultFromBatch(rendered, kind string, patch timeline.StatePatch) ProcessResult {
	res := ProcessResult{Rendered: rendered, Batch: timeline.Batch{Patch: patch}}
	if rendered != "" {
		res.Batch.Entries = append(res.Batch.Entries, timeline.Entry{Kind: kind, Body: rendered})
	}
	return res
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func systemPatch(event system.Event) timeline.StatePatch {
	patch := timeline.StatePatch{}
	if event.Model != "" {
		patch.Model = timeline.StringPtr(event.Model)
	}
	if event.ClaudeCodeVersion != "" {
		patch.Version = timeline.StringPtr(event.ClaudeCodeVersion)
	}
	if event.CWD != "" {
		patch.CWD = timeline.StringPtr(event.CWD)
	}
	if len(event.Tools) > 0 {
		patch.ToolsCount = timeline.IntPtr(len(event.Tools))
	}
	if len(event.Agents) > 0 {
		patch.Agents = append([]string(nil), event.Agents...)
	}
	if event.PermissionMode != "" {
		patch.PermissionMode = timeline.StringPtr(event.PermissionMode)
	}
	return patch
}

func resultPatch(event result.Event) timeline.StatePatch {
	return timeline.StatePatch{
		ClearActivity: true,
		TurnCount:     timeline.IntPtr(event.NumTurns),
		TotalCost:     timeline.FloatPtr(event.TotalCostUSD),
		IsError:       timeline.BoolPtr(event.IsError),
		DurationMS:    timeline.IntPtr(event.DurationMS),
		DurationAPIMS: timeline.IntPtr(event.DurationAPIMS),
		InputTokens:   timeline.IntPtr(event.Usage.InputTokens),
		OutputTokens:  timeline.IntPtr(event.Usage.OutputTokens),
		CacheCreated:  timeline.IntPtr(event.Usage.CacheCreationInputTokens),
		CacheRead:     timeline.IntPtr(event.Usage.CacheReadInputTokens),
	}
}

func toolUseResultPatch(raw json.RawMessage) timeline.StatePatch {
	if len(raw) == 0 {
		return timeline.StatePatch{}
	}

	patch := timeline.StatePatch{ClearActivity: true}
	var envelope struct {
		NewTodos json.RawMessage `json:"newTodos"`
		Tasks    json.RawMessage `json:"tasks"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return patch
	}
	if envelope.NewTodos != nil {
		var todos []state.Todo
		if err := json.Unmarshal(envelope.NewTodos, &todos); err == nil {
			patch.ReplaceTodos = true
			patch.Todos = make([]timeline.Todo, len(todos))
			for i, todo := range todos {
				patch.Todos[i] = timeline.Todo{Content: todo.Content, Status: todo.Status, ActiveForm: todo.ActiveForm}
			}
		}
		return patch
	}
	if envelope.Tasks != nil {
		var tasks []state.TaskListItem
		if err := json.Unmarshal(envelope.Tasks, &tasks); err == nil {
			patch.ReplaceTodos = true
			patch.Todos = make([]timeline.Todo, len(tasks))
			for i, task := range tasks {
				patch.Todos[i] = timeline.Todo{Content: task.Subject, Status: task.Status}
			}
		}
	}
	return patch
}

// detectAgent records which CLI produced the stream so the TUI can brand
// itself accordingly. Codex envelope events are unambiguous; every other
// concrete event type comes from Claude Code's stream-json. Ignored or empty
// events leave the current value untouched so a seeded agent (prompt mode)
// survives until a definitive event arrives.
func (p *EventProcessor) detectAgent(event Event) {
	switch event.(type) {
	case CodexEvent:
		p.state.ApplyPatch(timeline.StatePatch{Agent: timeline.StringPtr(config.AgentCodex)})
	case SystemEvent, SubAgentSystemEvent, AssistantEvent, UserEvent, StreamEvent, ResultEvent:
		p.state.ApplyPatch(timeline.StatePatch{Agent: timeline.StringPtr(config.AgentClaude)})
	}
}

// HasPendingTools returns whether there are pending tools awaiting results.
func (p *EventProcessor) HasPendingTools() bool {
	return p.renderers.PendingTools.Len() > 0 || len(p.activities) > 0
}

// PendingActivities returns live provider-neutral timeline activities.
func (p *EventProcessor) PendingActivities() []timeline.Activity {
	activities := make([]timeline.Activity, 0, len(p.activities))
	for _, id := range p.activityOrder {
		if activity, ok := p.activities[id]; ok {
			activities = append(activities, activity)
		}
	}
	return activities
}

func (p *EventProcessor) setActivity(activity timeline.Activity) {
	if activity.ID == "" {
		return
	}
	if p.activities == nil {
		p.activities = make(map[string]timeline.Activity)
	}
	if _, exists := p.activities[activity.ID]; !exists {
		p.activityOrder = append(p.activityOrder, activity.ID)
	}
	p.activities[activity.ID] = activity
}

func (p *EventProcessor) removeActivity(id string) {
	if id == "" {
		return
	}
	delete(p.activities, id)
	for i, activeID := range p.activityOrder {
		if activeID == id {
			p.activityOrder = append(p.activityOrder[:i], p.activityOrder[i+1:]...)
			return
		}
	}
}

func (p *EventProcessor) clearActivities() {
	clear(p.activities)
	p.activityOrder = nil
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

	patch := systemPatch(event)
	p.state.ApplyPatch(patch)

	// Skip rendering for system events with no meaningful data.
	// These are typically subagent events that lack parent_tool_use_id.
	if event.Model == "" && event.CWD == "" && len(event.Tools) == 0 {
		return ProcessResult{}
	}

	rendered := p.renderers.System.RenderToString(event)
	return processResultFromBatch(rendered, "system", patch)
}

func (p *EventProcessor) processAssistant(event assistant.Event) ProcessResult {
	patch := timeline.StatePatch{IncrementTurns: 1}

	// Accumulate per-turn token usage for real-time tracking
	if u := event.Message.Usage; u != nil {
		patch.AddUsage = &timeline.Usage{
			InputTokens:  u.InputTokens,
			OutputTokens: u.OutputTokens,
			CacheCreated: u.CacheCreationInputTokens,
			CacheRead:    u.CacheReadInputTokens,
		}
	}
	p.state.ApplyPatch(patch)

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
				activity := timeline.Activity{
					ID:       block.ID,
					ParentID: stringValue(event.ParentToolUseID),
					Name:     block.Name,
					Input:    tools.GetToolArgFromBlock(block),
				}
				p.state.ApplyPatch(timeline.StatePatch{CurrentActivity: &activity})
				patch.CurrentActivity = &activity
				p.setActivity(activity)
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

	res := processResultFromBatch(rendered, "assistant", patch)
	res.HasPendingTools = r.PendingTools.Len() > 0
	return res
}

func (p *EventProcessor) processUser(event user.Event) ProcessResult {
	patch := toolUseResultPatch(event.ToolUseResult)
	p.state.ApplyPatch(patch)
	r := p.renderers

	var content strings.Builder
	var isNested bool

	// Check if this is a sub-agent prompt (text content with parent_tool_use_id)
	if event.ParentToolUseID != nil && p.isSubAgentPrompt(event) {
		// Resolve the parent tool early and render its header
		if resolved := r.PendingTools.ResolveParentEarly(*event.ParentToolUseID); resolved != nil {
			isNested = resolved.IsNested
			if activity, ok := p.activities[*event.ParentToolUseID]; ok {
				activity.HeaderRendered = true
				p.setActivity(activity)
			}
			str, _ := tools.RenderResolved(*resolved)
			content.WriteString(str)
		}
		// Render the prompt text with truncation
		if isNested {
			content.WriteString(r.User.RenderNestedSubAgentPromptToString(event))
		} else {
			content.WriteString(r.User.RenderSubAgentPromptToString(event))
		}
		res := processResultFromBatch(content.String(), "user", patch)
		res.HasPendingTools = r.PendingTools.Len() > 0
		return res
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
		if c.Type == "tool_result" {
			p.removeActivity(c.ToolUseID)
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
		patch.ClearActivity = true
		p.state.ApplyPatch(timeline.StatePatch{ClearActivity: true})
	}

	// Render the tool result (with nested prefix if applicable)
	if isNested {
		content.WriteString(r.User.RenderNestedToString(event))
	} else {
		content.WriteString(r.User.RenderToString(event))
	}

	res := processResultFromBatch(content.String(), "user", patch)
	res.HasPendingTools = r.PendingTools.Len() > 0
	return res
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
		activity := timeline.Activity{Name: toolName}
		patch := timeline.StatePatch{CurrentActivity: &activity}
		p.state.ApplyPatch(patch)
		return processResultFromBatch(rendered, "stream", patch)
	}

	return processResultFromRendered(rendered, "stream")
}

func (p *EventProcessor) processResult(event result.Event) ProcessResult {
	r := p.renderers

	var content strings.Builder

	// Flush any orphaned pending tools using the tracker's method
	orphaned := r.PendingTools.FlushAll()
	for _, o := range orphaned {
		p.removeActivity(o.ID)
		str, _ := tools.RenderResolved(o.ResolvedTool)
		content.WriteString(str)
		content.WriteString(style.OutputPrefix + style.MutedText("(no result)") + "\n")
	}
	p.clearActivities()

	patch := resultPatch(event)
	p.state.ApplyPatch(patch)
	content.WriteString(r.Result.RenderToString(event))

	return processResultFromBatch(content.String(), "result", patch)
}

// processCodex handles an event from the Codex CLI stream. Codex events are
// rendered by a dedicated codex.Renderer; this method also folds the event's
// effect into the shared state (token usage, the live "Running" spinner, and
// the sidebar task list) so the TUI reflects a codex stream the same way it
// reflects a Claude one.
func (p *EventProcessor) processCodex(event codex.Event) ProcessResult {
	p.applyCodexState(event)
	res := processResultFromBatch(p.renderers.Codex.Render(event), "codex", codexPatch(event))
	res.HasPendingTools = len(p.activities) > 0
	return res
}

func codexPatch(event codex.Event) timeline.StatePatch {
	switch event.Type {
	case codex.TypeTurnStarted:
		return timeline.StatePatch{IncrementTurns: 1}
	case codex.TypeTurnCompleted:
		patch := timeline.StatePatch{ClearActivity: true}
		if event.Usage != nil {
			patch.AddUsage = &timeline.Usage{
				InputTokens:     event.Usage.InputTokens,
				OutputTokens:    event.Usage.OutputTokens,
				CacheRead:       event.Usage.CachedInputTokens,
				ReasoningTokens: event.Usage.ReasoningOutputTokens,
			}
		}
		return patch
	case codex.TypeTurnFailed:
		return timeline.StatePatch{ClearActivity: true}
	case codex.TypeItemStarted, codex.TypeItemUpdated, codex.TypeItemCompleted:
		if event.Item == nil {
			return timeline.StatePatch{}
		}
		completed := event.Type == codex.TypeItemCompleted
		switch event.Item.Type {
		case codex.ItemTodoList:
			return timeline.StatePatch{ReplaceTodos: true, Todos: codexTodos(event.Item.Items)}
		case codex.ItemCommandExecution:
			return codexActivityPatch(completed, "Shell", codex.ShellCommand(event.Item.Command))
		case codex.ItemMCPToolCall:
			return codexActivityPatch(completed, codex.MCPLabel(event.Item), "")
		case codex.ItemFileChange:
			return codexActivityPatch(completed, "Edit", codex.FileChangeSummary(event.Item.Changes))
		case codex.ItemWebSearch:
			return codexActivityPatch(completed, "Web Search", event.Item.Query)
		}
	}
	return timeline.StatePatch{}
}

func codexActivityPatch(completed bool, name, input string) timeline.StatePatch {
	if completed {
		return timeline.StatePatch{ClearActivity: true}
	}
	activity := timeline.Activity{Name: name, Input: input}
	return timeline.StatePatch{CurrentActivity: &activity}
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
		p.state.ApplyPatch(timeline.StatePatch{IncrementTurns: 1})
	case codex.TypeTurnCompleted:
		if event.Usage != nil {
			u := event.Usage
			p.state.ApplyPatch(timeline.StatePatch{AddUsage: &timeline.Usage{
				InputTokens:     u.InputTokens,
				OutputTokens:    u.OutputTokens,
				CacheRead:       u.CachedInputTokens,
				ReasoningTokens: u.ReasoningOutputTokens,
			}})
		}
		p.clearCodexActiveTools()
	case codex.TypeTurnFailed:
		p.clearCodexActiveTools()
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
		p.state.ApplyPatch(timeline.StatePatch{ReplaceTodos: true, Todos: codexTodos(item.Items)})
	case codex.ItemCommandExecution:
		p.updateCodexActiveTool(completed, item.ID, "Shell", codex.ShellCommand(item.Command))
	case codex.ItemMCPToolCall:
		p.updateCodexActiveTool(completed, item.ID, codex.MCPLabel(item), "")
	case codex.ItemFileChange:
		p.updateCodexActiveTool(completed, item.ID, "Edit", codex.FileChangeSummary(item.Changes))
	case codex.ItemWebSearch:
		p.updateCodexActiveTool(completed, item.ID, "Web Search", item.Query)
	}
}

// updateCodexActiveTool shows the newest in-flight codex tool in the live
// spinner while preserving older overlapping items. Codex can start multiple
// command_execution items before completing them; when the visible item
// completes, the spinner falls back to the next still-active item instead of
// going blank.
func (p *EventProcessor) updateCodexActiveTool(completed bool, id, name, input string) {
	if id == "" {
		if completed {
			p.state.ApplyPatch(timeline.StatePatch{ClearActivity: true})
		} else {
			activity := timeline.Activity{Name: name, Input: input}
			p.state.ApplyPatch(timeline.StatePatch{CurrentActivity: &activity})
		}
		return
	}

	if completed {
		delete(p.codexActiveTools, id)
		p.removeCodexToolOrder(id)
		p.removeActivity(id)
		p.restoreLatestCodexTool()
		return
	}

	if p.codexActiveTools == nil {
		p.codexActiveTools = make(map[string]codexActiveTool)
	}
	if _, exists := p.codexActiveTools[id]; !exists {
		p.codexToolOrder = append(p.codexToolOrder, id)
	}
	p.codexActiveTools[id] = codexActiveTool{name: name, input: input}
	activity := timeline.Activity{ID: id, Name: name, Input: input}
	p.setActivity(activity)
	p.state.ApplyPatch(timeline.StatePatch{CurrentActivity: &activity})
}

func (p *EventProcessor) restoreLatestCodexTool() {
	for i := len(p.codexToolOrder) - 1; i >= 0; i-- {
		id := p.codexToolOrder[i]
		if tool, ok := p.codexActiveTools[id]; ok {
			activity := timeline.Activity{ID: id, Name: tool.name, Input: tool.input}
			p.state.ApplyPatch(timeline.StatePatch{CurrentActivity: &activity})
			return
		}
	}
	p.state.ApplyPatch(timeline.StatePatch{ClearActivity: true})
}

func (p *EventProcessor) clearCodexActiveTools() {
	clear(p.codexActiveTools)
	p.codexToolOrder = nil
	p.clearActivities()
	p.state.ApplyPatch(timeline.StatePatch{ClearActivity: true})
}

func (p *EventProcessor) removeCodexToolOrder(id string) {
	for i, activeID := range p.codexToolOrder {
		if activeID == id {
			p.codexToolOrder = append(p.codexToolOrder[:i], p.codexToolOrder[i+1:]...)
			return
		}
	}
}

// codexTodos maps codex todo_list items onto the shared todo model. Codex only
// reports a boolean completion (no explicit in-progress marker), so each item is
// either completed or pending.
func codexTodos(items []codex.TodoItem) []timeline.Todo {
	todos := make([]timeline.Todo, len(items))
	for i, it := range items {
		status := "pending"
		if it.Completed {
			status = "completed"
		}
		todos[i] = timeline.Todo{Content: it.Text, Status: status}
	}
	return todos
}
