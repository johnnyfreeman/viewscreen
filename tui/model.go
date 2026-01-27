package tui

import (
	"bufio"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/result"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/stream"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/system"
	"github.com/johnnyfreeman/viewscreen/tools"
	"github.com/johnnyfreeman/viewscreen/types"
	"github.com/johnnyfreeman/viewscreen/user"
)

// Model is the main Bubbletea model for the TUI
type Model struct {
	width          int
	height         int
	viewport       viewport.Model
	spinner        spinner.Model
	state          *state.State
	content        *strings.Builder // pointer to avoid copy issues
	stdinDone      bool
	scanner        *bufio.Scanner
	sidebarStyles  SidebarStyles
	streamRenderer *stream.Renderer
	ready          bool
	pendingTools   *tools.ToolUseTracker

	// Renderers for events
	systemRenderer    *system.Renderer
	assistantRenderer *assistant.Renderer
	userRenderer      *user.Renderer
	resultRenderer    *result.Renderer
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

	// Create renderers
	streamRenderer := stream.NewRenderer()

	return Model{
		spinner:        s,
		state:          state.NewState(),
		content:        &strings.Builder{},
		scanner:        scanner,
		sidebarStyles:  NewSidebarStyles(),
		streamRenderer: streamRenderer,
		pendingTools:   tools.NewToolUseTracker(),

		// Initialize package-level renderers
		systemRenderer:    system.NewRenderer(),
		assistantRenderer: assistant.NewRenderer(),
		userRenderer:      user.NewRenderer(),
		resultRenderer:    result.NewRenderer(),
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
		if m.pendingTools.Len() > 0 {
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
	switch msg := msg.(type) {
	case SystemEventMsg:
		m.state.UpdateFromSystemEvent(msg.Event)
		rendered := m.systemRenderer.RenderToString(msg.Event)
		m.content.WriteString(rendered)
		m.viewport.SetContent(m.content.String())
		m.viewport.GotoBottom()

	case AssistantEventMsg:
		m.state.IncrementTurnCount()
		// Buffer tool_use blocks - don't write to content yet
		// They'll be rendered with spinner in updateViewportWithPendingTools
		for _, block := range msg.Event.Message.Content {
			if block.Type == "tool_use" && block.ID != "" {
				if !m.streamRenderer.InToolUseBlock {
					m.pendingTools.Add(block.ID, block, msg.Event.ParentToolUseID)
					// Set tool state so sidebar shows spinner too
					m.state.SetCurrentTool(block.Name, tools.GetToolArgFromBlock(block))
				}
			}
		}
		// Render text blocks only (tools are buffered)
		rendered := m.assistantRenderer.RenderToString(
			msg.Event,
			m.streamRenderer.InTextBlock,
			true, // Suppress tool rendering - we handle it separately
		)
		m.content.WriteString(rendered)
		m.streamRenderer.ResetBlockState()
		m.updateViewportWithPendingTools()
		m.viewport.GotoBottom()

	case UserEventMsg:
		m.state.UpdateFromToolUseResult(msg.Event.ToolUseResult)
		// Match results with pending tools and render header + result together
		var isNested bool
		for _, content := range msg.Event.Message.Content {
			if content.Type == "tool_result" && content.ToolUseID != "" {
				if pending, ok := m.pendingTools.Get(content.ToolUseID); ok {
					// Check if this is a nested tool (parent is still pending)
					isNested = m.pendingTools.IsNested(pending)
					// Now write the tool header (with bullet, not spinner)
					var str string
					var ctx tools.ToolContext
					if isNested {
						str, ctx = tools.RenderNestedToolUseToString(pending.Block)
					} else {
						str, ctx = tools.RenderToolUseToString(pending.Block)
					}
					m.content.WriteString(str)
					// Set tool context for syntax highlighting of results
					m.userRenderer.SetToolContext(ctx)
					m.pendingTools.Remove(content.ToolUseID)
				}
			}
		}
		// Clear tool state if no more pending tools
		if m.pendingTools.Len() == 0 {
			m.state.ClearCurrentTool()
		}
		// Render the tool result (with nested prefix if applicable)
		if isNested {
			rendered := m.userRenderer.RenderNestedToString(msg.Event)
			m.content.WriteString(rendered)
		} else {
			rendered := m.userRenderer.RenderToString(msg.Event)
			m.content.WriteString(rendered)
		}
		m.updateViewportWithPendingTools()
		m.viewport.GotoBottom()

	case StreamEventMsg:
		// Handle stream events
		rendered := m.streamRenderer.RenderToString(msg.Event)
		if rendered != "" {
			m.content.WriteString(rendered)
			m.viewport.SetContent(m.content.String())
			m.viewport.GotoBottom()
		}

		// Update state for tool progress tracking
		if msg.Event.Event.Type == "content_block_start" && m.streamRenderer.InToolUseBlock {
			m.state.SetCurrentTool(m.streamRenderer.CurrentBlockType, "")
		}

	case ResultEventMsg:
		// Flush any orphaned pending tools before rendering result
		m.pendingTools.ForEach(func(id string, pending tools.PendingTool) {
			// Write with bullet (not spinner) since we're done
			str, _ := tools.RenderToolUseToString(pending.Block)
			m.content.WriteString(str)
			m.content.WriteString(style.OutputPrefix + style.Muted.Render("(no result)") + "\n")
		})
		m.pendingTools.Clear()
		m.state.ClearCurrentTool()
		m.state.UpdateFromResultEvent(msg.Event)
		rendered := m.resultRenderer.RenderToString(msg.Event)
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
	m.pendingTools.ForEach(func(id string, pending tools.PendingTool) {
		// Check if this is a nested child tool
		isNested := m.pendingTools.IsNested(pending)
		// Render tool header with spinner instead of bullet
		content += m.renderToolHeaderWithSpinner(pending.Block, isNested)
	})
	m.viewport.SetContent(content)
}

// renderToolHeaderWithSpinner renders a tool header with spinner instead of bullet
func (m *Model) renderToolHeaderWithSpinner(block types.ContentBlock, isNested bool) string {
	args := tools.GetToolArgFromBlock(block)

	// Truncate long args
	if len(args) > 80 {
		args = args[:77] + "..."
	}

	// Build header with spinner: ◐ToolName args
	// For nested tools, add the nested prefix: │ ◐ToolName args
	var result string
	if isNested {
		result = style.NestedPrefix + m.spinner.View() + style.ApplyThemeBoldGradient(block.Name)
	} else {
		result = m.spinner.View() + style.ApplyThemeBoldGradient(block.Name)
	}
	if args != "" {
		// Apply dotted underline for file path tools
		if tools.IsFilePathTool(block.Name) {
			result += " " + style.Muted.Render(style.DottedUnderline(args))
		} else {
			result += " " + style.Muted.Render(args)
		}
	}
	result += "\n"
	return result
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
