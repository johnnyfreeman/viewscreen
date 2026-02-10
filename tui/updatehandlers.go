package tui

import (
	"bufio"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	claudepkg "github.com/johnnyfreeman/viewscreen/claude"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/state"
)

// handleKeyMsg processes keyboard input and returns the model and any command.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	// When prompt editor is active, capture all keys for prompt editing
	if m.promptEditor.Active {
		return m.handlePromptEditorKeyMsg(msg)
	}

	// When search input is active, capture all keys for the search query
	if m.search.Active {
		return m.handleSearchKeyMsg(msg)
	}

	// During auto-exit countdown:
	// - q/ctrl+c: quit immediately
	// - space/enter: skip countdown and exit (continue the loop)
	// - any other key: cancel countdown and browse
	if m.autoExitRemaining > 0 {
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "space", "enter":
			return m, tea.Quit
		default:
			m.autoExitRemaining = 0
			return m, nil
		}
	}

	// Keys that work regardless of modal state
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelpModal = !m.showHelpModal
		return m, nil
	case "d":
		if m.layoutMode == LayoutHeader {
			m.showDetailsModal = !m.showDetailsModal
		}
		return m, nil
	case "esc":
		if m.showHelpModal {
			m.showHelpModal = false
		} else if m.showDetailsModal {
			m.showDetailsModal = false
		} else if m.search.HasQuery() {
			m.search.Clear()
		}
		return m, nil
	}

	// All remaining keys are navigation/action keys blocked by modals
	if m.showDetailsModal || m.showHelpModal {
		return m, nil
	}

	switch msg.String() {
	case "f":
		m.followMode = !m.followMode
		if m.followMode {
			m.viewport.GotoBottom()
		}
	case "/":
		m.search.Enter()
	case "n":
		if m.search.HasQuery() {
			m.search.NextMatch()
			m.scrollToSearchMatch()
		}
	case "N":
		if m.search.HasQuery() {
			m.search.PrevMatch()
			m.scrollToSearchMatch()
		}
	case "e":
		if m.stdinDone {
			m.promptEditor.Enter(m.state.Prompt)
		}
	case "up", "k":
		m.followMode = false
		m.viewport.ScrollUp(1)
	case "down", "j":
		m.viewport.ScrollDown(1)
	case "pgup":
		m.followMode = false
		m.viewport.HalfPageUp()
	case "pgdown":
		m.viewport.HalfPageDown()
	case "home", "g":
		m.followMode = false
		m.viewport.GotoTop()
	case "end", "G":
		m.followMode = true
		m.viewport.GotoBottom()
	}
	return m, nil
}

// handlePromptEditorKeyMsg processes keyboard input while prompt editing is active.
func (m Model) handlePromptEditorKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel editing, restore original prompt
		m.promptEditor.Cancel(m.state.Prompt)
	case "enter":
		// Confirm the edited prompt
		m.state.Prompt = m.promptEditor.Value
		m.prompt = m.promptEditor.Value
		m.promptEditor.Exit()
		if m.claudeProcess != nil {
			return m, func() tea.Msg { return RerunMsg{Prompt: m.state.Prompt} }
		}
	case "backspace":
		m.promptEditor.Backspace()
	case "delete":
		m.promptEditor.Delete()
	case "left":
		m.promptEditor.CursorLeft()
	case "right":
		m.promptEditor.CursorRight()
	case "home", "ctrl+a":
		m.promptEditor.CursorHome()
	case "end", "ctrl+e":
		m.promptEditor.CursorEnd()
	case "ctrl+c":
		return m, tea.Quit
	default:
		text := msg.String()
		if len(text) == 1 {
			m.promptEditor.TypeRune(rune(text[0]))
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

	// Reserve space for prompt editor when active
	if m.promptEditor.Active {
		contentHeight--
	}

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(contentHeight)

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
	// In subprocess mode, don't auto-exit — user may want to browse or re-run
	if m.claudeProcess != nil {
		return m, nil
	}
	if m.autoExit {
		m.autoExitRemaining = 5
		return m, AutoExitTick()
	}
	return m, nil
}

// handleRerun kills the old claude process, resets state, and spawns a new run.
func (m Model) handleRerun(msg RerunMsg) (Model, tea.Cmd) {
	// Kill old process
	if m.claudeProcess != nil {
		m.claudeProcess.Kill()
		m.claudeProcess.Wait()
	}

	// Reset state
	m.content = &strings.Builder{}
	m.stdinDone = false
	m.autoExitRemaining = 0
	st := state.NewState()
	st.Prompt = msg.Prompt
	m.state = st
	m.processor = events.NewEventProcessor(st)
	m.prompt = msg.Prompt
	m.followMode = true
	m.search.Clear()

	// Spawn new claude process
	proc, err := claudepkg.Start(msg.Prompt, nil)
	if err != nil {
		m.content.WriteString("Error starting claude: " + err.Error() + "\n")
		m.viewport.SetContent(m.content.String())
		m.stdinDone = true
		return m, nil
	}

	m.claudeProcess = proc
	m.inputReader = proc.Stdout()

	// Create new scanner
	m.scanner = bufio.NewScanner(m.inputReader)
	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	m.scanner.Buffer(buf, maxCapacity)

	// Clear viewport
	m.viewport.SetContent("")

	return m, ReadStdinLine(m.scanner)
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
