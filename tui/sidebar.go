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
	"time"

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

// RenderTokenUsage renders the token usage section.
func (r *SidebarRenderer) RenderTokenUsage(input, output int) string {
	if input == 0 && output == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(r.RenderLabelValue("Tokens",
		fmt.Sprintf("↑%s ↓%s", formatTokenCount(input), formatTokenCount(output))))
	return sb.String()
}

// RenderCacheUsage renders the cache read/created token section.
func (r *SidebarRenderer) RenderCacheUsage(cacheRead, cacheCreated int) string {
	if cacheRead == 0 && cacheCreated == 0 {
		return ""
	}

	value := fmt.Sprintf("⟳%s ✦%s", formatTokenCount(cacheRead), formatTokenCount(cacheCreated))
	return r.RenderLabelValue("Cache", value)
}

// formatTokenCount formats a token count compactly (e.g., 1234 -> "1.2k", 1234567 -> "1.2M").
func formatTokenCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// RenderElapsed renders the session elapsed time.
func (r *SidebarRenderer) RenderElapsed(elapsed time.Duration) string {
	return r.RenderLabelValue("Elapsed", formatDuration(elapsed))
}

// RenderCostRate renders the cost rate ($/min) section.
func (r *SidebarRenderer) RenderCostRate(costRate float64) string {
	if costRate == 0 {
		return ""
	}
	return r.RenderLabelValue("Rate", formatCostRate(costRate))
}

// formatCostRate formats a cost-per-minute value as a compact string.
// Uses different precision based on the magnitude.
func formatCostRate(rate float64) string {
	switch {
	case rate >= 1.0:
		return fmt.Sprintf("$%.2f/min", rate)
	case rate >= 0.01:
		return fmt.Sprintf("$%.3f/min", rate)
	default:
		return fmt.Sprintf("$%.4f/min", rate)
	}
}

// formatDuration formats a duration as a compact human-readable string.
// Examples: "5s", "1m 23s", "1h 5m"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
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

// RenderFollowIndicator renders the follow mode status indicator.
func (r *SidebarRenderer) RenderFollowIndicator(followMode bool) string {
	if followMode {
		return ""
	}
	return style.WarningText("⏸ Paused") + " " + style.MutedText("[f]") + "\n\n"
}

// RenderAutoExitStatus renders the auto-exit countdown or stream complete status.
func (r *SidebarRenderer) RenderAutoExitStatus(stdinDone bool, autoExitRemaining int) string {
	if !stdinDone {
		return ""
	}
	if autoExitRemaining > 0 {
		return style.MutedText(fmt.Sprintf("Exiting in %ds...", autoExitRemaining)) + "\n" +
			style.MutedText("space to skip") + "\n\n"
	}
	return style.MutedText("Stream complete") + "\n\n"
}

// ScrollPosition holds the viewport scroll state for display.
type ScrollPosition struct {
	AtTop   bool
	AtBottom bool
	Percent float64 // 0.0 to 1.0
}

// FormatScrollPosition returns a compact scroll position string.
// Returns "Top", "Bot", or "XX%" based on the scroll position.
func FormatScrollPosition(pos ScrollPosition) string {
	if pos.AtTop {
		return "Top"
	}
	if pos.AtBottom {
		return "Bot"
	}
	pct := int(pos.Percent * 100)
	if pct < 1 {
		pct = 1
	}
	if pct > 99 {
		pct = 99
	}
	return fmt.Sprintf("%d%%", pct)
}

// RenderScrollPosition renders the scroll position indicator.
func (r *SidebarRenderer) RenderScrollPosition(pos ScrollPosition) string {
	return r.RenderLabelValue("Position", FormatScrollPosition(pos))
}

// Render renders the complete sidebar by composing all sections.
func (r *SidebarRenderer) Render(s *state.State, height int, followMode bool, scrollPos ScrollPosition, stdinDone bool, autoExitRemaining int) string {
	var sb strings.Builder

	sb.WriteString(r.RenderLogo())
	sb.WriteString("\n")
	sb.WriteString(r.RenderAutoExitStatus(stdinDone, autoExitRemaining))
	sb.WriteString(r.RenderFollowIndicator(followMode))
	sb.WriteString(r.RenderPrompt(s.Prompt))
	sb.WriteString(r.RenderSessionInfo(s.Model, s.TurnCount, s.TotalCost))
	sb.WriteString(r.RenderCostRate(s.CostRate()))
	sb.WriteString(r.RenderElapsed(s.Elapsed()))
	sb.WriteString(r.RenderScrollPosition(scrollPos))
	sb.WriteString(r.RenderTokenUsage(s.InputTokens, s.OutputTokens))
	sb.WriteString(r.RenderCacheUsage(s.CacheRead, s.CacheCreated))

	if s.ToolInProgress {
		sb.WriteString(r.RenderCurrentTool(s.CurrentTool, s.CurrentToolInput))
	}

	sb.WriteString(r.RenderTodos(s.Todos))

	return r.styles.Container.Height(height - 2).Render(sb.String())
}

// RenderSidebar renders the sidebar with session info and todos.
// This is the main entry point, kept for backward compatibility.
func RenderSidebar(s *state.State, spinner spinner.Model, height int, styles SidebarStyles, followMode bool, scrollPos ScrollPosition, stdinDone bool, autoExitRemaining int) string {
	r := NewSidebarRenderer(styles, spinner)
	return r.Render(s, height, followMode, scrollPos, stdinDone, autoExitRemaining)
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
// Format: ─── VIEWSCREEN ─── model │ 5 │ $0.12 │ 42% ─── [?] ───
func RenderHeader(s *state.State, width int, followMode bool, scrollPos ScrollPosition, stdinDone bool, autoExitRemaining int) string {
	logo := NewLogoRenderer()

	// Build the info section: model │ turns │ cost
	model := s.Model
	maxModelLen := 15
	if len(model) > maxModelLen {
		model = model[:maxModelLen-2] + ".."
	}

	elapsed := formatDuration(s.Elapsed())
	scrollStr := FormatScrollPosition(scrollPos)
	info := fmt.Sprintf("%s %s %d %s $%.2f %s %s %s %s",
		model,
		style.MutedText("│"),
		s.TurnCount,
		style.MutedText("│"),
		s.TotalCost,
		style.MutedText("│"),
		elapsed,
		style.MutedText("│"),
		scrollStr)

	// Auto-exit hint
	autoExitHint := ""
	autoExitHintLen := 0
	if stdinDone && autoExitRemaining > 0 {
		autoExitHint = style.MutedText(fmt.Sprintf("Exit %ds", autoExitRemaining))
		autoExitHintLen = 6 + len(fmt.Sprintf("%d", autoExitRemaining)) // "Exit " + N + "s"
	} else if stdinDone {
		autoExitHint = style.MutedText("Done")
		autoExitHintLen = 4
	}

	// Fixed parts
	title := logo.RenderTitle()
	keyHint := style.MutedText("[?]")

	// Paused indicator when follow mode is off
	pausedHint := ""
	pausedLen := 0
	if !followMode {
		pausedHint = style.WarningText("⏸")
		pausedLen = 1 // single character width
	}

	// Calculate decoration lengths
	// Raw lengths (without ANSI): "─── " + "VIEWSCREEN" + " ─── " + info + " ─── " + [paused] + "[?]" + " ───"
	titleLen := 10 // "VIEWSCREEN"
	infoLen := len(model) + 3 + len(fmt.Sprintf("%d", s.TurnCount)) + 3 + len(fmt.Sprintf("$%.2f", s.TotalCost)) + 3 + len(elapsed) + 3 + len(scrollStr)
	keyHintLen := 3 // "[?]"
	pausedExtra := 0
	if pausedLen > 0 {
		pausedExtra = pausedLen + 1 // pause icon + space
	}
	autoExitExtra := 0
	if autoExitHintLen > 0 {
		autoExitExtra = autoExitHintLen + 1 // hint + space
	}
	fixedLen := 4 + titleLen + 5 + infoLen + 5 + pausedExtra + autoExitExtra + keyHintLen + 4 // decorations + spaces

	// Remaining space for decorations
	remaining := max(width-fixedLen, 4)

	// Distribute decoration evenly
	leftDeco := strings.Repeat("─", 3)
	midDeco := strings.Repeat("─", 3)
	rightDeco := strings.Repeat("─", max(remaining, 1))

	// Build trailing indicators: [autoExit] [paused] [?]
	var trailing []string
	if autoExitHint != "" {
		trailing = append(trailing, autoExitHint)
	}
	if pausedHint != "" {
		trailing = append(trailing, pausedHint)
	}
	trailing = append(trailing, keyHint)
	trailingStr := strings.Join(trailing, " ")

	return fmt.Sprintf("%s %s %s %s %s %s %s",
		style.MutedText(leftDeco),
		title,
		style.MutedText(midDeco),
		info,
		style.MutedText(midDeco),
		trailingStr,
		style.MutedText(rightDeco))
}

// RenderHelpModal renders the keybindings help modal overlay.
func RenderHelpModal(width, height int, styles HeaderStyles, autoExitActive bool) string {
	var sb strings.Builder

	// Title
	sb.WriteString(style.SidebarValueText("Keybindings"))
	sb.WriteString("\n\n")

	// Keybinding entries
	bindings := []struct {
		key  string
		desc string
	}{
		{"j / ↓", "Scroll down"},
		{"k / ↑", "Scroll up"},
		{"PgDn", "Half page down"},
		{"PgUp", "Half page up"},
		{"g / Home", "Go to top"},
		{"G / End", "Go to bottom"},
		{"/", "Search"},
		{"n / N", "Next / prev match"},
		{"f", "Toggle follow mode"},
		{"d", "Toggle details"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}

	if autoExitActive {
		// Insert before the last entry (quit)
		bindings = append(bindings[:len(bindings)-1],
			struct{ key, desc string }{"space", "Skip countdown"},
			struct{ key, desc string }{"any key", "Cancel and browse"},
			bindings[len(bindings)-1],
		)
	}

	for _, b := range bindings {
		key := style.SidebarTodoActiveText(fmt.Sprintf("%-10s", b.key))
		desc := style.SidebarHeaderText(b.desc)
		sb.WriteString(key + " " + desc + "\n")
	}

	// Close hint
	sb.WriteString("\n")
	sb.WriteString(style.MutedText("Press ? or Esc to close"))

	modalContent := styles.Modal.Render(sb.String())

	// Center the modal
	modalHeight := strings.Count(modalContent, "\n") + 1
	modalWidth := lipgloss.Width(modalContent)

	topPadding := max((height-modalHeight)/2, 0)
	leftPadding := max((width-modalWidth)/2, 0)

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

// RenderDetailsModal renders the details modal overlay.
func RenderDetailsModal(s *state.State, sp spinner.Model, width, height int, styles HeaderStyles, followMode bool, scrollPos ScrollPosition, stdinDone bool, autoExitRemaining int) string {
	r := NewSidebarRenderer(NewSidebarStyles(), sp)

	var sb strings.Builder

	// Logo
	sb.WriteString(r.RenderLogo())
	sb.WriteString("\n")

	// Auto-exit status
	sb.WriteString(r.RenderAutoExitStatus(stdinDone, autoExitRemaining))

	// Follow mode indicator
	sb.WriteString(r.RenderFollowIndicator(followMode))

	// Prompt if available
	if s.Prompt != "" {
		sb.WriteString(r.RenderPrompt(s.Prompt))
	}

	// Session info
	sb.WriteString(r.RenderSessionInfo(s.Model, s.TurnCount, s.TotalCost))
	sb.WriteString(r.RenderCostRate(s.CostRate()))
	sb.WriteString(r.RenderElapsed(s.Elapsed()))
	sb.WriteString(r.RenderScrollPosition(scrollPos))
	sb.WriteString(r.RenderTokenUsage(s.InputTokens, s.OutputTokens))
	sb.WriteString(r.RenderCacheUsage(s.CacheRead, s.CacheCreated))

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
