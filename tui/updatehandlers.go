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
	// When search input is active, capture all keys for the search query
	if m.search.Active {
		return m.handleSearchKeyMsg(msg)
	}

	// Cancel auto-exit countdown on any key except q/ctrl+c
	if m.autoExitRemaining > 0 {
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		default:
			m.autoExitRemaining = 0
			return m, nil
		}
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit // tea.Quit is a func() Msg, which is a Cmd
	case "?":
		// Toggle help modal
		m.showHelpModal = !m.showHelpModal
	case "d":
		// Toggle details modal (only in header mode)
		if m.layoutMode == LayoutHeader {
			m.showDetailsModal = !m.showDetailsModal
		}
	case "esc":
		// Close any open modal or clear search
		if m.showHelpModal {
			m.showHelpModal = false
		} else if m.showDetailsModal {
			m.showDetailsModal = false
		} else if m.search.HasQuery() {
			m.search.Clear()
		}
	case "f":
		// Toggle follow mode (auto-scroll)
		if !m.showDetailsModal && !m.showHelpModal {
			m.followMode = !m.followMode
			if m.followMode {
				m.viewport.GotoBottom()
			}
		}
	case "/":
		// Enter search mode
		if !m.showDetailsModal && !m.showHelpModal {
			m.search.Enter()
		}
	case "n":
		// Next search match
		if !m.showDetailsModal && !m.showHelpModal && m.search.HasQuery() {
			m.search.NextMatch()
			m.scrollToSearchMatch()
		}
	case "N":
		// Previous search match
		if !m.showDetailsModal && !m.showHelpModal && m.search.HasQuery() {
			m.search.PrevMatch()
			m.scrollToSearchMatch()
		}
	case "up", "k":
		if !m.showDetailsModal && !m.showHelpModal {
			m.followMode = false
			m.viewport.ScrollUp(1)
		}
	case "down", "j":
		if !m.showDetailsModal && !m.showHelpModal {
			m.viewport.ScrollDown(1)
		}
	case "pgup":
		if !m.showDetailsModal && !m.showHelpModal {
			m.followMode = false
			m.viewport.HalfPageUp()
		}
	case "pgdown":
		if !m.showDetailsModal && !m.showHelpModal {
			m.viewport.HalfPageDown()
		}
	case "home", "g":
		if !m.showDetailsModal && !m.showHelpModal {
			m.followMode = false
			m.viewport.GotoTop()
		}
	case "end", "G":
		if !m.showDetailsModal && !m.showHelpModal {
			m.followMode = true
			m.viewport.GotoBottom()
		}
	}
	return m, nil
}

// handleSearchKeyMsg processes keyboard input while search mode is active.
func (m Model) handleSearchKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search, clear query
		m.search.Clear()
	case "enter":
		// Confirm search, exit input mode but keep query
		m.search.Exit()
	case "backspace":
		m.search.Backspace()
		m.search.UpdateMatches(m.content.String())
		m.scrollToSearchMatch()
	case "ctrl+c":
		return m, tea.Quit
	default:
		// Type character into search query
		text := msg.String()
		if len(text) == 1 {
			m.search.TypeRune(rune(text[0]))
			m.search.UpdateMatches(m.content.String())
			m.scrollToSearchMatch()
		}
	}
	return m, nil
}

// scrollToSearchMatch scrolls the viewport to show the current search match.
func (m *Model) scrollToSearchMatch() {
	line := m.search.CurrentLine()
	if line < 0 {
		return
	}
	m.viewport.SetYOffset(line)
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

	// Reserve space for search bar when active
	if m.search.Active || m.search.HasQuery() {
		contentHeight--
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

	// Update markdown renderer width so text reflows to fit visible viewport
	m.processor.SetWidth(contentWidth)

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
func (m Model) handleStdinClosed() (Model, tea.Cmd) {
	m.stdinDone = true
	if m.autoExit {
		m.autoExitRemaining = 5
		return m, AutoExitTick()
	}
	return m, nil
}

// handleAutoExitTick processes a countdown tick for auto-exit.
func (m Model) handleAutoExitTick() (Model, tea.Cmd) {
	if m.autoExitRemaining <= 0 {
		return m, nil
	}
	m.autoExitRemaining--
	if m.autoExitRemaining <= 0 {
		return m, tea.Quit
	}
	return m, AutoExitTick()
}

// handleParseError processes event parsing errors.
func (m Model) handleParseError(msg events.ParseError) Model {
	cfg := config.DefaultProvider{}
	if cfg.IsVerbose() {
		m.content.WriteString("Parse error: " + msg.Line + "\n")
		m.viewport.SetContent(m.content.String())
		if m.followMode {
			m.viewport.GotoBottom()
		}
	}
	return m
}
