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

const (
	sidebarWidth    = 30
	breakpointWidth = 80 // below this, use header mode
	headerHeight    = 1  // single line header
	modalWidth      = 40 // width of details modal
)

// LayoutMode determines how the UI is rendered based on terminal width
type LayoutMode int

const (
	LayoutSidebar LayoutMode = iota // sidebar on right (>= 80 cols)
	LayoutHeader                    // header on top (< 80 cols)
)

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

// SidebarRenderer renders the sidebar by composing focused sub-renderers.
// Each sub-renderer (LogoRenderer, TodoRenderer) handles one concern.
type SidebarRenderer struct {
	styles  SidebarStyles
	width   int
	logo    *LogoRenderer
	todo    *TodoRenderer
	spinner spinner.Model
}

// NewSidebarRenderer creates a new sidebar renderer with composed sub-renderers.
func NewSidebarRenderer(styles SidebarStyles, spinner spinner.Model) *SidebarRenderer {
	return &SidebarRenderer{
		styles:  styles,
		width:   sidebarWidth,
		logo:    NewLogoRenderer(),
		todo:    NewTodoRenderer(sidebarWidth, spinner),
		spinner: spinner,
	}
}

// RenderLogo delegates to the LogoRenderer.
func (r *SidebarRenderer) RenderLogo() string {
	return r.logo.Render()
}

// RenderPrompt renders the user's prompt if available.
func (r *SidebarRenderer) RenderPrompt(prompt string) string {
	if prompt == "" {
		return ""
	}
	wrapped := textutil.WrapText(prompt, r.width-4)
	return style.SidebarPromptText("\""+wrapped+"\"") + "\n\n"
}

// RenderLabelValue renders a label/value pair (used for model, turns, cost).
func (r *SidebarRenderer) RenderLabelValue(label, value string) string {
	return style.SidebarHeaderText(label) + "\n" +
		style.SidebarValueText(value) + "\n\n"
}

// RenderSessionInfo renders model, turns, and cost.
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

	sb.WriteString(r.spinner.View())
	sb.WriteString(" ")
	sb.WriteString(style.SidebarTodoActiveText(textutil.Truncate(toolText, r.width-6)))
	sb.WriteString("\n\n")

	return sb.String()
}

// RenderTodo delegates to the TodoRenderer.
func (r *SidebarRenderer) RenderTodo(todo state.Todo) string {
	return r.todo.RenderItem(todo)
}

// RenderTodos delegates to the TodoRenderer.
func (r *SidebarRenderer) RenderTodos(todos []state.Todo) string {
	return r.todo.RenderList(todos)
}

// Render renders the complete sidebar by composing all sections.
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

// HeaderStyles holds the lipgloss styles for header layout.
type HeaderStyles struct {
	Container lipgloss.Style
	Modal     lipgloss.Style
}

// NewHeaderStyles creates header styles.
func NewHeaderStyles() HeaderStyles {
	return HeaderStyles{
		Container: lipgloss.NewStyle(),
		Modal: lipgloss.NewStyle().
			Width(modalWidth).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(string(style.CurrentTheme.FgMuted))),
	}
}

// RenderHeader renders a single-line header for narrow terminals.
// Format: ─── VIEWSCREEN ─── model │ 5 │ $0.12 ─── [d] ───
func RenderHeader(s *state.State, width int) string {
	logo := NewLogoRenderer()

	// Build the info section: model │ turns │ cost
	model := s.Model
	maxModelLen := 15
	if len(model) > maxModelLen {
		model = model[:maxModelLen-2] + ".."
	}

	info := fmt.Sprintf("%s %s %d %s $%.2f",
		model,
		style.MutedText("│"),
		s.TurnCount,
		style.MutedText("│"),
		s.TotalCost)

	// Fixed parts
	title := logo.RenderTitle()
	keyHint := style.MutedText("[d]")

	// Calculate decoration lengths
	// Raw lengths (without ANSI): "─── " + "VIEWSCREEN" + " ─── " + info + " ─── " + "[d]" + " ───"
	titleLen := 10 // "VIEWSCREEN"
	infoLen := len(model) + 3 + len(fmt.Sprintf("%d", s.TurnCount)) + 3 + len(fmt.Sprintf("$%.2f", s.TotalCost))
	keyHintLen := 3 // "[d]"
	fixedLen := 4 + titleLen + 5 + infoLen + 5 + keyHintLen + 4 // decorations + spaces

	// Remaining space for decorations
	remaining := max(width-fixedLen, 4)

	// Distribute decoration evenly
	leftDeco := strings.Repeat("─", 3)
	midDeco := strings.Repeat("─", 3)
	rightDeco := strings.Repeat("─", max(remaining, 1))

	return fmt.Sprintf("%s %s %s %s %s %s %s",
		style.MutedText(leftDeco),
		title,
		style.MutedText(midDeco),
		info,
		style.MutedText(midDeco),
		keyHint,
		style.MutedText(rightDeco))
}

// RenderDetailsModal renders the details modal overlay.
func RenderDetailsModal(s *state.State, sp spinner.Model, width, height int, styles HeaderStyles) string {
	r := NewSidebarRenderer(NewSidebarStyles(), sp)

	var sb strings.Builder

	// Logo
	sb.WriteString(r.RenderLogo())
	sb.WriteString("\n")

	// Prompt if available
	if s.Prompt != "" {
		sb.WriteString(r.RenderPrompt(s.Prompt))
	}

	// Session info
	sb.WriteString(r.RenderSessionInfo(s.Model, s.TurnCount, s.TotalCost))

	// Current tool
	if s.ToolInProgress {
		sb.WriteString(r.RenderCurrentTool(s.CurrentTool, s.CurrentToolInput))
	}

	// Todos
	sb.WriteString(r.RenderTodos(s.Todos))

	// Close hint
	sb.WriteString("\n")
	sb.WriteString(style.MutedText("Press d or Esc to close"))

	modalContent := styles.Modal.Render(sb.String())

	// Center the modal
	modalHeight := strings.Count(modalContent, "\n") + 1
	modalWidth := lipgloss.Width(modalContent)

	topPadding := max((height-modalHeight)/2, 0)
	leftPadding := max((width-modalWidth)/2, 0)

	// Build centered modal
	var result strings.Builder
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}

	for _, line := range strings.Split(modalContent, "\n") {
		result.WriteString(strings.Repeat(" ", leftPadding))
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}
