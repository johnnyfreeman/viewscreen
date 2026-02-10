package tui

import (
	"fmt"
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

// RenderProgressBar renders a compact progress bar with completion count.
// Format: ████░░░░░░ 3/8
// The bar width adapts to the available space minus the count label.
func (r *TodoRenderer) RenderProgressBar(completed, total int) string {
	if total == 0 {
		return ""
	}

	label := fmt.Sprintf(" %d/%d", completed, total)
	barWidth := max(r.width-4-len(label), 4) // padding and label

	filled := 0
	if total > 0 {
		filled = (completed * barWidth) / total
	}
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	if completed == total {
		return style.SuccessText(bar) + style.SidebarTodoDoneText(label) + "\n"
	}
	return style.AccentText(bar) + style.SidebarHeaderText(label) + "\n"
}

// RenderList renders a list of todos with a header and progress bar.
// Returns empty string if the list is empty.
func (r *TodoRenderer) RenderList(todos []state.Todo) string {
	if len(todos) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(style.SidebarHeaderText("Tasks"))
	sb.WriteString("\n")

	// Progress bar when there are 2+ todos
	if len(todos) >= 2 {
		completed := 0
		for _, todo := range todos {
			if todo.Status == "completed" {
				completed++
			}
		}
		sb.WriteString(r.RenderProgressBar(completed, len(todos)))
	}

	for _, todo := range todos {
		sb.WriteString(r.RenderItem(todo))
	}

	return sb.String()
}
