package tui

import (
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/events"
)

// handleKeyMsg processes keyboard input and returns the model and any command.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit // tea.Quit is a func() Msg, which is a Cmd
	case "d":
		// Toggle details modal (only in header mode)
		if m.layoutMode == LayoutHeader {
			m.showDetailsModal = !m.showDetailsModal
		}
	case "esc":
		// Close details modal
		if m.showDetailsModal {
			m.showDetailsModal = false
		}
	case "up", "k":
		if !m.showDetailsModal {
			m.viewport.ScrollUp(1)
		}
	case "down", "j":
		if !m.showDetailsModal {
			m.viewport.ScrollDown(1)
		}
	case "pgup":
		if !m.showDetailsModal {
			m.viewport.HalfPageUp()
		}
	case "pgdown":
		if !m.showDetailsModal {
			m.viewport.HalfPageDown()
		}
	case "home", "g":
		if !m.showDetailsModal {
			m.viewport.GotoTop()
		}
	case "end", "G":
		if !m.showDetailsModal {
			m.viewport.GotoBottom()
		}
	}
	return m, nil
}

// handleWindowSizeMsg processes terminal resize events.
func (m Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	// Determine layout mode based on terminal width
	if m.width < breakpointWidth {
		m.layoutMode = LayoutHeader
	} else {
		m.layoutMode = LayoutSidebar
	}

	// Calculate content dimensions based on layout mode
	var contentWidth, contentHeight int
	switch m.layoutMode {
	case LayoutHeader:
		// Full width minus padding
		contentWidth = max(m.width-2, 20)
		// Height minus header and margin
		contentHeight = m.height - headerHeight - 1
	default:
		// Width minus sidebar and border
		contentWidth = max(m.width-sidebarWidth-3, 20)
		contentHeight = m.height - 2
	}

	if !m.ready {
		// First time setup - use functional options for v2 API
		m.viewport = viewport.New(
			viewport.WithWidth(contentWidth),
			viewport.WithHeight(contentHeight),
		)
		m.viewport.YPosition = 0
		m.ready = true
	} else {
		m.viewport.SetWidth(contentWidth)
		m.viewport.SetHeight(contentHeight)
	}

	// Update viewport content
	m.viewport.SetContent(m.content.String())

	return m
}

// handleSpinnerTick processes spinner animation ticks.
func (m Model) handleSpinnerTick(msg spinner.TickMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)

	// Refresh viewport to animate spinner for pending tools
	if m.processor.HasPendingTools() {
		m.updateViewportWithPendingTools()
	}

	return m, cmd
}

// handleRawLine processes a line read from stdin.
func (m Model) handleRawLine(msg RawLineMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Parse the line and dispatch appropriate message
	parsedMsg := ParseEvent(msg.Line)
	if parsedMsg != nil {
		// Process the parsed event immediately
		var cmd tea.Cmd
		m, cmd = m.processEvent(parsedMsg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Continue reading stdin
	cmds = append(cmds, ReadStdinLine(m.scanner))

	return m, tea.Batch(cmds...)
}

// handleStdinClosed processes the stdin closed signal.
func (m Model) handleStdinClosed() Model {
	m.stdinDone = true
	// Don't quit immediately - let user view the content
	// They can press 'q' to quit
	return m
}

// handleParseError processes event parsing errors.
func (m Model) handleParseError(msg events.ParseError) Model {
	cfg := config.DefaultProvider{}
	if cfg.IsVerbose() {
		m.content.WriteString("Parse error: " + msg.Line + "\n")
		m.viewport.SetContent(m.content.String())
		m.viewport.GotoBottom()
	}
	return m
}
