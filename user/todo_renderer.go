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
//
// This renderer uses Ultraviolet-based styling methods (UV*) for proper
// style/content separation. This ensures that styled todo items can be
// safely composed with other styles (like the prefixed writer's output
// formatting) without ANSI escape sequence conflicts.
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
	// Uses Ultraviolet-based styling for composition-safe output
	pw := textutil.NewPrefixedWriter(ctx.Output, ctx.OutputPrefix, ctx.OutputContinue)

	for _, todo := range todoResult.NewTodos {
		var statusIndicator string
		var contentRenderer func(string) string

		switch todo.Status {
		case "completed":
			// Use UV methods for proper style/content separation
			statusIndicator = tr.styleApplier.UVSuccessText("✓")
			contentRenderer = tr.styleApplier.UVMutedText
		case "in_progress":
			statusIndicator = tr.styleApplier.UVWarningText("→")
			contentRenderer = func(s string) string { return s } // No special styling
		default: // "pending"
			statusIndicator = tr.styleApplier.UVMutedText("○")
			contentRenderer = tr.styleApplier.UVMutedText
		}

		content := todo.Content
		if todo.Status == "in_progress" && todo.ActiveForm != "" {
			content = todo.ActiveForm
		}

		pw.WriteLinef("%s %s", statusIndicator, contentRenderer(content))
	}

	return true
}
