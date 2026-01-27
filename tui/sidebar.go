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

// RenderSidebar renders the sidebar with session info and todos
func RenderSidebar(s *state.State, spinner spinner.Model, height int, styles SidebarStyles) string {
	var sb strings.Builder

	// Logo with gradient
	sb.WriteString(renderLogo())
	sb.WriteString("\n")

	// Prompt (if available)
	if s.Prompt != "" {
		// Word-wrap the prompt to fit sidebar
		wrapped := textutil.WrapText(s.Prompt, sidebarWidth-4)
		sb.WriteString(styles.Prompt.Render("\""+wrapped+"\""))
		sb.WriteString("\n\n")
	}

	// Model name (truncate if needed)
	modelName := s.Model
	if len(modelName) > sidebarWidth-4 {
		modelName = modelName[:sidebarWidth-7] + "..."
	}
	sb.WriteString(styles.Label.Render("Model"))
	sb.WriteString("\n")
	sb.WriteString(styles.Value.Render(modelName))
	sb.WriteString("\n\n")

	// Turn count
	sb.WriteString(styles.Label.Render("Turns"))
	sb.WriteString("\n")
	sb.WriteString(styles.Value.Render(fmt.Sprintf("%d", s.TurnCount)))
	sb.WriteString("\n\n")

	// Cost
	sb.WriteString(styles.Label.Render("Cost"))
	sb.WriteString("\n")
	sb.WriteString(styles.Value.Render(fmt.Sprintf("$%.4f", s.TotalCost)))
	sb.WriteString("\n\n")

	// Current tool (if any)
	if s.ToolInProgress {
		sb.WriteString(styles.Header.Render("Running"))
		sb.WriteString("\n")
		toolText := s.CurrentTool
		if s.CurrentToolInput != "" && len(s.CurrentToolInput) < 20 {
			toolText += " " + s.CurrentToolInput
		}
		sb.WriteString(spinner.View())
		sb.WriteString(" ")
		sb.WriteString(styles.TodoActive.Render(textutil.Truncate(toolText, sidebarWidth-6)))
		sb.WriteString("\n\n")
	}

	// Tasks Header
	if len(s.Todos) > 0 {
		sb.WriteString(styles.Header.Render("Tasks"))
		sb.WriteString("\n")

		for _, todo := range s.Todos {
			switch todo.Status {
			case "completed":
				sb.WriteString(style.Success.Render("✓ "))
				text := todo.Subject
				if text == "" {
					text = todo.ActiveForm
				}
				sb.WriteString(styles.TodoDone.Render(textutil.Truncate(text, sidebarWidth-6)))
			case "in_progress":
				sb.WriteString(spinner.View())
				text := todo.ActiveForm
				if text == "" {
					text = todo.Subject
				}
				sb.WriteString(styles.TodoActive.Render(textutil.Truncate(text, sidebarWidth-6)))
			default: // pending
				sb.WriteString(styles.TodoPending.Render("○ "))
				text := todo.Subject
				if text == "" {
					text = todo.ActiveForm
				}
				sb.WriteString(styles.TodoPending.Render(textutil.Truncate(text, sidebarWidth-6)))
			}
			sb.WriteString("\n")
		}
	}

	content := sb.String()

	// Apply container style with fixed width and height
	return styles.Container.Height(height - 2).Render(content)
}

// renderLogo renders the ASCII logo with a gradient and decorations
func renderLogo() string {
	var sb strings.Builder

	// Subtle decoration style
	darkDeco := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	// Top decoration (dots)
	topDeco := "· · · · · · · · · · · · ·"
	sb.WriteString(darkDeco.Render(topDeco))
	sb.WriteString("\n")

	// "claude" in small text, left-aligned
	sb.WriteString(style.Muted.Render("claude"))
	sb.WriteString("\n")

	// "viewscreen" logo with gradient
	for _, line := range logoLines {
		sb.WriteString(style.ApplyThemeBoldGradient(line))
		sb.WriteString("\n")
	}

	// Bottom decoration (dots)
	botDeco := "· · · · · · · · · · · · ·"
	sb.WriteString(darkDeco.Render(botDeco))
	sb.WriteString("\n")

	return sb.String()
}
