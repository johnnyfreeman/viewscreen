package tui

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
)

// Model is the main Bubbletea model for the TUI
type Model struct {
	width         int
	height        int
	viewport      viewport.Model
	spinner       spinner.Model
	state         *state.State
	content       *strings.Builder // pointer to avoid copy issues
	stdinDone     bool
	scanner       *bufio.Scanner
	sidebarStyles SidebarStyles
	ready         bool
	renderers     *events.RendererSet
}

// NewModel creates a new TUI model
func NewModel() Model {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	// Create scanner for stdin with large buffer
	scanner := bufio.NewScanner(os.Stdin)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	return Model{
		spinner:       s,
		state:         state.NewState(),
		content:       &strings.Builder{},
		scanner:       scanner,
		sidebarStyles: NewSidebarStyles(),
		renderers:     events.NewRendererSet(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		ReadStdinLine(m.scanner),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		case "pgup":
			m.viewport.HalfPageUp()
		case "pgdown":
			m.viewport.HalfPageDown()
		case "home", "g":
			m.viewport.GotoTop()
		case "end", "G":
			m.viewport.GotoBottom()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate content width (total - sidebar - border)
		contentWidth := max(m.width-sidebarWidth-3, 20)

		if !m.ready {
			// First time setup
			m.viewport = viewport.New(contentWidth, m.height-2)
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.viewport.Width = contentWidth
			m.viewport.Height = m.height - 2
		}

		// Update viewport content
		m.viewport.SetContent(m.content.String())

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		// Refresh viewport to animate spinner for pending tools
		if m.renderers.PendingTools.Len() > 0 {
			m.updateViewportWithPendingTools()
		}

	case RawLineMsg:
		// Parse the line and dispatch appropriate message
		parsedMsg := ParseEvent(msg.Line)
		if parsedMsg != nil {
			// Process the parsed event immediately
			m, cmd = m.processEvent(parsedMsg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Continue reading stdin
		cmds = append(cmds, ReadStdinLine(m.scanner))

	case StdinClosedMsg:
		m.stdinDone = true
		// Don't quit immediately - let user view the content
		// They can press 'q' to quit

	case ParseErrorMsg:
		// Optionally show parse errors in verbose mode
		if config.Verbose {
			m.content.WriteString("Parse error: " + msg.Line + "\n")
			m.viewport.SetContent(m.content.String())
			m.viewport.GotoBottom()
		}
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// processEvent handles a parsed event message
func (m Model) processEvent(msg tea.Msg) (Model, tea.Cmd) {
	r := m.renderers

	switch msg := msg.(type) {
	case SystemEventMsg:
		m.state.UpdateFromSystemEvent(msg.Event)
		rendered := r.System.RenderToString(msg.Event)
		m.content.WriteString(rendered)
		m.viewport.SetContent(m.content.String())
		m.viewport.GotoBottom()

	case AssistantEventMsg:
		m.state.IncrementTurnCount()
		// Buffer tool_use blocks using the events package helper
		if events.BufferToolUse(msg.Event, r.PendingTools, r.Stream) {
			// Set sidebar state to show the first pending tool
			for _, block := range msg.Event.Message.Content {
				if block.Type == "tool_use" && block.ID != "" {
					m.state.SetCurrentTool(block.Name, tools.GetToolArgFromBlock(block))
					break
				}
			}
		}
		// Render text blocks only (tools are buffered)
		rendered := r.Assistant.RenderToString(
			msg.Event,
			r.Stream.InTextBlock(),
			true, // Suppress tool rendering - we handle it separately
		)
		m.content.WriteString(rendered)
		r.Stream.ResetBlockState()
		m.updateViewportWithPendingTools()
		m.viewport.GotoBottom()

	case UserEventMsg:
		m.state.UpdateFromToolUseResult(msg.Event.ToolUseResult)
		// Match tool results with pending tools using the events package
		matched := events.MatchToolResults(msg.Event, r.PendingTools)

		// Render matched tool headers and set context
		var isNested bool
		for _, match := range matched {
			isNested = match.IsNested
			var str string
			var ctx tools.ToolContext
			if match.IsNested {
				str, ctx = tools.RenderNestedToolUseToString(match.Block)
			} else {
				str, ctx = tools.RenderToolUseToString(match.Block)
			}
			m.content.WriteString(str)
			r.User.SetToolContext(ctx)
		}

		// Clear tool state if no more pending tools
		if r.PendingTools.Len() == 0 {
			m.state.ClearCurrentTool()
		}
		// Render the tool result (with nested prefix if applicable)
		if isNested {
			rendered := r.User.RenderNestedToString(msg.Event)
			m.content.WriteString(rendered)
		} else {
			rendered := r.User.RenderToString(msg.Event)
			m.content.WriteString(rendered)
		}
		m.updateViewportWithPendingTools()
		m.viewport.GotoBottom()

	case StreamEventMsg:
		// Handle stream events
		rendered := r.Stream.RenderToString(msg.Event)
		if rendered != "" {
			m.content.WriteString(rendered)
			m.viewport.SetContent(m.content.String())
			m.viewport.GotoBottom()
		}

		// Update state for tool progress tracking
		if msg.Event.Event.Type == "content_block_start" && r.Stream.InToolUseBlock() {
			m.state.SetCurrentTool(r.Stream.CurrentBlockType(), "")
		}

	case ResultEventMsg:
		// Flush any orphaned pending tools using the events package
		orphaned := events.FlushOrphanedTools(r.PendingTools)
		for _, o := range orphaned {
			str, _ := tools.RenderToolUseToString(o.Block)
			m.content.WriteString(str)
			m.content.WriteString(style.OutputPrefix + style.Muted.Render("(no result)") + "\n")
		}
		m.state.ClearCurrentTool()
		m.state.UpdateFromResultEvent(msg.Event)
		rendered := r.Result.RenderToString(msg.Event)
		m.content.WriteString(rendered)
		m.viewport.SetContent(m.content.String()) // No pending tools, just set content
		m.viewport.GotoBottom()
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	return m.renderLayout()
}

// updateViewportWithPendingTools updates the viewport content, rendering pending tools with spinner
func (m *Model) updateViewportWithPendingTools() {
	content := m.content.String()
	// Render pending tools with spinner instead of bullet
	m.renderers.PendingTools.ForEach(func(id string, pending tools.PendingTool) {
		// Check if this is a nested child tool
		isNested := m.renderers.PendingTools.IsNested(pending)
		// Render tool header with spinner instead of bullet
		content += m.renderToolHeaderWithSpinner(pending.Block, isNested)
	})
	m.viewport.SetContent(content)
}

// renderToolHeaderWithSpinner renders a tool header with spinner instead of bullet
func (m *Model) renderToolHeaderWithSpinner(block types.ContentBlock, isNested bool) string {
	var input map[string]any
	if len(block.Input) > 0 {
		_ = json.Unmarshal(block.Input, &input)
	}

	opts := tools.HeaderOptions{
		Icon: m.spinner.View(),
	}
	if isNested {
		opts.Prefix = style.NestedPrefix
	}

	out := render.StringOutput()
	tools.RenderHeaderTo(out, block.Name, input, opts)
	return out.String()
}

// renderLayout composes the main content area and sidebar
func (m Model) renderLayout() string {
	// Render sidebar
	sidebar := RenderSidebar(m.state, m.spinner, m.height, m.sidebarStyles)

	// Render main content with viewport
	mainContent := m.viewport.View()

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sidebar)
}

// Run starts the TUI
func Run() error {
	// Initialize styles (needed for renderers)
	render.NewMarkdownRenderer(config.NoColor, 80)

	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
