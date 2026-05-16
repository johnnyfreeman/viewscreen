package user

import (
	"encoding/json"

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
