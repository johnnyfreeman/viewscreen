package user

import (
	"encoding/json"
	"fmt"

	"github.com/johnnyfreeman/viewscreen/render"
)

// TodoRenderer handles rendering of todo results with visual status indicators.
type TodoRenderer struct {
	styleApplier StyleApplier
}

// NewTodoRenderer creates a new TodoRenderer with the given dependencies.
func NewTodoRenderer(styleApplier StyleApplier) *TodoRenderer {
	return &TodoRenderer{
		styleApplier: styleApplier,
	}
}

// TryRender attempts to render a todo result to the given output.
// Returns true if it was a todo result and was rendered, false otherwise.
func (tr *TodoRenderer) TryRender(out *render.Output, toolUseResult json.RawMessage, outputPrefix, outputContinue string) bool {
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
	for i, todo := range todoResult.NewTodos {
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

		// Use OutputPrefix for first line, OutputContinue for rest
		prefix := outputContinue
		if i == 0 {
			prefix = outputPrefix
		}

		content := todo.Content
		if todo.Status == "in_progress" && todo.ActiveForm != "" {
			content = todo.ActiveForm
		}

		fmt.Fprintf(out, "%s%s %s\n", prefix, statusIndicator, contentRenderer(content))
	}

	return true
}
