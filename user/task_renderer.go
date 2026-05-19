package user

import (
	"encoding/json"
	"strings"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/textutil"
)

// TaskListRenderer renders TaskList tool results with visual status indicators.
// The Task* tools replaced TodoWrite in Claude Code 2.1.142; a TaskList result
// carries the full task list, the same role newTodos played for TodoWrite.
type TaskListRenderer struct {
	styleApplier render.StyleApplier
}

// NewTaskListRenderer creates a new TaskListRenderer with the given dependencies.
func NewTaskListRenderer(styleApplier render.StyleApplier) *TaskListRenderer {
	return &TaskListRenderer{styleApplier: styleApplier}
}

// TryRender implements ResultRenderer interface.
// Attempts to render a TaskList result with visual status indicators.
// Returns true if it was a non-empty task list and was rendered.
func (tr *TaskListRenderer) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	if len(toolUseResult) == 0 {
		return false
	}

	var result struct {
		Tasks []state.TaskListItem `json:"tasks"`
	}
	if err := json.Unmarshal(toolUseResult, &result); err != nil {
		return false
	}
	if len(result.Tasks) == 0 {
		return false
	}

	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)
	for _, task := range result.Tasks {
		var statusIndicator string
		var contentRenderer func(string) string

		switch task.Status {
		case "completed":
			statusIndicator = tr.styleApplier.SuccessText("✓")
			contentRenderer = tr.styleApplier.MutedText
		case "in_progress":
			statusIndicator = tr.styleApplier.WarningText("→")
			contentRenderer = func(s string) string { return s }
		default: // "pending"
			statusIndicator = tr.styleApplier.MutedText("○")
			contentRenderer = tr.styleApplier.MutedText
		}

		pw.WriteLinef("%s %s", statusIndicator, contentRenderer(task.Subject))
	}

	return true
}

// TaskCreateRenderer renders a TaskCreate tool result, which carries the id of
// the newly created task.
type TaskCreateRenderer struct {
	styleApplier render.StyleApplier
}

// NewTaskCreateRenderer creates a new TaskCreateRenderer.
func NewTaskCreateRenderer(styleApplier render.StyleApplier) *TaskCreateRenderer {
	return &TaskCreateRenderer{styleApplier: styleApplier}
}

// TryRender implements ResultRenderer. It only handles TaskCreate results;
// TaskGet returns a shape-identical "task" object and must not be matched here.
func (tr *TaskCreateRenderer) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	if ctx.ToolName != "TaskCreate" || len(toolUseResult) == 0 {
		return false
	}

	var result struct {
		Task *struct {
			ID string `json:"id"`
		} `json:"task"`
	}
	if err := json.Unmarshal(toolUseResult, &result); err != nil || result.Task == nil || result.Task.ID == "" {
		return false
	}

	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)
	pw.WriteLine(tr.styleApplier.MutedText("created task #" + result.Task.ID))
	return true
}

// TaskUpdateRenderer renders a TaskUpdate tool result, surfacing the status
// change rather than a bare line count.
type TaskUpdateRenderer struct {
	styleApplier render.StyleApplier
}

// NewTaskUpdateRenderer creates a new TaskUpdateRenderer.
func NewTaskUpdateRenderer(styleApplier render.StyleApplier) *TaskUpdateRenderer {
	return &TaskUpdateRenderer{styleApplier: styleApplier}
}

// TryRender implements ResultRenderer.
func (tr *TaskUpdateRenderer) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	if ctx.ToolName != "TaskUpdate" || len(toolUseResult) == 0 {
		return false
	}

	var result struct {
		StatusChange *struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"statusChange"`
		UpdatedFields []string `json:"updatedFields"`
	}
	if err := json.Unmarshal(toolUseResult, &result); err != nil {
		return false
	}

	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)

	if result.StatusChange != nil && result.StatusChange.To != "" {
		pw.WriteLinef("%s %s %s",
			tr.styleApplier.MutedText(result.StatusChange.From),
			tr.styleApplier.MutedText("→"),
			tr.statusText(result.StatusChange.To))
		return true
	}

	if len(result.UpdatedFields) > 0 {
		pw.WriteLine(tr.styleApplier.MutedText("updated " + strings.Join(result.UpdatedFields, ", ")))
		return true
	}

	return false
}

// statusText styles a task status by its meaning.
func (tr *TaskUpdateRenderer) statusText(status string) string {
	switch status {
	case "completed":
		return tr.styleApplier.SuccessText(status)
	case "in_progress":
		return tr.styleApplier.WarningText(status)
	default:
		return tr.styleApplier.MutedText(status)
	}
}
