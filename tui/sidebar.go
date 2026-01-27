package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
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

// SidebarStyles holds the lipgloss styles for the sidebar
type SidebarStyles struct {
	Container   lipgloss.Style
	Logo        lipgloss.Style
	Header      lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	TodoPending lipgloss.Style
	TodoActive  lipgloss.Style
	TodoDone    lipgloss.Style
	Divider     lipgloss.Style
	Prompt      lipgloss.Style
}

// NewSidebarStyles creates the sidebar styles
func NewSidebarStyles() SidebarStyles {
	return SidebarStyles{
		Container: lipgloss.NewStyle().
			Width(sidebarWidth).
			Padding(1, 2),
		Logo: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")),
		TodoPending: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		TodoActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")), // White like todo renderer
		TodoDone: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")), // Muted like todo renderer
		Divider: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		Prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),
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

// RenderLogo renders the ASCII logo with gradient and decorations
func (r *SidebarRenderer) RenderLogo() string {
	var sb strings.Builder

	darkDeco := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	deco := "· · · · · · · · · · · · ·"

	sb.WriteString(darkDeco.Render(deco))
	sb.WriteString("\n")
	sb.WriteString(style.Muted.Render("claude"))
	sb.WriteString("\n")

	for _, line := range logoLines {
		sb.WriteString(style.ApplyThemeBoldGradient(line))
		sb.WriteString("\n")
	}

	sb.WriteString(darkDeco.Render(deco))
	sb.WriteString("\n")

	return sb.String()
}

// RenderPrompt renders the user's prompt if available
func (r *SidebarRenderer) RenderPrompt(prompt string) string {
	if prompt == "" {
		return ""
	}
	wrapped := textutil.WrapText(prompt, r.width-4)
	return r.styles.Prompt.Render("\""+wrapped+"\"") + "\n\n"
}

// RenderLabelValue renders a label/value pair (used for model, turns, cost)
func (r *SidebarRenderer) RenderLabelValue(label, value string) string {
	return r.styles.Label.Render(label) + "\n" +
		r.styles.Value.Render(value) + "\n\n"
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

// RenderCurrentTool renders the currently running tool with spinner
func (r *SidebarRenderer) RenderCurrentTool(toolName, toolInput string) string {
	if toolName == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(r.styles.Header.Render("Running"))
	sb.WriteString("\n")

	toolText := toolName
	if toolInput != "" && len(toolInput) < 20 {
		toolText += " " + toolInput
	}

	sb.WriteString(r.spinner.View())
	sb.WriteString(" ")
	sb.WriteString(r.styles.TodoActive.Render(textutil.Truncate(toolText, r.width-6)))
	sb.WriteString("\n\n")

	return sb.String()
}

// RenderTodo renders a single todo item
func (r *SidebarRenderer) RenderTodo(todo state.Todo) string {
	var sb strings.Builder
	maxWidth := r.width - 6

	switch todo.Status {
	case "completed":
		sb.WriteString(style.Success.Render("✓ "))
		text := todo.Subject
		if text == "" {
			text = todo.ActiveForm
		}
		sb.WriteString(r.styles.TodoDone.Render(textutil.Truncate(text, maxWidth)))

	case "in_progress":
		sb.WriteString(r.spinner.View())
		text := todo.ActiveForm
		if text == "" {
			text = todo.Subject
		}
		sb.WriteString(r.styles.TodoActive.Render(textutil.Truncate(text, maxWidth)))

	default: // pending
		sb.WriteString(r.styles.TodoPending.Render("○ "))
		text := todo.Subject
		if text == "" {
			text = todo.ActiveForm
		}
		sb.WriteString(r.styles.TodoPending.Render(textutil.Truncate(text, maxWidth)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// RenderTodos renders the todo list section
func (r *SidebarRenderer) RenderTodos(todos []state.Todo) string {
	if len(todos) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(r.styles.Header.Render("Tasks"))
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
