package tui

import (
	"strings"

	"charm.land/bubbles/v2/spinner"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/textutil"
)

// TodoRenderer renders todo items with status indicators.
// It's a focused component that handles only todo-related rendering.
type TodoRenderer struct {
	width   int
	spinner spinner.Model
}

// NewTodoRenderer creates a new todo renderer.
func NewTodoRenderer(width int, spinner spinner.Model) *TodoRenderer {
	return &TodoRenderer{
		width:   width,
		spinner: spinner,
	}
}

// RenderItem renders a single todo item with appropriate status indicator.
func (r *TodoRenderer) RenderItem(todo state.Todo) string {
	var sb strings.Builder
	maxWidth := r.width - 6

	switch todo.Status {
	case "completed":
		sb.WriteString(style.SuccessText("✓ "))
		text := todo.Subject
		if text == "" {
			text = todo.ActiveForm
		}
		sb.WriteString(style.SidebarTodoDoneText(textutil.Truncate(text, maxWidth)))

	case "in_progress":
		sb.WriteString(r.spinner.View())
		text := todo.ActiveForm
		if text == "" {
			text = todo.Subject
		}
		sb.WriteString(style.SidebarTodoActiveText(textutil.Truncate(text, maxWidth)))

	default: // pending
		sb.WriteString(style.SidebarTodoPendingText("○ "))
		text := todo.Subject
		if text == "" {
			text = todo.ActiveForm
		}
		sb.WriteString(style.SidebarTodoPendingText(textutil.Truncate(text, maxWidth)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// RenderList renders a list of todos with a header.
// Returns empty string if the list is empty.
func (r *TodoRenderer) RenderList(todos []state.Todo) string {
	if len(todos) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(style.SidebarHeaderText("Tasks"))
	sb.WriteString("\n")

	for _, todo := range todos {
		sb.WriteString(r.RenderItem(todo))
	}

	return sb.String()
}
