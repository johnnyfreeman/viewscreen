package tui

import (
	"bufio"
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
	"github.com/johnnyfreeman/viewscreen/tools"
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
	processor     *events.EventProcessor
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

	st := state.NewState()
	return Model{
		spinner:       s,
		state:         st,
		content:       &strings.Builder{},
		scanner:       scanner,
		sidebarStyles: NewSidebarStyles(),
		processor:     events.NewEventProcessor(st),
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
		if m.processor.HasPendingTools() {
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
	// Convert TUI message to events.Event
	var event events.Event
	switch msg := msg.(type) {
	case SystemEventMsg:
		event = events.SystemEvent{Data: msg.Event}
	case AssistantEventMsg:
		event = events.AssistantEvent{Data: msg.Event}
	case UserEventMsg:
		event = events.UserEvent{Data: msg.Event}
	case StreamEventMsg:
		event = events.StreamEvent{Data: msg.Event}
	case ResultEventMsg:
		event = events.ResultEvent{Data: msg.Event}
	default:
		return m, nil
	}

	// Process the event through the EventProcessor
	result := m.processor.Process(event)

	// Append rendered content
	if result.Rendered != "" {
		m.content.WriteString(result.Rendered)
	}

	// Update viewport based on whether there are pending tools
	if result.HasPendingTools {
		m.updateViewportWithPendingTools()
	} else {
		m.viewport.SetContent(m.content.String())
	}
	m.viewport.GotoBottom()

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
	m.processor.ForEachPendingTool(func(id string, pending tools.PendingTool) {
		content += m.processor.RenderPendingTool(pending, m.spinner.View())
	})
	m.viewport.SetContent(content)
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
