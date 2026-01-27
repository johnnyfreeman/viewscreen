// Package tui provides the terminal user interface components.
//
// This file uses Ultraviolet for text styling to maintain proper style/content
// separation. Lipgloss is only used for layout concerns (Container with padding,
// width, height). All text coloring uses UV functions from style/uvstyle.go
// to avoid escape sequence conflicts when styled text is composed together.
package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/textutil"
)

// Logo - "viewscreen" in big ASCII art
var logoLines = []string{
	"█ █ █ █▀▀ █ █ █",
	"▀▄▀ █ ██▄ ▀▄▀▄▀",
	"█▀ █▀▀ █▀█ █▀▀ █▀▀ █▄ █",
	"▄█ █▄▄ █▀▄ ██▄ ██▄ █ ▀█",
}

const sidebarWidth = 30

// SidebarStyles holds the lipgloss styles for layout-only concerns.
// Text styling is handled by Ultraviolet functions in style/uvstyle.go
// to avoid escape sequence conflicts when styled content is composed.
type SidebarStyles struct {
	// Container is the only lipgloss style - used for layout (padding, width, height)
	Container lipgloss.Style
}

// NewSidebarStyles creates the sidebar styles.
// Only Container uses lipgloss (for layout). All text styling uses Ultraviolet.
func NewSidebarStyles() SidebarStyles {
	return SidebarStyles{
		Container: lipgloss.NewStyle().
			Width(sidebarWidth).
			Padding(1, 2),
	}
}

// SidebarRenderer renders individual sidebar sections.
// Each method is independent and testable.
type SidebarRenderer struct {
	styles  SidebarStyles
	width   int
	spinner spinner.Model
}

// NewSidebarRenderer creates a new sidebar renderer
func NewSidebarRenderer(styles SidebarStyles, spinner spinner.Model) *SidebarRenderer {
	return &SidebarRenderer{
		styles:  styles,
		width:   sidebarWidth,
		spinner: spinner,
	}
}

// RenderLogo renders the ASCII logo with gradient and decorations.
// Uses Ultraviolet for text styling to avoid escape sequence conflicts.
func (r *SidebarRenderer) RenderLogo() string {
	var sb strings.Builder

	deco := "· · · · · · · · · · · · ·"

	sb.WriteString(style.SidebarDecoText(deco))
	sb.WriteString("\n")
	sb.WriteString(style.MutedText("claude"))
	sb.WriteString("\n")

	for _, line := range logoLines {
		sb.WriteString(style.ApplyThemeBoldGradient(line))
		sb.WriteString("\n")
	}

	sb.WriteString(style.SidebarDecoText(deco))
	sb.WriteString("\n")

	return sb.String()
}

// RenderPrompt renders the user's prompt if available.
// Uses Ultraviolet for text styling.
func (r *SidebarRenderer) RenderPrompt(prompt string) string {
	if prompt == "" {
		return ""
	}
	wrapped := textutil.WrapText(prompt, r.width-4)
	return style.SidebarPromptText("\""+wrapped+"\"") + "\n\n"
}

// RenderLabelValue renders a label/value pair (used for model, turns, cost).
// Uses Ultraviolet for text styling.
func (r *SidebarRenderer) RenderLabelValue(label, value string) string {
	return style.SidebarHeaderText(label) + "\n" +
		style.SidebarValueText(value) + "\n\n"
}

// RenderSessionInfo renders model, turns, and cost
func (r *SidebarRenderer) RenderSessionInfo(model string, turns int, cost float64) string {
	var sb strings.Builder

	// Truncate model name if needed
	modelName := model
	if len(modelName) > r.width-4 {
		modelName = modelName[:r.width-7] + "..."
	}

	sb.WriteString(r.RenderLabelValue("Model", modelName))
	sb.WriteString(r.RenderLabelValue("Turns", fmt.Sprintf("%d", turns)))
	sb.WriteString(r.RenderLabelValue("Cost", fmt.Sprintf("$%.4f", cost)))

	return sb.String()
}

// RenderCurrentTool renders the currently running tool with spinner.
// Uses Ultraviolet for text styling.
func (r *SidebarRenderer) RenderCurrentTool(toolName, toolInput string) string {
	if toolName == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(style.SidebarHeaderText("Running"))
	sb.WriteString("\n")

	toolText := toolName
	if toolInput != "" && len(toolInput) < 20 {
		toolText += " " + toolInput
	}

	sb.WriteString(style.SpinnerText(r.spinner.View()))
	sb.WriteString(" ")
	sb.WriteString(style.SidebarTodoActiveText(textutil.Truncate(toolText, r.width-6)))
	sb.WriteString("\n\n")

	return sb.String()
}

// RenderTodo renders a single todo item.
// Uses Ultraviolet for text styling.
func (r *SidebarRenderer) RenderTodo(todo state.Todo) string {
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
		sb.WriteString(style.SpinnerText(r.spinner.View()))
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

// RenderTodos renders the todo list section.
// Uses Ultraviolet for text styling.
func (r *SidebarRenderer) RenderTodos(todos []state.Todo) string {
	if len(todos) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(style.SidebarHeaderText("Tasks"))
	sb.WriteString("\n")

	for _, todo := range todos {
		sb.WriteString(r.RenderTodo(todo))
	}

	return sb.String()
}

// Render renders the complete sidebar
func (r *SidebarRenderer) Render(s *state.State, height int) string {
	var sb strings.Builder

	sb.WriteString(r.RenderLogo())
	sb.WriteString("\n")
	sb.WriteString(r.RenderPrompt(s.Prompt))
	sb.WriteString(r.RenderSessionInfo(s.Model, s.TurnCount, s.TotalCost))

	if s.ToolInProgress {
		sb.WriteString(r.RenderCurrentTool(s.CurrentTool, s.CurrentToolInput))
	}

	sb.WriteString(r.RenderTodos(s.Todos))

	return r.styles.Container.Height(height - 2).Render(sb.String())
}

// RenderSidebar renders the sidebar with session info and todos.
// This is the main entry point, kept for backward compatibility.
func RenderSidebar(s *state.State, spinner spinner.Model, height int, styles SidebarStyles) string {
	r := NewSidebarRenderer(styles, spinner)
	return r.Render(s, height)
}
