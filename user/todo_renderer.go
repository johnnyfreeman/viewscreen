package user

import (
	"encoding/json"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/textutil"
)

// Todo represents a single todo item
type Todo struct {
	Content    string `json:"content"`
	Status     string `json:"status"` // "pending", "in_progress", "completed"
	ActiveForm string `json:"activeForm"`
}

// TodoResult represents the tool_use_result for TodoWrite operations
type TodoResult struct {
	OldTodos []Todo `json:"oldTodos"`
	NewTodos []Todo `json:"newTodos"`
}

// TodoRenderer handles rendering of todo results with visual status indicators.
type TodoRenderer struct {
	styleApplier render.StyleApplier
}

// NewTodoRenderer creates a new TodoRenderer with the given dependencies.
func NewTodoRenderer(styleApplier render.StyleApplier) *TodoRenderer {
	return &TodoRenderer{
		styleApplier: styleApplier,
	}
}

// TryRender implements ResultRenderer interface.
// Attempts to render a todo list with visual status indicators.
// Returns true if it was a todo result and was rendered, false otherwise.
func (tr *TodoRenderer) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	if len(toolUseResult) == 0 {
		return false
	}

	var todoResult TodoResult
	if err := json.Unmarshal(toolUseResult, &todoResult); err != nil {
		return false
	}

	// Check if this looks like a todo result (has newTodos array)
	if len(todoResult.NewTodos) == 0 {
		return false
	}

	// Render each todo with status indicator
	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)

	for _, todo := range todoResult.NewTodos {
		var statusIndicator string
		var contentRenderer func(string) string

		switch todo.Status {
		case "completed":
			statusIndicator = tr.styleApplier.SuccessRender("✓")
			contentRenderer = tr.styleApplier.MutedRender
		case "in_progress":
			statusIndicator = tr.styleApplier.WarningRender("→")
			contentRenderer = func(s string) string { return s } // No special styling
		default: // "pending"
			statusIndicator = tr.styleApplier.MutedRender("○")
			contentRenderer = tr.styleApplier.MutedRender
		}

		content := todo.Content
		if todo.Status == "in_progress" && todo.ActiveForm != "" {
			content = todo.ActiveForm
		}

		pw.WriteLinef("%s %s", statusIndicator, contentRenderer(content))
	}

	return true
}
